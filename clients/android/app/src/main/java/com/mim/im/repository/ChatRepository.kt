package com.mim.im.repository

import android.content.Context
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
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.asRequestBody
import java.io.File
import java.io.FileOutputStream
import java.util.UUID

class ChatRepository(
    private val context: Context,
    private val prefs: PreferencesHelper,
    private val apiClient: ApiClient,
    private val wsManager: WsManager,
    private val messageDao: MessageDao,
    private val conversationDao: ConversationDao
) {
    companion object {
        private const val TAG = "ChatRepository"
    }

    private val gson = Gson()
    private val scope = CoroutineScope(Dispatchers.IO + SupervisorJob())

    val currentOwnerId: String
        get() = prefs.currentUser?.id?.toString() ?: "0"

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
    }

    private suspend fun handleIncomingMessage(wsMsg: WsChatMessage) {
        val ownerId = currentOwnerId
        var fileName: String? = null
        var fileSize: Long = 0
        var mediaUrl: String? = null

        if (wsMsg.msgType == MsgType.FILE || wsMsg.msgType == MsgType.IMAGE) {
            try {
                val meta = gson.fromJson(wsMsg.content, FileContentMeta::class.java)
                fileName = meta.name
                fileSize = meta.size
                mediaUrl = meta.url
            } catch (_: Exception) {}
        }

        val msgEntity = MessageEntity(
            accountOwnerId = ownerId,
            msgId = if (wsMsg.msgId > 0) wsMsg.msgId.toString() else wsMsg.clientMsgId,
            conversationId = wsMsg.conversationId,
            senderId = wsMsg.senderId.toString(),
            receiverId = wsMsg.receiverId.toString(),
            convType = if (wsMsg.convType == ConvType.GROUP) "group" else "single",
            contentType = when (wsMsg.msgType) {
                MsgType.IMAGE -> "image"
                MsgType.FILE -> "file"
                MsgType.REVOKE -> "revoked"
                else -> "text"
            },
            content = wsMsg.content,
            mediaUrl = mediaUrl,
            fileName = fileName,
            fileSize = fileSize,
            seq = wsMsg.seqId,
            status = "sent",
            timestamp = wsMsg.sendTime
        )
        messageDao.insertOrUpdate(msgEntity)

        // Update conversation summary
        val existingConv = conversationDao.getConversationById(ownerId, wsMsg.conversationId)
        val preview = when (wsMsg.msgType) {
            MsgType.IMAGE -> "[图片]"
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
            unreadCount = (existingConv?.unreadCount ?: 0) + 1
        )
        conversationDao.insertOrUpdate(updatedConv)
    }

    private suspend fun handleMessageAck(ack: WsMessageAck) {
        val ownerId = currentOwnerId
        if (ack.code == 0) {
            messageDao.updateStatus(ownerId, ack.clientMsgId, "sent")
        } else {
            messageDao.updateStatus(ownerId, ack.clientMsgId, "failed")
        }
    }

    private suspend fun handleRevoke(revokeMsg: WsChatMessage) {
        val ownerId = currentOwnerId
        messageDao.markAsRevoked(ownerId, revokeMsg.msgId.toString())
    }

    fun getConversations(): Flow<List<ConversationEntity>> {
        return conversationDao.getAllConversations(currentOwnerId)
    }

    fun getMessages(conversationId: String): Flow<List<MessageEntity>> {
        return messageDao.getMessagesByConversation(currentOwnerId, conversationId)
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
                        lastMessageTime = System.currentTimeMillis(),
                        unreadCount = dto.unreadCount
                    )
                }
                for (conv in entities) {
                    conversationDao.insertOrUpdate(conv)
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to refresh conversations", e)
        }
    }

    suspend fun sendTextMessage(
        conversationId: String,
        convType: Int,
        targetId: Long,
        text: String
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
        mimeType: String = "application/octet-stream"
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
            Log.e(TAG, "File upload server error: ${resp.body()?.msg}")
            return false
        }

        val meta = resp.body()?.data ?: return false
        val contentJson = gson.toJson(meta)
        val isImage = mimeType.startsWith("image/")

        val entity = MessageEntity(
            accountOwnerId = ownerId,
            msgId = clientMsgId,
            conversationId = conversationId,
            senderId = senderId.toString(),
            receiverId = targetId.toString(),
            convType = if (convType == ConvType.GROUP) "group" else "single",
            contentType = if (isImage) "image" else "file",
            content = contentJson,
            mediaUrl = meta.url,
            fileName = meta.name,
            fileSize = meta.size,
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
            msgType = if (isImage) MsgType.IMAGE else MsgType.FILE,
            content = contentJson,
            sendTime = System.currentTimeMillis()
        )
        return wsManager.sendMessage(wsMsg)
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
                // Ensure full URL if relative
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

                // Update database
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
}
