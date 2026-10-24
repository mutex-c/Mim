package com.mim.im.data.model

import com.google.gson.annotations.SerializedName

// --- Auth Models ---

data class LoginRequest(
    val account: String,
    val password: String,
    val platform: String = "android",
    @SerializedName("device_id") val deviceId: String = "android-device-01"
)

data class RegisterRequest(
    val account: String,
    val password: String,
    val nickname: String,
    val avatar: String = ""
)

data class ApiResponse<T>(
    val code: Int = 0,
    val msg: String = "",
    val data: T? = null
)

data class LoginResponseData(
    val user: UserDto,
    val token: TokenPairDto
)

data class TokenPairDto(
    @SerializedName("access_token") val accessToken: String = "",
    @SerializedName("refresh_token") val refreshToken: String = "",
    @SerializedName("expires_in") val expiresIn: Long = 0
)

data class UserDto(
    val id: Long = 0,
    val account: String = "",
    val nickname: String = "",
    val avatar: String = "",
    val signature: String = "",
    val role: Int = 1
)

data class UserInfo(
    val id: String = "",
    val username: String = "",
    val nickname: String = "",
    val avatar: String = ""
)

// --- Protocol Messages ---

object MsgType {
    const val TEXT = 0
    const val MARKDOWN = 1
    const val IMAGE = 2
    const val VOICE = 3
    const val VIDEO = 4
    const val FILE = 5
    const val STICKER = 6
    const val REVOKE = 7
    const val WEBRTC = 8
    const val SYSTEM = 9
}

object ConvType {
    const val SINGLE = 0
    const val GROUP = 1
}

data class WsChatMessage(
    @SerializedName("msg_id") val msgId: Long = 0,
    @SerializedName("client_msg_id") val clientMsgId: String = "",
    @SerializedName("conversation_id") val conversationId: String = "",
    @SerializedName("conv_type") val convType: Int = ConvType.SINGLE,
    @SerializedName("sender_id") val senderId: Long = 0,
    @SerializedName("receiver_id") val receiverId: Long = 0,
    @SerializedName("msg_type") val msgType: Int = MsgType.TEXT,
    val content: String = "",
    @SerializedName("quoted_msg_id") val quotedMsgId: Long = 0,
    @SerializedName("quoted_content") val quotedContent: String = "",
    @SerializedName("seq_id") val seqId: Long = 0,
    @SerializedName("send_time") val sendTime: Long = System.currentTimeMillis(),
    @SerializedName("is_encrypted") val isEncrypted: Boolean = false
)

data class WsMessageAck(
    @SerializedName("client_msg_id") val clientMsgId: String = "",
    @SerializedName("msg_id") val msgId: Long = 0,
    @SerializedName("seq_id") val seqId: Long = 0,
    val code: Int = 0,
    @SerializedName("error_msg") val errorMsg: String = "",
    @SerializedName("server_time") val serverTime: Long = 0
)

data class WsSyncRequest(
    @SerializedName("conversation_id") val conversationId: String = "",
    @SerializedName("last_seq_id") val lastSeqId: Long = 0,
    val limit: Int = 50
)

data class WsSyncResponse(
    @SerializedName("conversation_id") val conversationId: String = "",
    val messages: List<WsChatMessage> = emptyList(),
    @SerializedName("has_more") val hasMore: Boolean = false,
    @SerializedName("latest_seq_id") val latestSeqId: Long = 0
)

// File attachment metadata inside message content or for upload
data class FileContentMeta(
    val url: String = "",
    val name: String = "",
    val size: Long = 0,
    val mime: String = ""
)

// --- Conversation & Group HTTP Models ---

data class ConversationDto(
    val id: Long = 0,
    @SerializedName("conversation_id") val conversationId: String = "",
    @SerializedName("user_id") val userId: Long = 0,
    @SerializedName("target_id") val targetId: Long = 0,
    @SerializedName("conv_type") val convType: Int = 0,
    @SerializedName("last_msg_id") val lastMsgId: Long = 0,
    @SerializedName("last_msg_seq") val lastMsgSeq: Long = 0,
    @SerializedName("last_msg_preview") val lastMsgPreview: String = "",
    @SerializedName("last_msg_time") val lastMsgTime: String = "",
    @SerializedName("unread_count") val unreadCount: Int = 0,
    @SerializedName("is_pinned") val isPinned: Int = 0,
    @SerializedName("is_muted") val isMuted: Int = 0,
    val title: String? = null,
    @SerializedName("avatar_text") val avatarText: String? = null,
    @SerializedName("is_online") val isOnline: Boolean = false
)

