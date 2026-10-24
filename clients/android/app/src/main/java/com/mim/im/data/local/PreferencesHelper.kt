package com.mim.im.data.local

import android.content.Context
import android.content.SharedPreferences
import com.google.gson.Gson
import com.mim.im.data.model.UserInfo
import com.mim.im.ui.theme.ColorPalette
import com.mim.im.ui.theme.ThemeMode
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

class PreferencesHelper(context: Context) {
    private val prefs: SharedPreferences = context.getSharedPreferences("mim_prefs", Context.MODE_PRIVATE)
    private val gson = Gson()

    companion object {
        private const val KEY_SERVER_URL = "server_url"
        private const val KEY_TOKEN = "jwt_token"
        private const val KEY_USER_INFO = "user_info"
        private const val KEY_THEME_MODE = "theme_mode"
        private const val KEY_PALETTE = "palette"
        const val DEFAULT_SERVER_URL = "http://10.0.2.2:8080"
    }

    var serverUrl: String
        get() = prefs.getString(KEY_SERVER_URL, DEFAULT_SERVER_URL) ?: DEFAULT_SERVER_URL
        set(value) = prefs.edit().putString(KEY_SERVER_URL, value.trimEnd('/')).apply()

    val wsUrl: String
        get() {
            val base = serverUrl
            val wsBase = when {
                base.startsWith("https://") -> base.replaceFirst("https://", "wss://")
                base.startsWith("http://") -> base.replaceFirst("http://", "ws://")
                else -> "ws://$base"
            }
            return "$wsBase/ws"
        }

    var token: String?
        get() = prefs.getString(KEY_TOKEN, null)
        set(value) = prefs.edit().putString(KEY_TOKEN, value).apply()

    var currentUser: UserInfo?
        get() {
            val json = prefs.getString(KEY_USER_INFO, null) ?: return null
            return try {
                gson.fromJson(json, UserInfo::class.java)
            } catch (e: Exception) {
                null
            }
        }
        set(value) {
            if (value == null) {
                prefs.edit().remove(KEY_USER_INFO).apply()
            } else {
                prefs.edit().putString(KEY_USER_INFO, gson.toJson(value)).apply()
            }
        }

    private val _themeModeFlow = MutableStateFlow(
        ThemeMode.fromId(prefs.getString(KEY_THEME_MODE, ThemeMode.SYSTEM.id) ?: ThemeMode.SYSTEM.id)
    )
    val themeModeFlow = _themeModeFlow.asStateFlow()

    var themeMode: ThemeMode
        get() = _themeModeFlow.value
        set(value) {
            _themeModeFlow.value = value
            prefs.edit().putString(KEY_THEME_MODE, value.id).apply()
        }

    private val _paletteFlow = MutableStateFlow(
        ColorPalette.fromId(prefs.getString(KEY_PALETTE, ColorPalette.PINE.id) ?: ColorPalette.PINE.id)
    )
    val paletteFlow = _paletteFlow.asStateFlow()

    var palette: ColorPalette
        get() = _paletteFlow.value
        set(value) {
            _paletteFlow.value = value
            prefs.edit().putString(KEY_PALETTE, value.id).apply()
        }

    fun clearAuth() {
        prefs.edit().remove(KEY_TOKEN).remove(KEY_USER_INFO).apply()
    }
}
