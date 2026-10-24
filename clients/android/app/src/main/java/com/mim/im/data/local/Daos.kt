package com.mim.im.data.local

import androidx.room.*
import kotlinx.coroutines.flow.Flow

@Dao
interface MessageDao {
    @Query("SELECT * FROM messages WHERE accountOwnerId = :accountOwnerId AND conversationId = :conversationId ORDER BY timestamp ASC")
    fun getMessagesByConversation(accountOwnerId: String, conversationId: String): Flow<List<MessageEntity>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertOrUpdate(message: MessageEntity)

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(messages: List<MessageEntity>)

    @Query("UPDATE messages SET status = :status WHERE accountOwnerId = :accountOwnerId AND msgId = :msgId")
    suspend fun updateStatus(accountOwnerId: String, msgId: String, status: String)

    @Query("UPDATE messages SET isDownloaded = :isDownloaded, localFilePath = :localPath WHERE accountOwnerId = :accountOwnerId AND msgId = :msgId")
    suspend fun updateDownloadStatus(accountOwnerId: String, msgId: String, isDownloaded: Boolean, localPath: String)

    @Query("UPDATE messages SET contentType = 'revoked', content = '对方撤回了一条消息' WHERE accountOwnerId = :accountOwnerId AND msgId = :msgId")
    suspend fun markAsRevoked(accountOwnerId: String, msgId: String)

    @Query("SELECT MAX(seq) FROM messages WHERE accountOwnerId = :accountOwnerId")
    suspend fun getMaxSeq(accountOwnerId: String): Long?

    @Query("SELECT * FROM messages WHERE accountOwnerId = :accountOwnerId AND msgId = :msgId LIMIT 1")
    suspend fun getMessageById(accountOwnerId: String, msgId: String): MessageEntity?

    @Query("DELETE FROM messages WHERE accountOwnerId = :accountOwnerId AND conversationId = :conversationId")
    suspend fun deleteByConversation(accountOwnerId: String, conversationId: String)
}

@Dao
interface ConversationDao {
    @Query("SELECT * FROM conversations WHERE accountOwnerId = :accountOwnerId ORDER BY isPinned DESC, lastMessageTime DESC")
    fun getAllConversations(accountOwnerId: String): Flow<List<ConversationEntity>>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertOrUpdate(conversation: ConversationEntity)

    @Query("UPDATE conversations SET unreadCount = 0 WHERE accountOwnerId = :accountOwnerId AND id = :id")
    suspend fun clearUnread(accountOwnerId: String, id: String)

    @Query("UPDATE conversations SET isPinned = :isPinned WHERE accountOwnerId = :accountOwnerId AND id = :id")
    suspend fun updatePinned(accountOwnerId: String, id: String, isPinned: Boolean)

    @Query("UPDATE conversations SET isMuted = :isMuted WHERE accountOwnerId = :accountOwnerId AND id = :id")
    suspend fun updateMuted(accountOwnerId: String, id: String, isMuted: Boolean)

    @Query("DELETE FROM conversations WHERE accountOwnerId = :accountOwnerId AND id = :id")
    suspend fun delete(accountOwnerId: String, id: String)

    @Query("SELECT * FROM conversations WHERE accountOwnerId = :accountOwnerId AND id = :id LIMIT 1")
    suspend fun getConversationById(accountOwnerId: String, id: String): ConversationEntity?
}
