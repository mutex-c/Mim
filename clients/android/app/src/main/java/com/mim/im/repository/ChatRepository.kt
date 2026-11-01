package com.mim.im.repository

import android.content.Context
import android.net.Uri
import android.util.Log
import com.google.gson.Gson
import com.mim.im.api.ApiClient
import com.mim.im.data.local.ConversationDao
import com.mim.im.data.local.ConversationEntity
import com.mim.im.data.local.MessageDao
import com.mim.im.data.local.MessageEntity
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.data.model.*
import com.mim.im.network.WsManager
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.asRequestBody
import java.io.File
import java.io.FileOutputStream
import java.net.URLEncoder
import java.util.UUID

class ChatRepository(
    private val context: Context,
    private val prefs: PreferencesHelper,
    private val apiClient: ApiClient,
    private val wsManager: WsManager,
    private val messageDao: MessageDao,
    private val conversationDao: ConversationDao,
    val webRtcManager: com.mim.im.webrtc.WebRtcManager? = null
) {
    companion object {
        private const val TAG = "ChatRepository"
    }

    private val gson = Gson()
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    val currentOwnerId: String
        get() = prefs.currentUser?.id?.toString() ?: "0"

    fun resolveMediaUrl(url: String?): String {
        if (url.isNullOrBlank()) return ""
        if (url.startsWith("http://") || url.startsWith("https://")) return url
        val cleanBase = prefs.serverUrl.trimEnd('/')
        val cleanPath = if (url.startsWith("/")) url else "/$url"
        return "$cleanBase$cleanPath"
    }

    // --- WebRTC Audio & Video Call State ---
    private val _activeCall = MutableStateFlow<CallSessionState?>(null)
    val activeCall: StateFlow<CallSessionState?> = _activeCall.asStateFlow()

    private val _incomingCall = MutableStateFlow<CallSessionState?>(null)
    val incomingCall: StateFlow<CallSessionState?> = _incomingCall.asStateFlow()

    private val _isPeerCameraOff = MutableStateFlow(false)
    val isPeerCameraOff: StateFlow<Boolean> = _isPeerCameraOff.asStateFlow()

    private var pendingIncomingOffer: WebRTCSignalDto? = null
    private var callTimerJob: Job? = null

    init {
        observeWsEvents()
    }

    private fun observeWsEvents() {
        // 1. Observe incoming messages
        scope.launch {
            wsManager.incomingMessages.collect { wsMsg ->
                handleIncomingMessage(wsMsg)
            }
        }

        // 2. Observe ACKs
        scope.launch {
            wsManager.messageAcks.collect { ack ->
                handleMessageAck(ack)
            }
        }

        // 3. Observe Revokes
        scope.launch {
            wsManager.revokeEvents.collect { revokeMsg ->
                handleRevoke(revokeMsg)
            }
        }

        // 4. Observe WebRTC Signals
        scope.launch {
            wsManager.webRtcSignals.collect { signal ->
                handleWebRTCSignal(signal)
            }
        }

        // 5. Observe connection status to refresh conversations on connect/reconnect
        scope.launch {
            wsManager.status.collect { status ->
                if (status == com.mim.im.network.WsStatus.CONNECTED) {
                    refreshConversations()
                }
            }
        }
    }

    private suspend fun handleIncomingMessage(wsMsg: WsChatMessage) {
        val ownerId = currentOwnerId
        var fileName: String? = null
        var fileSize: Long = 0
        var mediaUrl: String? = null

        if (wsMsg.msgType == MsgType.FILE || wsMsg.msgType == MsgType.IMAGE || wsMsg.msgType == MsgType.VOICE) {
            val raw = wsMsg.content.trim()
            if (raw.startsWith("{")) {
                try {
                    val meta = gson.fromJson(raw, FileContentMeta::class.java)
                    fileName = meta.name
                    fileSize = meta.size
                    mediaUrl = meta.url
                } catch (_: Exception) {}
            }
            if (mediaUrl.isNullOrEmpty() && (raw.startsWith("/") || raw.startsWith("http"))) {
                mediaUrl = raw
                try {
                    val uri = Uri.parse(raw)
                    val qName = uri.getQueryParameter("name")
                    if (!qName.isNullOrBlank() && fileName.isNullOrBlank()) {
                        fileName = Uri.decode(qName)
                    }
                    val qSize = uri.getQueryParameter("size")
                    if (!qSize.isNullOrBlank() && fileSize == 0L) {
                        fileSize = parseFileSize(qSize)
                    }
                } catch (_: Exception) {}
            }
            if (fileName.isNullOrBlank()) {
                fileName = when (wsMsg.msgType) {
                    MsgType.IMAGE -> "image.jpg"
                    MsgType.VOICE -> "voice.mp4"
                    else -> "file"
                }
            }
        }

        val targetMsgId = if (wsMsg.msgId > 0) wsMsg.msgId.toString() else wsMsg.clientMsgId
        val existing = messageDao.getMessageById(ownerId, targetMsgId)
            ?: (if (wsMsg.clientMsgId.isNotEmpty()) messageDao.getMessageById(ownerId, wsMsg.clientMsgId) else null)

        val localFilePath = existing?.localFilePath
        val isDownloaded = existing?.isDownloaded ?: (!localFilePath.isNullOrEmpty() && File(localFilePath).exists())

        val msgEntity = MessageEntity(
            accountOwnerId = ownerId,
            msgId = targetMsgId,
            conversationId = wsMsg.conversationId,
            senderId = wsMsg.senderId.toString(),
            receiverId = wsMsg.receiverId.toString(),
            convType = if (wsMsg.convType == ConvType.GROUP) "group" else "single",
            contentType = when (wsMsg.msgType) {
                MsgType.IMAGE -> "image"
                MsgType.VOICE -> "voice"
                MsgType.FILE -> "file"
                MsgType.REVOKE -> "revoked"
                else -> "text"
            },
            content = wsMsg.content,
            quotedMsgId = if (wsMsg.quotedMsgId > 0) wsMsg.quotedMsgId.toString() else "",
            quotedContent = wsMsg.quotedContent,
            mediaUrl = mediaUrl,
            fileName = fileName ?: existing?.fileName,
            fileSize = if (fileSize > 0) fileSize else (existing?.fileSize ?: 0L),
            localFilePath = localFilePath,
            isDownloaded = isDownloaded,
            seq = wsMsg.seqId,
            status = "sent",
            timestamp = wsMsg.sendTime
        )
        if (wsMsg.clientMsgId.isNotEmpty() && wsMsg.msgId > 0 && wsMsg.clientMsgId != wsMsg.msgId.toString()) {
            messageDao.deleteMessage(ownerId, wsMsg.clientMsgId)
        }
        messageDao.insertOrUpdate(msgEntity)

        // Update conversation summary
        val existingConv = conversationDao.getConversationById(ownerId, wsMsg.conversationId)
        val preview = when (wsMsg.msgType) {
            MsgType.IMAGE -> "[图片]"
            MsgType.VOICE -> "[语音留言]"
            MsgType.FILE -> "[文件] ${fileName ?: ""}"
            MsgType.REVOKE -> "[对方撤回了一条消息]"
            else -> wsMsg.content
        }

        val updatedConv = ConversationEntity(
            accountOwnerId = ownerId,
            id = wsMsg.conversationId,
            convType = if (wsMsg.convType == ConvType.GROUP) "group" else "single",
            targetId = if (wsMsg.convType == ConvType.GROUP) wsMsg.receiverId.toString() else wsMsg.senderId.toString(),
            title = existingConv?.title ?: ("用户 " + wsMsg.senderId),
            avatar = existingConv?.avatar ?: "",
            lastMessage = preview,
            lastMessageTime = wsMsg.sendTime,
            unreadCount = (existingConv?.unreadCount ?: 0) + 1,
            isPinned = existingConv?.isPinned ?: false,
            isMuted = existingConv?.isMuted ?: false
        )
        conversationDao.insertOrUpdate(updatedConv)
    }

    private suspend fun handleMessageAck(ack: WsMessageAck) {
        val ownerId = currentOwnerId
        Log.d(TAG, "handleMessageAck: clientMsgId=${ack.clientMsgId}, code=${ack.code}, msgId=${ack.msgId}, seq=${ack.seqId}")
        if (ack.clientMsgId.isNotEmpty()) {
            if (ack.code == 0) {
                messageDao.updateStatus(ownerId, ack.clientMsgId, "sent")
            } else {
                messageDao.updateStatus(ownerId, ack.clientMsgId, "failed")
            }
        }
    }

    private suspend fun handleRevoke(revokeMsg: WsChatMessage) {
        val ownerId = currentOwnerId
        messageDao.markAsRevoked(ownerId, revokeMsg.msgId.toString())
    }

    private suspend fun handleWebRTCSignal(signal: WebRTCSignalDto) {
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L
        if (signal.toUserId != currentUserId && signal.toUserId != 0L) return

        when (signal.signalType.uppercase()) {
            "OFFER" -> {
                pendingIncomingOffer = signal
                val isVideo = signal.data.contains("\"callType\":\"video\"") || signal.data.contains("video")
                val callerName = try {
                    val friends = apiClient.service.getFriends().body()?.data
                    friends?.find { it.id == signal.fromUserId }?.displayName ?: "用户 ${signal.fromUserId}"
                } catch (_: Exception) {
                    "用户 ${signal.fromUserId}"
                }

                _incomingCall.value = CallSessionState(
                    callId = signal.callId,
                    peerId = signal.fromUserId,
                    peerName = callerName,
                    isVideo = isVideo,
                    isOutgoing = false,
                    isConnected = false,
                    durationSeconds = 0
                )
            }
            "ANSWER" -> {
                val current = _activeCall.value
                if (current != null && current.callId == signal.callId) {
                    _activeCall.value = current.copy(isConnected = true)
                    startCallTimer()
                    webRtcManager?.handleAnswer(signal)
                }
            }
            "CANDIDATE" -> {
                webRtcManager?.handleCandidate(signal)
            }
            "CAMERA_STATE" -> {
                try {
                    val json = org.json.JSONObject(signal.data)
                    val isMuted = json.optBoolean("isVideoMuted", false)
                    _isPeerCameraOff.value = isMuted
                } catch (_: Exception) {}
            }
            "REJECT", "HANGUP" -> {
                if (_incomingCall.value?.callId == signal.callId) {
                    _incomingCall.value = null
                }
                if (_activeCall.value?.callId == signal.callId) {
                    stopCallTimer()
                    _activeCall.value = null
                }
                _isPeerCameraOff.value = false
                webRtcManager?.endCall()
                pendingIncomingOffer = null
            }
        }
    }

    fun initiateCall(targetId: Long, targetName: String, isVideo: Boolean) {
        val callId = "call_" + System.currentTimeMillis()
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L

        _isPeerCameraOff.value = false
        _activeCall.value = CallSessionState(
            callId = callId,
            peerId = targetId,
            peerName = targetName,
            isVideo = isVideo,
            isOutgoing = true,
            isConnected = false,
            durationSeconds = 0
        )

        webRtcManager?.initiateCall(callId, targetId, isVideo) { sigType, toUid, cId, data ->
            val signal = WebRTCSignalDto(
                callId = cId,
                fromUserId = currentUserId,
                toUserId = toUid,
                signalType = sigType,
                data = data
            )
            wsManager.sendWebRTCSignal(signal)
        }
    }

    fun acceptIncomingCall() {
        val incoming = _incomingCall.value ?: return
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L

        _incomingCall.value = null
        _isPeerCameraOff.value = false
        _activeCall.value = incoming.copy(isConnected = true)
        startCallTimer()

        val offer = pendingIncomingOffer
        if (offer != null) {
            webRtcManager?.handleOffer(offer) { sigType, toUid, cId, data ->
                val signal = WebRTCSignalDto(
                    callId = cId,
                    fromUserId = currentUserId,
                    toUserId = toUid,
                    signalType = sigType,
                    data = data
                )
                wsManager.sendWebRTCSignal(signal)
            }
            pendingIncomingOffer = null
        } else {
            val signal = WebRTCSignalDto(
                callId = incoming.callId,
                fromUserId = currentUserId,
                toUserId = incoming.peerId,
                signalType = "ANSWER",
                data = "{\"callType\":\"${if (incoming.isVideo) "video" else "audio"}\"}"
            )
            wsManager.sendWebRTCSignal(signal)
        }
    }

    fun rejectIncomingCall() {
        val incoming = _incomingCall.value ?: return
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L

        _incomingCall.value = null
        _isPeerCameraOff.value = false
        pendingIncomingOffer = null
        webRtcManager?.endCall()

        val signal = WebRTCSignalDto(
            callId = incoming.callId,
            fromUserId = currentUserId,
            toUserId = incoming.peerId,
            signalType = "REJECT",
            data = ""
        )
        wsManager.sendWebRTCSignal(signal)
    }

    fun endActiveCall() {
        val active = _activeCall.value ?: return
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L

        stopCallTimer()
        _activeCall.value = null
        _isPeerCameraOff.value = false
        pendingIncomingOffer = null
        webRtcManager?.endCall()

        val signal = WebRTCSignalDto(
            callId = active.callId,
            fromUserId = currentUserId,
            toUserId = active.peerId,
            signalType = "HANGUP",
            data = ""
        )
        wsManager.sendWebRTCSignal(signal)
    }

    fun toggleCamera(cameraOff: Boolean) {
        webRtcManager?.toggleCamera(cameraOff)
        val active = _activeCall.value ?: return
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L
        val signal = WebRTCSignalDto(
            callId = active.callId,
            fromUserId = currentUserId,
            toUserId = active.peerId,
            signalType = "CAMERA_STATE",
            data = """{"isVideoMuted":$cameraOff}"""
        )
        wsManager.sendWebRTCSignal(signal)
    }

    private fun startCallTimer() {
        stopCallTimer()
        callTimerJob = scope.launch {
            while (isActive) {
                delay(1000L)
                val current = _activeCall.value ?: break
                if (current.isConnected) {
                    _activeCall.value = current.copy(durationSeconds = current.durationSeconds + 1)
                }
            }
        }
    }

    private fun stopCallTimer() {
        callTimerJob?.cancel()
        callTimerJob = null
    }

    suspend fun getCallHistory(): Result<List<CallHistoryItemDto>> {
        val currentUserId = currentOwnerId.toLongOrNull() ?: 0L
        return try {
            val resp = apiClient.service.getCallHistory(currentUserId)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(resp.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to fetch call history"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun getConversations(): Flow<List<ConversationEntity>> {
        return conversationDao.getAllConversations(currentOwnerId)
    }

    fun getMessages(conversationId: String): Flow<List<MessageEntity>> {
        return messageDao.getMessagesByConversation(currentOwnerId, conversationId)
    }

    suspend fun getMessageById(msgId: String): MessageEntity? {
        return messageDao.getMessageById(currentOwnerId, msgId)
    }

    private fun parseIsoTime(timeStr: String?): Long {
        if (timeStr.isNullOrBlank()) return System.currentTimeMillis()
        timeStr.toLongOrNull()?.let { return it }
        return try {
            java.time.Instant.parse(timeStr).toEpochMilli()
        } catch (_: Exception) {
            try {
                java.time.OffsetDateTime.parse(timeStr).toInstant().toEpochMilli()
            } catch (_: Exception) {
                try {
                    java.text.SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss", java.util.Locale.getDefault())
                        .parse(timeStr)?.time ?: System.currentTimeMillis()
                } catch (_: Exception) {
                    System.currentTimeMillis()
                }
            }
        }
    }

    private fun parseFileSize(sizeStr: String?): Long {
        if (sizeStr.isNullOrBlank()) return 0L
        sizeStr.toLongOrNull()?.let { return it }
        val clean = Uri.decode(sizeStr).trim().uppercase()
        return try {
            when {
                clean.endsWith("GB") -> {
                    val num = clean.removeSuffix("GB").trim().toDouble()
                    (num * 1024 * 1024 * 1024).toLong()
                }
                clean.endsWith("MB") -> {
                    val num = clean.removeSuffix("MB").trim().toDouble()
                    (num * 1024 * 1024).toLong()
                }
                clean.endsWith("KB") -> {
                    val num = clean.removeSuffix("KB").trim().toDouble()
                    (num * 1024).toLong()
                }
                clean.endsWith("B") -> {
                    val num = clean.removeSuffix("B").trim().toDouble()
                    num.toLong()
                }
                else -> 0L
            }
        } catch (_: Exception) {
            0L
        }
    }

    suspend fun refreshConversations() {
        val ownerId = currentOwnerId
        try {
            val resp = apiClient.service.getConversations()
            if (resp.isSuccessful && resp.body()?.code == 0) {
                val list = resp.body()?.data ?: emptyList()
                val entities = list.map { dto ->
                    ConversationEntity(
                        accountOwnerId = ownerId,
                        id = dto.conversationId,
                        convType = if (dto.convType == 1) "group" else "single",
                        targetId = dto.targetId.toString(),
                        title = dto.title ?: ("会话 " + dto.targetId),
                        avatar = dto.avatarText ?: "",
                        lastMessage = dto.lastMsgPreview,
                        lastMessageTime = parseIsoTime(dto.lastMsgTime),
                        unreadCount = dto.unreadCount,
                        isPinned = dto.isPinned == 1,
                        isMuted = dto.isMuted == 1
                    )
                }
                if (entities.isNotEmpty()) {
                    conversationDao.insertAll(entities)
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to refresh conversations", e)
        }
    }

    suspend fun syncRemoteMessages(conversationId: String) {
        val ownerId = currentOwnerId
        try {
            val localMaxSeq = messageDao.getMaxSeqByConversation(ownerId, conversationId) ?: 0L
            val resp = apiClient.service.getMessages(convId = conversationId, lastSeqId = localMaxSeq, limit = 50)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                val list = resp.body()?.data ?: emptyList()
                // Clean up any temporary clientMsgId rows if remote server has synced the finalized message
                list.forEach { wsMsg ->
                    if (wsMsg.clientMsgId.isNotEmpty() && wsMsg.msgId > 0 && wsMsg.clientMsgId != wsMsg.msgId.toString()) {
                        messageDao.deleteMessage(ownerId, wsMsg.clientMsgId)
                    }
                }
                val entities = list.map { wsMsg ->
                    var fileName: String? = null
                    var fileSize: Long = 0
                    var mediaUrl: String? = null

                    if (wsMsg.msgType == MsgType.FILE || wsMsg.msgType == MsgType.IMAGE || wsMsg.msgType == MsgType.VOICE) {
                        val raw = wsMsg.content.trim()
                        if (raw.startsWith("{")) {
                            try {
                                val meta = gson.fromJson(raw, FileContentMeta::class.java)
                                fileName = meta.name
                                fileSize = meta.size
                                mediaUrl = meta.url
                            } catch (_: Exception) {}
                        }
                        if (mediaUrl.isNullOrEmpty() && (raw.startsWith("/") || raw.startsWith("http"))) {
                            mediaUrl = raw
                            try {
                                val uri = Uri.parse(raw)
                                val qName = uri.getQueryParameter("name")
                                if (!qName.isNullOrBlank() && fileName.isNullOrBlank()) {
                                    fileName = Uri.decode(qName)
                                }
                                val qSize = uri.getQueryParameter("size")
                                if (!qSize.isNullOrBlank() && fileSize == 0L) {
                                    fileSize = parseFileSize(qSize)
                                }
                            } catch (_: Exception) {}
                        }
                        if (fileName.isNullOrBlank()) {
                            fileName = when (wsMsg.msgType) {
                                MsgType.IMAGE -> "image.jpg"
                                MsgType.VOICE -> "voice.mp4"
                                else -> "file"
                            }
                        }
                    }

                    val targetMsgId = if (wsMsg.msgId > 0) wsMsg.msgId.toString() else wsMsg.clientMsgId
                    val existing = messageDao.getMessageById(ownerId, targetMsgId)
                        ?: (if (wsMsg.clientMsgId.isNotEmpty()) messageDao.getMessageById(ownerId, wsMsg.clientMsgId) else null)

                    val localFilePath = existing?.localFilePath
                    val isDownloaded = existing?.isDownloaded ?: (!localFilePath.isNullOrEmpty() && File(localFilePath).exists())

                    MessageEntity(
                        accountOwnerId = ownerId,
                        msgId = targetMsgId,
                        conversationId = wsMsg.conversationId,
                        senderId = wsMsg.senderId.toString(),
                        receiverId = wsMsg.receiverId.toString(),
                        convType = if (wsMsg.convType == ConvType.GROUP) "group" else "single",
                        contentType = when (wsMsg.msgType) {
                            MsgType.IMAGE -> "image"
                            MsgType.VOICE -> "voice"
                            MsgType.FILE -> "file"
                            MsgType.REVOKE -> "revoked"
                            else -> "text"
                        },
                        content = wsMsg.content,
                        quotedMsgId = if (wsMsg.quotedMsgId > 0) wsMsg.quotedMsgId.toString() else "",
                        quotedContent = wsMsg.quotedContent,
                        mediaUrl = mediaUrl,
                        fileName = fileName ?: existing?.fileName,
                        fileSize = if (fileSize > 0) fileSize else (existing?.fileSize ?: 0L),
                        localFilePath = localFilePath,
                        isDownloaded = isDownloaded,
                        seq = wsMsg.seqId,
                        status = "sent",
                        timestamp = wsMsg.sendTime
                    )
                }
                if (entities.isNotEmpty()) {
                    messageDao.insertAll(entities)
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to sync remote messages for $conversationId", e)
        }
    }

    suspend fun sendTextMessage(
        conversationId: String,
        convType: Int,
        targetId: Long,
        text: String,
        quotedMsgId: Long = 0L,
        quotedContent: String = ""
    ): Boolean {
        val ownerId = currentOwnerId
        val clientMsgId = UUID.randomUUID().toString()
        val senderId = ownerId.toLongOrNull() ?: 0L

        val entity = MessageEntity(
            accountOwnerId = ownerId,
            msgId = clientMsgId,
            conversationId = conversationId,
            senderId = senderId.toString(),
            receiverId = targetId.toString(),
            convType = if (convType == ConvType.GROUP) "group" else "single",
            contentType = "text",
            content = text,
            quotedMsgId = if (quotedMsgId > 0) quotedMsgId.toString() else "",
            quotedContent = quotedContent,
            seq = 0,
            status = "sending",
            timestamp = System.currentTimeMillis()
        )
        messageDao.insertOrUpdate(entity)

        val wsMsg = WsChatMessage(
            msgId = 0,
            clientMsgId = clientMsgId,
            conversationId = conversationId,
            convType = convType,
            senderId = senderId,
            receiverId = targetId,
            msgType = MsgType.TEXT,
            content = text,
            quotedMsgId = quotedMsgId,
            quotedContent = quotedContent,
            sendTime = System.currentTimeMillis()
        )
        val sent = wsManager.sendMessage(wsMsg)
        if (!sent) {
            messageDao.updateStatus(ownerId, clientMsgId, "failed")
        }
        return sent
    }

    suspend fun sendFileMessage(
        conversationId: String,
        convType: Int,
        targetId: Long,
        file: File,
        mimeType: String = "application/octet-stream",
        quotedMsgId: Long = 0L,
        quotedContent: String = ""
    ): Boolean {
        val ownerId = currentOwnerId
        val clientMsgId = UUID.randomUUID().toString()
        val senderId = ownerId.toLongOrNull() ?: 0L

        // 1. Upload file via HTTP multipart
        val requestFile = file.asRequestBody(mimeType.toMediaTypeOrNull())
        val body = MultipartBody.Part.createFormData("file", file.name, requestFile)

        val resp = try {
            apiClient.service.uploadFile(body)
        } catch (e: Exception) {
            Log.e(TAG, "File upload failed", e)
            return false
        }

        if (!resp.isSuccessful || resp.body()?.code != 0) {
            Log.e(TAG, "File upload server error: ${resp.body()?.msg ?: resp.errorBody()?.string()}")
            return false
        }

        val meta = resp.body()?.data ?: return false
        val isImage = mimeType.startsWith("image/")
        val isVoice = mimeType.startsWith("audio/")

        val contentType = when {
            isImage -> "image"
            isVoice -> "voice"
            else -> "file"
        }

        val msgType = when {
            isImage -> MsgType.IMAGE
            isVoice -> MsgType.VOICE
            else -> MsgType.FILE
        }

        // Standardize file URL across platforms
        val rawUrl = meta.url.ifBlank { "/api/file/download?file_id=${meta.fileId}&name=${URLEncoder.encode(file.name, "UTF-8")}" }
        var finalUrl = if (!rawUrl.contains("name=")) {
            "$rawUrl${if (rawUrl.contains("?")) "&" else "?"}name=${URLEncoder.encode(file.name, "UTF-8")}"
        } else rawUrl
        if (!finalUrl.contains("size=")) {
            finalUrl = "$finalUrl${if (finalUrl.contains("?")) "&" else "?"}size=${file.length()}"
        }

        val entity = MessageEntity(
            accountOwnerId = ownerId,
            msgId = clientMsgId,
            conversationId = conversationId,
            senderId = senderId.toString(),
            receiverId = targetId.toString(),
            convType = if (convType == ConvType.GROUP) "group" else "single",
            contentType = contentType,
            content = finalUrl,
            quotedMsgId = if (quotedMsgId > 0) quotedMsgId.toString() else "",
            quotedContent = quotedContent,
            mediaUrl = finalUrl,
            fileName = file.name,
            fileSize = if (meta.size > 0) meta.size else file.length(),
            localFilePath = file.absolutePath,
            isDownloaded = true,
            seq = 0,
            status = "sending",
            timestamp = System.currentTimeMillis()
        )
        messageDao.insertOrUpdate(entity)

        val wsMsg = WsChatMessage(
            msgId = 0,
            clientMsgId = clientMsgId,
            conversationId = conversationId,
            convType = convType,
            senderId = senderId,
            receiverId = targetId,
            msgType = msgType,
            content = finalUrl,
            quotedMsgId = quotedMsgId,
            quotedContent = quotedContent,
            sendTime = System.currentTimeMillis()
        )
        val sent = wsManager.sendMessage(wsMsg)
        if (!sent) {
            messageDao.updateStatus(ownerId, clientMsgId, "failed")
        }
        return sent
    }

    suspend fun downloadFile(
        msg: MessageEntity,
        onProgress: (Float) -> Unit,
        onSuccess: (File) -> Unit,
        onError: (String) -> Unit
    ) {
        val url = msg.mediaUrl ?: run {
            onError("Media URL is missing")
            return
        }

        withContext(Dispatchers.IO) {
            try {
                val downloadUrl = if (url.startsWith("http://") || url.startsWith("https://")) {
                    url
                } else {
                    "${prefs.serverUrl.trimEnd('/')}${if (url.startsWith("/")) url else "/$url"}"
                }

                val request = Request.Builder().url(downloadUrl).build()
                val response = OkHttpClient().newCall(request).execute()
                if (!response.isSuccessful) {
                    withContext(Dispatchers.Main) { onError("Download failed: HTTP ${response.code}") }
                    return@withContext
                }

                val body = response.body ?: run {
                    withContext(Dispatchers.Main) { onError("Empty response body") }
                    return@withContext
                }

                val fileName = msg.fileName ?: "mim_${System.currentTimeMillis()}"
                val downloadsDir = context.getExternalFilesDir("Download") ?: context.cacheDir
                val targetFile = File(downloadsDir, fileName)

                body.byteStream().use { input ->
                    FileOutputStream(targetFile).use { output ->
                        val buffer = ByteArray(8192)
                        var bytesRead: Int
                        var totalBytes: Long = 0
                        val contentLength = body.contentLength()

                        while (input.read(buffer).also { bytesRead = it } != -1) {
                            output.write(buffer, 0, bytesRead)
                            totalBytes += bytesRead
                            if (contentLength > 0) {
                                val progress = totalBytes.toFloat() / contentLength
                                withContext(Dispatchers.Main) { onProgress(progress) }
                            }
                        }
                        output.flush()
                    }
                }

                messageDao.updateDownloadStatus(currentOwnerId, msg.msgId, true, targetFile.absolutePath)
                withContext(Dispatchers.Main) { onSuccess(targetFile) }
            } catch (e: Exception) {
                Log.e(TAG, "Download error", e)
                withContext(Dispatchers.Main) { onError(e.message ?: "Download failed") }
            }
        }
    }

    suspend fun revokeMessage(msg: MessageEntity): Boolean {
        return try {
            val req = RevokeRequest(
                msgId = msg.msgId.toLongOrNull() ?: 0L,
                clientMsgId = msg.msgId
            )
            val resp = apiClient.service.revokeMessage(req)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                messageDao.markAsRevoked(currentOwnerId, msg.msgId)
                true
            } else {
                false
            }
        } catch (e: Exception) {
            Log.e(TAG, "Revoke failed", e)
            false
        }
    }

    suspend fun markConversationRead(conversationId: String) {
        val ownerId = currentOwnerId
        conversationDao.clearUnread(ownerId, conversationId)
        try {
            apiClient.service.readConversation(ReadConversationRequest(conversationId))
        } catch (_: Exception) {}
    }

    // --- Conversation Actions ---

    suspend fun pinConversation(conversationId: String, isPinned: Boolean): Result<Unit> {
        return try {
            val resp = apiClient.service.pinConversation(PinConversationReq(conversationId, if (isPinned) 1 else 0))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                conversationDao.updatePinned(currentOwnerId, conversationId, isPinned)
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to pin conversation"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun muteConversation(conversationId: String, isMuted: Boolean): Result<Unit> {
        return try {
            val resp = apiClient.service.muteConversation(MuteConversationReq(conversationId, if (isMuted) 1 else 0))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                conversationDao.updateMuted(currentOwnerId, conversationId, isMuted)
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to mute conversation"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteConversation(conversationId: String): Result<Unit> {
        return try {
            val resp = apiClient.service.deleteConversation(ConvActionReq(conversationId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                conversationDao.delete(currentOwnerId, conversationId)
                messageDao.deleteByConversation(currentOwnerId, conversationId)
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to delete conversation"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun clearMessages(conversationId: String): Result<Unit> {
        return try {
            val resp = apiClient.service.clearMessages(ConvActionReq(conversationId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                messageDao.deleteByConversation(currentOwnerId, conversationId)
                val existing = conversationDao.getConversationById(currentOwnerId, conversationId)
                if (existing != null) {
                    conversationDao.insertOrUpdate(existing.copy(lastMessage = "", unreadCount = 0))
                }
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to clear messages"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    // --- Group Management API Calls ---

    suspend fun createGroup(title: String, memberIds: List<Long>): Result<GroupInfoDto> {
        return try {
            val resp = apiClient.service.createGroup(CreateGroupReq(title = title, memberIds = memberIds))
            if (resp.isSuccessful && resp.body()?.code == 0 && resp.body()?.data != null) {
                Result.success(resp.body()!!.data!!)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to create group"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getGroupInfo(groupId: Long): Result<GroupInfoDto> {
        return try {
            val resp = apiClient.service.getGroupInfo(groupId)
            if (resp.isSuccessful && resp.body()?.code == 0 && resp.body()?.data != null) {
                Result.success(resp.body()!!.data!!)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to get group info"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getGroupMembers(groupId: Long): Result<List<GroupMemberDto>> {
        return try {
            val resp = apiClient.service.getGroupMembers(groupId)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(resp.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to fetch members"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateGroup(
        groupId: Long,
        title: String = "",
        notice: String = "",
        avatar: String = "",
        isAllMuted: Int = 0
    ): Result<Unit> {
        return try {
            val req = UpdateGroupReq(
                groupId = groupId,
                title = title,
                notice = notice,
                avatar = avatar,
                isAllMuted = isAllMuted
            )
            val resp = apiClient.service.updateGroup(req)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to update group"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun inviteGroup(groupId: Long, memberIds: List<Long>): Result<Unit> {
        return try {
            val resp = apiClient.service.inviteGroup(InviteGroupReq(groupId = groupId, memberIds = memberIds))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to invite members"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun muteMember(groupId: Long, targetId: Long, durationSeconds: Int): Result<Unit> {
        return try {
            val resp = apiClient.service.muteMember(MuteMemberReq(groupId = groupId, targetId = targetId, duration = durationSeconds))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to mute member"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun kickMember(groupId: Long, targetId: Long): Result<Unit> {
        return try {
            val resp = apiClient.service.kickMember(KickMemberReq(groupId = groupId, targetId = targetId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to kick member"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun setGroupRole(groupId: Long, targetId: Long, role: Int): Result<Unit> {
        return try {
            val resp = apiClient.service.setGroupRole(SetGroupRoleReq(groupId = groupId, targetId = targetId, role = role))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to set member role"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun transferGroupOwner(groupId: Long, newOwnerId: Long): Result<Unit> {
        return try {
            val resp = apiClient.service.transferGroupOwner(TransferGroupOwnerReq(groupId = groupId, newOwnerId = newOwnerId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to transfer ownership"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun disbandGroup(groupId: Long): Result<Unit> {
        return try {
            val resp = apiClient.service.disbandGroup(GroupIdReq(groupId = groupId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to disband group"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun quitGroup(groupId: Long): Result<Unit> {
        return try {
            val resp = apiClient.service.quitGroup(GroupIdReq(groupId = groupId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to quit group"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getAISummary(conversationId: String, lang: String = "zh"): Result<AISummaryResp> {
        return try {
            val resp = apiClient.service.getAISummary(AISummaryReq(conversationId, lang))
            if (resp.isSuccessful && resp.body()?.code == 0 && resp.body()?.data != null) {
                Result.success(resp.body()!!.data!!)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "生成 AI 会话总结失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getAIReply(conversationId: String, count: Int = 3): Result<List<AIReplyItem>> {
        return try {
            val resp = apiClient.service.getAIReply(AIReplyReq(conversationId, count))
            if (resp.isSuccessful && resp.body()?.code == 0 && resp.body()?.data != null) {
                Result.success(resp.body()!!.data!!.suggestions)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "获取智能回复失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getAIConfig(): Result<AIConfigDto> {
        return try {
            val resp = apiClient.service.getAIConfig()
            if (resp.isSuccessful && resp.body()?.code == 0 && resp.body()?.data != null) {
                Result.success(resp.body()!!.data!!)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "获取 AI 配置失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun saveAIConfig(config: AIConfigDto): Result<Boolean> {
        return try {
            val resp = apiClient.service.saveAIConfig(config)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(true)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "保存 AI 配置失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun testAIConfig(config: AIConfigDto): Result<Long> {
        return try {
            val resp = apiClient.service.testAIConfig(
                AITestReq(
                    baseUrl = config.baseUrl,
                    apiKey = config.apiKey,
                    model = config.model
                )
            )
            if (resp.isSuccessful && resp.body()?.code == 0 && resp.body()?.data != null) {
                Result.success(resp.body()!!.data!!.latencyMs)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "测试连接失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateProfile(nickname: String, signature: String, avatar: String): Result<Boolean> {
        return try {
            val resp = apiClient.service.updateProfile(UpdateProfileReq(nickname, avatar, signature))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(true)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "更新个人资料失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun updateConfig(searchMode: Int, needConfirm: Boolean): Result<Boolean> {
        return try {
            val searchByAcc = if (searchMode >= 1) 1 else 0
            val searchByNick = if (searchMode >= 2) 1 else 0
            val verifyMode = if (needConfirm) 2 else 1
            val resp = apiClient.service.updateConfig(UpdateConfigReq(searchByAcc, searchByNick, verifyMode))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(true)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "更新配置失败"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}
