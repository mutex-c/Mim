package com.mim.im.data.local

import androidx.room.Entity
import androidx.room.Index
import androidx.room.PrimaryKey

@Entity(
    tableName = "messages",
    primaryKeys = ["accountOwnerId", "msgId"],
    indices = [
        Index(value = ["accountOwnerId", "conversationId"]),
        Index(value = ["accountOwnerId", "seq"])
    ]
)
data class MessageEntity(
    val accountOwnerId: String,
    val msgId: String,
    val conversationId: String,
    val senderId: String,
    val senderName: String = "",
    val senderAvatar: String = "",
    val receiverId: String = "",
    val groupId: String = "",
    val convType: String = "single",
    val contentType: String = "text",
    val content: String = "",
    val quotedMsgId: String = "",
    val quotedContent: String = "",
    val mediaUrl: String? = null,
    val fileName: String? = null,
    val fileSize: Long = 0,
    val localFilePath: String? = null,
    val isDownloaded: Boolean = false,
    val seq: Long = 0,
    val status: String = "sent", // sending, sent, failed, revoked
    val timestamp: Long = System.currentTimeMillis()
)

@Entity(
    tableName = "conversations",
    primaryKeys = ["accountOwnerId", "id"],
    indices = [
        Index(value = ["accountOwnerId", "lastMessageTime"])
    ]
)
data class ConversationEntity(
    val accountOwnerId: String,
    val id: String, // e.g. "single_user123" or "group_grp456"
    val convType: String = "single",
    val targetId: String = "",
    val title: String = "",
    val avatar: String = "",
    val lastMessage: String = "",
    val lastMessageTime: Long = System.currentTimeMillis(),
    val unreadCount: Int = 0,
    val isPinned: Boolean = false,
    val isMuted: Boolean = false
)
