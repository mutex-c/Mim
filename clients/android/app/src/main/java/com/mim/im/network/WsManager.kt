package com.mim.im.network

import android.util.Log
import com.google.gson.Gson
import com.mim.im.api.ApiClient
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.data.model.*
import com.mim.im.protocol.CmdType
import com.mim.im.protocol.Frame
import com.mim.im.protocol.FrameFlag
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.StateFlow
import okhttp3.Request
import okhttp3.Response
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okio.ByteString
import okio.ByteString.Companion.toByteString

enum class WsStatus {
    DISCONNECTED,
    CONNECTING,
    CONNECTED
}

class WsManager(
    private val prefs: PreferencesHelper,
    private val apiClient: ApiClient
) {
    companion object {
        private const val TAG = "WsManager"
        private const val HEARTBEAT_INTERVAL_MS = 25_000L
    }

    private val gson = Gson()
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    private var webSocket: WebSocket? = null
    private var heartbeatJob: Job? = null
    private var reconnectJob: Job? = null
    private var reconnectAttempts = 0

    private val _status = MutableStateFlow(WsStatus.DISCONNECTED)
    val status: StateFlow<WsStatus> = _status

    private val _incomingMessages = MutableSharedFlow<WsChatMessage>(extraBufferCapacity = 64)
    val incomingMessages: SharedFlow<WsChatMessage> = _incomingMessages

    private val _messageAcks = MutableSharedFlow<WsMessageAck>(extraBufferCapacity = 64)
    val messageAcks: SharedFlow<WsMessageAck> = _messageAcks

    private val _syncResponses = MutableSharedFlow<WsSyncResponse>(extraBufferCapacity = 16)
    val syncResponses: SharedFlow<WsSyncResponse> = _syncResponses

    private val _revokeEvents = MutableSharedFlow<WsChatMessage>(extraBufferCapacity = 16)
    val revokeEvents: SharedFlow<WsChatMessage> = _revokeEvents

    private val _webRtcSignals = MutableSharedFlow<WebRTCSignalDto>(extraBufferCapacity = 32)
    val webRtcSignals: SharedFlow<WebRTCSignalDto> = _webRtcSignals

    @Synchronized
    fun connect() {
        val token = prefs.token ?: return
        if (_status.value == WsStatus.CONNECTED || _status.value == WsStatus.CONNECTING) {
            return
        }

        _status.value = WsStatus.CONNECTING
        val wsUrl = "${prefs.wsUrl}?token=$token&platform=android&device_id=android_${android.os.Build.MODEL.replace(" ", "_")}"
        Log.d(TAG, "Connecting to WebSocket: $wsUrl")

        val request = Request.Builder()
            .url(wsUrl)
            .build()

        webSocket = apiClient.okHttpClient.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                Log.i(TAG, "WebSocket connected successfully")
                _status.value = WsStatus.CONNECTED
                reconnectAttempts = 0
                startHeartbeat()
            }

            override fun onMessage(webSocket: WebSocket, bytes: ByteString) {
                handleBinaryFrame(bytes.toByteArray())
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                Log.d(TAG, "Received text frame: $text")
            }

            override fun onClosing(webSocket: WebSocket, code: Int, reason: String) {
                Log.w(TAG, "WebSocket closing: $code / $reason")
                _status.value = WsStatus.DISCONNECTED
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                Log.w(TAG, "WebSocket closed: $code / $reason")
                _status.value = WsStatus.DISCONNECTED
                stopHeartbeat()
                scheduleReconnect()
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                Log.e(TAG, "WebSocket failure: ${t.message}", t)
                _status.value = WsStatus.DISCONNECTED
                stopHeartbeat()
                scheduleReconnect()
            }
        })
    }

    private fun handleBinaryFrame(data: ByteArray) {
        val frame = Frame.decode(data) ?: run {
            Log.e(TAG, "Failed to decode binary frame of ${data.size} bytes")
            return
        }

        when (frame.cmdType) {
            CmdType.HeartbeatPong -> {
                Log.d(TAG, "Heartbeat Pong received, Seq: ${frame.seqId}")
            }
            CmdType.KickOut -> {
                Log.w(TAG, "Account kicked out: logged in from another device")
                scope.launch {
                    disconnect()
                }
            }
            CmdType.ChatMessagePush, CmdType.GroupMessagePush -> {
                try {
                    val json = String(frame.payload, Charsets.UTF_8)
                    val msg = gson.fromJson(json, WsChatMessage::class.java)
                    Log.d(TAG, "Incoming chat message: ${msg.clientMsgId}, content: ${msg.content}")
                    scope.launch {
                        _incomingMessages.emit(msg)
                    }
                    // Automatically send delivery ACK back to server
                    sendDeliveryAck(msg.clientMsgId, msg.msgId, msg.conversationId, msg.seqId)
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to parse ChatMessagePush payload", e)
                }
            }
            CmdType.ChatMessageAck -> {
                try {
                    val json = String(frame.payload, Charsets.UTF_8)
                    val ack = gson.fromJson(json, WsMessageAck::class.java)
                    Log.d(TAG, "Received Message Ack: clientMsgId=${ack.clientMsgId}, seq=${ack.seqId}, code=${ack.code}, msgId=${ack.msgId}")
                    scope.launch {
                        _messageAcks.emit(ack)
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to parse ChatMessageAck payload", e)
                }
            }
            CmdType.RevokeNotify -> {
                try {
                    val json = String(frame.payload, Charsets.UTF_8)
                    val msg = gson.fromJson(json, WsChatMessage::class.java)
                    Log.d(TAG, "Received Revoke Push: ${msg.msgId}")
                    scope.launch {
                        _revokeEvents.emit(msg)
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to parse RevokePush payload", e)
                }
            }
            CmdType.SyncResp, CmdType.BatchSyncResp -> {
                try {
                    val json = String(frame.payload, Charsets.UTF_8)
                    val syncResp = gson.fromJson(json, WsSyncResponse::class.java)
                    Log.d(TAG, "Received Sync Response, msgs: ${syncResp.messages.size}")
                    scope.launch {
                        _syncResponses.emit(syncResp)
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to parse SyncResp payload", e)
                }
            }
            CmdType.WebRTCSignal -> {
                try {
                    val json = String(frame.payload, Charsets.UTF_8)
                    val signal = gson.fromJson(json, WebRTCSignalDto::class.java)
                    Log.d(TAG, "Received WebRTC Signal: ${signal.signalType}, callId: ${signal.callId}")
                    scope.launch {
                        _webRtcSignals.emit(signal)
                    }
                } catch (e: Exception) {
                    Log.e(TAG, "Failed to parse WebRTCSignal payload", e)
                }
            }
            else -> {
                Log.d(TAG, "Unhandled CmdType: 0x${frame.cmdType.toString(16)}")
            }
        }
    }

    fun sendMessage(msg: WsChatMessage): Boolean {
        val ws = webSocket ?: return false
        val json = gson.toJson(msg)
        val payload = json.toByteArray(Charsets.UTF_8)
        val frame = Frame(
            cmdType = CmdType.ChatMessageSend,
            seqId = msg.seqId,
            flag = FrameFlag.JSON,
            payload = payload
        )
        return try {
            val bytes = frame.encode()
            ws.send(bytes.toByteString())
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send message frame", e)
            false
        }
    }

    fun sendSyncRequest(conversationId: String, lastSeqId: Long, limit: Int = 50): Boolean {
        val ws = webSocket ?: return false
        val req = WsSyncRequest(conversationId, lastSeqId, limit)
        val json = gson.toJson(req)
        val frame = Frame(
            cmdType = CmdType.SyncReq,
            seqId = lastSeqId,
            flag = FrameFlag.JSON,
            payload = json.toByteArray(Charsets.UTF_8)
        )
        return try {
            ws.send(frame.encode().toByteString())
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send sync frame", e)
            false
        }
    }

    fun sendWebRTCSignal(signal: WebRTCSignalDto): Boolean {
        val ws = webSocket ?: return false
        val json = gson.toJson(signal)
        val payload = json.toByteArray(Charsets.UTF_8)
        val frame = Frame(
            cmdType = CmdType.WebRTCSignal,
            seqId = 0,
            flag = FrameFlag.JSON,
            payload = payload
        )
        return try {
            ws.send(frame.encode().toByteString())
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send WebRTC signal frame", e)
            false
        }
    }

    fun sendDeliveryAck(clientMsgId: String, msgId: Long, convId: String, seqId: Long): Boolean {
        val ws = webSocket ?: return false
        val payloadMap = mapOf(
            "client_msg_id" to clientMsgId,
            "msg_id" to msgId,
            "conversation_id" to convId,
            "seq_id" to seqId,
            "delivery_time" to System.currentTimeMillis()
        )
        val json = gson.toJson(payloadMap)
        val frame = Frame(
            cmdType = CmdType.DeliveryAck,
            seqId = seqId,
            flag = FrameFlag.JSON,
            payload = json.toByteArray(Charsets.UTF_8)
        )
        return try {
            ws.send(frame.encode().toByteString())
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send delivery ack frame", e)
            false
        }
    }

    fun sendReadReceipt(convId: String, readSeq: Long): Boolean {
        val ws = webSocket ?: return false
        val payloadMap = mapOf(
            "conversation_id" to convId,
            "read_seq_id" to readSeq,
            "read_time" to System.currentTimeMillis()
        )
        val json = gson.toJson(payloadMap)
        val frame = Frame(
            cmdType = CmdType.ReadReceipt,
            seqId = readSeq,
            flag = FrameFlag.JSON,
            payload = json.toByteArray(Charsets.UTF_8)
        )
        return try {
            ws.send(frame.encode().toByteString())
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send read receipt frame", e)
            false
        }
    }

    private fun startHeartbeat() {
        stopHeartbeat()
        heartbeatJob = scope.launch {
            while (isActive) {
                delay(HEARTBEAT_INTERVAL_MS)
                if (_status.value == WsStatus.CONNECTED) {
                    val pingPayload = "{\"timestamp\":${System.currentTimeMillis()}}".toByteArray(Charsets.UTF_8)
                    val frame = Frame(
                        cmdType = CmdType.HeartbeatPing,
                        seqId = 0,
                        flag = FrameFlag.JSON,
                        payload = pingPayload
                    )
                    webSocket?.send(frame.encode().toByteString())
                }
            }
        }
    }

    private fun stopHeartbeat() {
        heartbeatJob?.cancel()
        heartbeatJob = null
    }

    private fun scheduleReconnect() {
        if (prefs.token.isNullOrEmpty()) return
        reconnectJob?.cancel()
        reconnectJob = scope.launch {
            reconnectAttempts++
            val delayMs = (2000L * (1 shl (reconnectAttempts - 1).coerceAtMost(4))).coerceAtMost(30_000L)
            Log.d(TAG, "Scheduling reconnect #$reconnectAttempts in ${delayMs}ms")
            delay(delayMs)
            connect()
        }
    }

    @Synchronized
    fun disconnect() {
        reconnectJob?.cancel()
        stopHeartbeat()
        webSocket?.close(1000, "Normal closure")
        webSocket = null
        _status.value = WsStatus.DISCONNECTED
    }
}
