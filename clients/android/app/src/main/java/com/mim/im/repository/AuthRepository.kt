package com.mim.im.repository

import com.mim.im.api.ApiClient
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.data.model.*
import com.mim.im.network.WsManager
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class AuthRepository(
    private val prefs: PreferencesHelper,
    private val apiClient: ApiClient,
    private val wsManager: WsManager
) {
    suspend fun login(account: String, pass: String): Result<UserDto> = withContext(Dispatchers.IO) {
        try {
            val req = LoginRequest(account = account, password = pass)
            val resp = apiClient.service.login(req)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                val data = resp.body()?.data ?: return@withContext Result.failure(Exception("Empty login data"))
                prefs.token = data.token.accessToken
                prefs.currentUser = UserInfo(
                    id = data.user.id.toString(),
                    username = data.user.account,
                    nickname = data.user.nickname,
                    avatar = data.user.avatar
                )
                wsManager.connect()
                Result.success(data.user)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Login failed (${resp.code()})"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun register(account: String, pass: String, nickname: String): Result<UserDto> = withContext(Dispatchers.IO) {
        try {
            val req = RegisterRequest(account = account, password = pass, nickname = nickname)
            val resp = apiClient.service.register(req)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                val user = resp.body()?.data ?: return@withContext Result.failure(Exception("Empty register data"))
                Result.success(user)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Registration failed (${resp.code()})"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getDemoUsers(): Result<List<UserDto>> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.getDemoUsers()
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(resp.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to fetch demo users"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getFriends(): Result<List<UserDto>> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.getFriends()
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(resp.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to fetch friends"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun searchUsers(keyword: String): Result<List<UserDto>> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.searchUsers(keyword)
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(resp.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to search users"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun sendFriendRequest(targetUserId: Long, msg: String): Result<String> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.addFriend(AddFriendReq(targetId = targetUserId, msg = msg))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                val message = resp.body()?.data?.get("msg")?.toString() ?: "申请已发送"
                Result.success(message)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to send friend request"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun getFriendRequests(): Result<List<FriendRequestDto>> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.getFriendRequests()
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(resp.body()?.data ?: emptyList())
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to get friend requests"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun handleFriendRequest(requestId: Long, action: Int): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.handleFriendRequest(HandleFriendRequestReq(requestId = requestId, action = action))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to process friend request"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteFriend(friendId: Long): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            val resp = apiClient.service.deleteFriend(DeleteFriendReq(friendId = friendId))
            if (resp.isSuccessful && resp.body()?.code == 0) {
                Result.success(Unit)
            } else {
                Result.failure(Exception(resp.body()?.msg ?: "Failed to delete friend"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun logout() {
        wsManager.disconnect()
        prefs.clearAuth()
    }

    fun isLoggedIn(): Boolean = !prefs.token.isNullOrEmpty()
}
