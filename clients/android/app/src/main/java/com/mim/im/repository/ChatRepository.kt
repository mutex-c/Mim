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

        if (wsMsg.msgType == MsgType.FILE || wsMsg.msgType == MsgType.IMAGE || wsMsg.msgType == MsgType.VOICE) {
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
                MsgType.VOICE -> "voice"
                MsgType.FILE -> "file"
                MsgType.REVOKE -> "revoked"
                else -> "text"
            },
            content = wsMsg.content,
            quotedMsgId = if (wsMsg.quotedMsgId > 0) wsMsg.quotedMsgId.toString() else "",
            quotedContent = wsMsg.quotedContent,
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

    suspend fun getMessageById(msgId: String): MessageEntity? {
        return messageDao.getMessageById(currentOwnerId, msgId)
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
                        unreadCount = dto.unreadCount,
                        isPinned = dto.isPinned == 1,
                        isMuted = dto.isMuted == 1
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
            Log.e(TAG, "File upload server error: ${resp.body()?.msg}")
            return false
        }

        val meta = resp.body()?.data ?: return false
        val contentJson = gson.toJson(meta)
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

        val entity = MessageEntity(
            accountOwnerId = ownerId,
            msgId = clientMsgId,
            conversationId = conversationId,
            senderId = senderId.toString(),
            receiverId = targetId.toString(),
            convType = if (convType == ConvType.GROUP) "group" else "single",
            contentType = contentType,
            content = contentJson,
            quotedMsgId = if (quotedMsgId > 0) quotedMsgId.toString() else "",
            quotedContent = quotedContent,
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
            msgType = msgType,
            content = contentJson,
            quotedMsgId = quotedMsgId,
            quotedContent = quotedContent,
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
}
