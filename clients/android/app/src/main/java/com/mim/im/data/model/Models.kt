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
    @SerializedName(value = "id", alternate = ["friend_id", "user_id"]) val id: Long = 0,
    val account: String = "",
    val nickname: String = "",
    val avatar: String = "",
    val signature: String = "",
    val role: Int = 1
) {
    val effectiveId: Long get() = id
    val displayName: String get() = nickname.ifBlank { account.ifBlank { "用户 $id" } }
}

data class DemoUserDto(
    val user: UserDto,
    val token: TokenPairDto
)

data class UserProfileRespDto(
    val user: UserDto,
    val config: Map<String, Any>? = null
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
    @SerializedName("msg_id", alternate = ["msgId", "MsgId", "id", "ID"]) val msgId: Long = 0,
    @SerializedName("client_msg_id", alternate = ["clientMsgId", "ClientMsgId"]) val clientMsgId: String = "",
    @SerializedName("conversation_id", alternate = ["conversationId"]) val conversationId: String = "",
    @SerializedName("conv_type", alternate = ["convType"]) val convType: Int = ConvType.SINGLE,
    @SerializedName("sender_id", alternate = ["senderId"]) val senderId: Long = 0,
    @SerializedName("receiver_id", alternate = ["receiverId"]) val receiverId: Long = 0,
    @SerializedName("msg_type", alternate = ["msgType"]) val msgType: Int = MsgType.TEXT,
    val content: String = "",
    @SerializedName("quoted_msg_id", alternate = ["quotedMsgId"]) val quotedMsgId: Long = 0,
    @SerializedName("quoted_content", alternate = ["quotedContent"]) val quotedContent: String = "",
    @SerializedName("seq_id", alternate = ["seqId", "SeqId"]) val seqId: Long = 0,
    @SerializedName("send_time", alternate = ["sendTime", "CreatedAt", "created_at"]) val sendTime: Long = System.currentTimeMillis(),
    @SerializedName("is_encrypted", alternate = ["isEncrypted"]) val isEncrypted: Boolean = false
)

data class WsMessageAck(
    @SerializedName("client_msg_id", alternate = ["clientMsgId", "ClientMsgId"]) val clientMsgId: String = "",
    @SerializedName("msg_id", alternate = ["msgId", "MsgId", "id", "ID"]) val msgId: Long = 0,
    @SerializedName("seq_id", alternate = ["seqId", "SeqId"]) val seqId: Long = 0,
    @SerializedName("code", alternate = ["Code"]) val code: Int = 0,
    @SerializedName("error_msg", alternate = ["errorMsg", "ErrorMsg"]) val errorMsg: String = "",
    @SerializedName("server_time", alternate = ["serverTime", "ServerTime"]) val serverTime: Long = 0
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
    @SerializedName("url", alternate = ["file_url", "fileUrl"]) val url: String = "",
    @SerializedName("name", alternate = ["file_name", "fileName"]) val name: String = "",
    @SerializedName("size", alternate = ["file_size", "fileSize"]) val size: Long = 0,
    @SerializedName("mime", alternate = ["mime_type", "mimeType"]) val mime: String = "",
    @SerializedName("file_id", alternate = ["id", "fileId"]) val fileId: Long = 0
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
    @SerializedName("sender_id") val senderId: Long = 0,
    @SerializedName("from_user_id") val fromUserId: Long = 0,
    @SerializedName("to_user_id") val toUserId: Long = 0,
    @SerializedName("nickname") val nickname: String = "",
    @SerializedName("from_nickname") val fromNickname: String = "",
    @SerializedName("avatar") val avatar: String = "",
    @SerializedName("from_avatar") val fromAvatar: String = "",
    @SerializedName("account") val account: String = "",
    @SerializedName("from_account") val fromAccount: String = "",
    @SerializedName("verify_msg") val verifyMsg: String = "",
    val message: String = "",
    val status: Int = 1,
    @SerializedName("created_at") val createdAt: String = ""
) {
    val effectiveSenderId: Long get() = if (senderId > 0) senderId else fromUserId
    val effectiveNickname: String get() = nickname.ifBlank { fromNickname.ifBlank { effectiveAccount } }
    val effectiveAccount: String get() = account.ifBlank { fromAccount.ifBlank { "User$effectiveSenderId" } }
    val effectiveMessage: String get() = verifyMsg.ifBlank { message }
}

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

// --- AI Summary DTOs ---

data class AISummaryReq(
    @SerializedName("conversation_id") val conversationId: String,
    val lang: String = "zh"
)

data class AISummaryResp(
    val summary: String = "",
    val count: Int = 0
)

// --- WebRTC Call Signaling DTOs ---

data class WebRTCSignalDto(
    @SerializedName("call_id") val callId: String = "",
    @SerializedName("from_user_id") val fromUserId: Long = 0,
    @SerializedName("to_user_id") val toUserId: Long = 0,
    @SerializedName("signal_type") val signalType: String = "", // "OFFER", "ANSWER", "CANDIDATE", "HANGUP", "REJECT"
    val data: String = ""
)

data class CallHistoryItemDto(
    val id: String = "",
    @SerializedName("from_user") val fromUser: String = "",
    @SerializedName("to_user") val toUser: String = "",
    @SerializedName("call_type") val callType: String = "", // "video" / "audio"
    val status: String = "", // "Completed" / "Missed" / "Rejected"
    val duration: Int = 0,
    @SerializedName("created_at") val createdAt: String = "",
    @SerializedName("is_caller") val isCaller: Boolean = false
)

data class CallSessionState(
    val callId: String,
    val peerId: Long,
    val peerName: String,
    val isVideo: Boolean,
    val isOutgoing: Boolean,
    val isConnected: Boolean = false,
    val durationSeconds: Int = 0
)

// --- User Profile & Configuration DTOs ---

data class UpdateProfileReq(
    val nickname: String,
    val avatar: String = "",
    val signature: String = ""
)

data class UpdateConfigReq(
    @SerializedName("search_by_account") val searchByAccount: Int = 1,
    @SerializedName("search_by_nickname") val searchByNickname: Int = 1,
    @SerializedName("friend_verify_mode") val friendVerifyMode: Int = 2
)

// --- AI Configuration & Smart Reply DTOs ---

data class AIConfigDto(
    @SerializedName("provider") val provider: String = "deepseek",
    @SerializedName("base_url") val baseUrl: String = "",
    @SerializedName("api_key") val apiKey: String = "",
    @SerializedName("model") val model: String = "",
    @SerializedName("is_custom") val isCustom: Boolean = false
)

data class AITestReq(
    @SerializedName("base_url") val baseUrl: String,
    @SerializedName("api_key") val apiKey: String,
    @SerializedName("model") val model: String
)

data class AITestResp(
    @SerializedName("latency_ms") val latencyMs: Long = 0
)

data class AIReplyReq(
    @SerializedName("conversation_id") val conversationId: String,
    @SerializedName("count") val count: Int = 3
)

data class AIReplyItem(
    @SerializedName("type") val type: String = "general",
    @SerializedName("text") val text: String = "",
    @SerializedName("label") val label: String? = null
)

data class AIReplyResp(
    @SerializedName("suggestions") val suggestions: List<AIReplyItem> = emptyList()
)