data class GroupCreateRequest(
    val name: String,
    val avatar: String = "",
    val description: String = ""
)

data class GroupInfoDto(
    val id: Long = 0,
    val title: String = "",
    val avatar: String = "",
    val notice: String = "",
    @SerializedName("owner_id") val ownerId: Long = 0,
    @SerializedName("member_count") val memberCount: Int = 1,
    @SerializedName("is_all_muted") val isAllMuted: Int = 0
)

data class GroupMemberDto(
    @SerializedName("user_id") val userId: Long = 0,
    val account: String = "",
    val nickname: String = "",
    val avatar: String = "",
    val role: Int = 3, // 1: Owner, 2: Admin, 3: Member
    @SerializedName("is_muted") val isMuted: Boolean = false
)

data class RevokeRequest(
    @SerializedName("msg_id") val msgId: Long = 0,
    @SerializedName("client_msg_id") val clientMsgId: String = ""
)

data class ReadConversationRequest(
    @SerializedName("conversation_id") val conversationId: String = ""
)


// --- Friend Request & Relationship DTOs ---

data class FriendRequestDto(
    val id: Long = 0,
    @SerializedName("from_user_id") val fromUserId: Long = 0,
    @SerializedName("to_user_id") val toUserId: Long = 0,
    @SerializedName("from_nickname") val fromNickname: String = "",
    @SerializedName("from_avatar") val fromAvatar: String = "",
    @SerializedName("from_account") val fromAccount: String = "",
    val message: String = "",
    val status: Int = 1,
    @SerializedName("created_at") val createdAt: String = ""
)

data class HandleFriendRequestReq(
    @SerializedName("request_id") val requestId: Long,
    val action: Int
)

data class AddFriendReq(
    @SerializedName("target_id") val targetId: Long,
    val msg: String = ""
)

data class DeleteFriendReq(
    @SerializedName("friend_id") val friendId: Long
)

// --- Group Management DTOs ---

data class CreateGroupReq(
    val title: String,
    val avatar: String = "",
    val notice: String = "",
    @SerializedName("member_ids") val memberIds: List<Long> = emptyList()
)

data class UpdateGroupReq(
    @SerializedName("group_id") val groupId: Long,
    val title: String = "",
    val notice: String = "",
    val avatar: String = "",
    @SerializedName("is_all_muted") val isAllMuted: Int = 0
)

data class InviteGroupReq(
    @SerializedName("group_id") val groupId: Long,
    @SerializedName("member_ids") val memberIds: List<Long>
)

data class MuteMemberReq(
    @SerializedName("group_id") val groupId: Long,
    @SerializedName("target_id") val targetId: Long,
    val duration: Int
)

data class KickMemberReq(
    @SerializedName("group_id") val groupId: Long,
    @SerializedName("target_id") val targetId: Long
)

data class GroupIdReq(
    @SerializedName("group_id") val groupId: Long
)

data class SetGroupRoleReq(
    @SerializedName("group_id") val groupId: Long,
    @SerializedName("target_id") val targetId: Long,
    val role: Int
)

data class TransferGroupOwnerReq(
    @SerializedName("group_id") val groupId: Long,
    @SerializedName("new_owner_id") val newOwnerId: Long
)

// --- Conversation Management DTOs ---

data class PinConversationReq(
    @SerializedName("conversation_id") val conversationId: String,
    @SerializedName("is_pinned") val isPinned: Int
)

data class MuteConversationReq(
    @SerializedName("conversation_id") val conversationId: String,
    @SerializedName("is_muted") val isMuted: Int
)

data class ConvActionReq(
    @SerializedName("conversation_id") val conversationId: String
)
