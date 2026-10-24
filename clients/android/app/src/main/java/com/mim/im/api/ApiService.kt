package com.mim.im.api

import com.mim.im.data.model.*
import okhttp3.MultipartBody
import retrofit2.Response
import retrofit2.http.*

interface ApiService {
    @POST("/api/auth/register")
    suspend fun register(@Body req: RegisterRequest): Response<ApiResponse<UserDto>>

    @POST("/api/auth/login")
    suspend fun login(@Body req: LoginRequest): Response<ApiResponse<LoginResponseData>>

    @GET("/api/auth/demo-users")
    suspend fun getDemoUsers(): Response<ApiResponse<List<UserDto>>>

    @GET("/api/user/profile")
    suspend fun getProfile(): Response<ApiResponse<UserDto>>

    @GET("/api/chat/conversations")
    suspend fun getConversations(): Response<ApiResponse<List<ConversationDto>>>

    @GET("/api/chat/messages")
    suspend fun getMessages(
        @Query("conv_id") convId: String,
        @Query("last_seq_id") lastSeqId: Long = 0,
        @Query("limit") limit: Int = 50
    ): Response<ApiResponse<List<WsChatMessage>>>

    @POST("/api/chat/revoke")
    suspend fun revokeMessage(@Body req: RevokeRequest): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/chat/read-conversation")
    suspend fun readConversation(@Body req: ReadConversationRequest): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/chat/pin-conversation")
    suspend fun pinConversation(@Body req: PinConversationReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/chat/mute-conversation")
    suspend fun muteConversation(@Body req: MuteConversationReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/chat/delete-conversation")
    suspend fun deleteConversation(@Body req: ConvActionReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/chat/clear-messages")
    suspend fun clearMessages(@Body req: ConvActionReq): Response<ApiResponse<Map<String, Any>>>

    @GET("/api/user/search")
    suspend fun searchUsers(@Query("keyword") keyword: String): Response<ApiResponse<List<UserDto>>>

    @GET("/api/user/friends")
    suspend fun getFriends(): Response<ApiResponse<List<UserDto>>>

    @POST("/api/user/add-friend")
    suspend fun addFriend(@Body req: AddFriendReq): Response<ApiResponse<Map<String, Any>>>

    @GET("/api/user/friend-requests")
    suspend fun getFriendRequests(): Response<ApiResponse<List<FriendRequestDto>>>

    @POST("/api/user/handle-friend-request")
    suspend fun handleFriendRequest(@Body req: HandleFriendRequestReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/user/delete-friend")
    suspend fun deleteFriend(@Body req: DeleteFriendReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/create")
    suspend fun createGroup(@Body req: CreateGroupReq): Response<ApiResponse<GroupInfoDto>>

    @GET("/api/group/info")
    suspend fun getGroupInfo(@Query("group_id") groupId: Long): Response<ApiResponse<GroupInfoDto>>

    @GET("/api/group/members")
    suspend fun getGroupMembers(@Query("group_id") groupId: Long): Response<ApiResponse<List<GroupMemberDto>>>

    @POST("/api/group/update")
    suspend fun updateGroup(@Body req: UpdateGroupReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/invite")
    suspend fun inviteGroup(@Body req: InviteGroupReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/mute-member")
    suspend fun muteMember(@Body req: MuteMemberReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/kick-member")
    suspend fun kickMember(@Body req: KickMemberReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/disband")
    suspend fun disbandGroup(@Body req: GroupIdReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/set-role")
    suspend fun setGroupRole(@Body req: SetGroupRoleReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/transfer-owner")
    suspend fun transferGroupOwner(@Body req: TransferGroupOwnerReq): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/quit")
    suspend fun quitGroup(@Body req: GroupIdReq): Response<ApiResponse<Map<String, Any>>>

    @Multipart
    @POST("/api/file/upload")
    suspend fun uploadFile(@Part file: MultipartBody.Part): Response<ApiResponse<FileContentMeta>>
}
