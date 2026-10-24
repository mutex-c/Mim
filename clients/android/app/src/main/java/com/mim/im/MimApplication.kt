package com.mim.im

import android.app.Application
import android.app.NotificationChannel
import android.app.NotificationManager
import android.os.Build
import com.mim.im.api.ApiClient
import com.mim.im.data.local.MimDatabase
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.network.WsManager
import com.mim.im.repository.AuthRepository
import com.mim.im.repository.ChatRepository

class MimApplication : Application() {

    companion object {
        const val CHANNEL_ID_MESSAGES = "mim_messages_channel"
        lateinit var instance: MimApplication
            private set
    }

    val database by lazy { MimDatabase.getDatabase(this) }
    val preferencesHelper by lazy { PreferencesHelper(this) }
    val apiClient by lazy { ApiClient(preferencesHelper) }
    val wsManager by lazy { WsManager(preferencesHelper, apiClient) }

    val chatRepository by lazy {
        ChatRepository(
            context = this,
            prefs = preferencesHelper,
            apiClient = apiClient,
            wsManager = wsManager,
            messageDao = database.messageDao(),
            conversationDao = database.conversationDao()
        )
    }

    val authRepository by lazy {
        AuthRepository(
            prefs = preferencesHelper,
            apiClient = apiClient,
            wsManager = wsManager
        )
    }

    override fun onCreate() {
        super.onCreate()
        instance = this

        createNotificationChannels()

        // Auto connect if logged in
        if (authRepository.isLoggedIn()) {
            wsManager.connect()
        }
    }

    private fun createNotificationChannels() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID_MESSAGES,
                "Mim 即时消息",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "接收来自 Mim 的单聊与群聊消息通知"
                enableVibration(true)
            }
            val manager = getSystemService(NotificationManager::class.java)
            manager.createNotificationChannel(channel)
        }
    }
}
