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

    @GET("/api/user/search")
    suspend fun searchUsers(@Query("keyword") keyword: String): Response<ApiResponse<List<UserDto>>>

    @GET("/api/user/friends")
    suspend fun getFriends(): Response<ApiResponse<List<UserDto>>>

    @POST("/api/user/add-friend")
    suspend fun addFriend(@Body body: Map<String, Long>): Response<ApiResponse<Map<String, Any>>>

    @POST("/api/group/create")
    suspend fun createGroup(@Body req: GroupCreateRequest): Response<ApiResponse<GroupInfoDto>>

    @GET("/api/group/info")
    suspend fun getGroupInfo(@Query("groupId") groupId: Long): Response<ApiResponse<GroupInfoDto>>

    @GET("/api/group/members")
    suspend fun getGroupMembers(@Query("groupId") groupId: Long): Response<ApiResponse<List<GroupMemberDto>>>

    @Multipart
    @POST("/api/file/upload")
    suspend fun uploadFile(@Part file: MultipartBody.Part): Response<ApiResponse<FileContentMeta>>
}
