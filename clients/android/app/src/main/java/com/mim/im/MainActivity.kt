package com.mim.im

import android.Manifest
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.core.content.ContextCompat
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.mim.im.ui.components.ActiveCallOverlay
import com.mim.im.ui.components.IncomingCallDialog
import com.mim.im.ui.screens.AuthScreen
import com.mim.im.ui.screens.ChatScreen
import com.mim.im.ui.screens.MainScreen
import com.mim.im.ui.theme.MimTheme
import java.net.URLDecoder
import java.net.URLEncoder

class MainActivity : ComponentActivity() {

    private val requestPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { _ ->
        // Permissions handled
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        // Check & request runtime permissions for audio calls, video calls and push notifications
        checkAndRequestPermissions()

        val app = application as MimApplication

        setContent {
            val palette by app.preferencesHelper.paletteFlow.collectAsState()
            val themeMode by app.preferencesHelper.themeModeFlow.collectAsState()
            val incomingCall by app.chatRepository.incomingCall.collectAsState()
            val activeCall by app.chatRepository.activeCall.collectAsState()

            MimTheme(palette = palette, themeMode = themeMode) {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MimTheme.colors.background
                ) {
                    Box(modifier = Modifier.fillMaxSize()) {
                        MimNavHost(app = app)

                        // Global Incoming Call Invitation Dialog
                        incomingCall?.let { inc ->
                            IncomingCallDialog(
                                callState = inc,
                                onAccept = { app.chatRepository.acceptIncomingCall() },
                                onReject = { app.chatRepository.rejectIncomingCall() }
                            )
                        }

                        // Global Active Call Fullscreen Stage
                        activeCall?.let { act ->
                            ActiveCallOverlay(
                                callState = act,
                                webRtcManager = app.webRtcManager,
                                chatRepository = app.chatRepository,
                                onEndCall = { app.chatRepository.endActiveCall() }
                            )
                        }
                    }
                }
            }
        }
    }

    private fun checkAndRequestPermissions() {
        val permissionsToRequest = mutableListOf(
            Manifest.permission.RECORD_AUDIO,
            Manifest.permission.CAMERA
        )
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            permissionsToRequest.add(Manifest.permission.POST_NOTIFICATIONS)
        }
        val needed = permissionsToRequest.filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }
        if (needed.isNotEmpty()) {
            requestPermissionLauncher.launch(needed.toTypedArray())
        }
    }
}

@Composable
fun MimNavHost(app: MimApplication) {
    val navController = rememberNavController()
    val startDestination = if (app.authRepository.isLoggedIn()) "main" else "auth"

    NavHost(
        navController = navController,
        startDestination = startDestination
    ) {
        composable("auth") {
            AuthScreen(
                authRepo = app.authRepository,
                prefs = app.preferencesHelper,
                onAuthSuccess = {
                    navController.navigate("main") {
                        popUpTo("auth") { inclusive = true }
                    }
                }
            )
        }

        composable("main") {
            MainScreen(
                prefs = app.preferencesHelper,
                authRepo = app.authRepository,
                chatRepo = app.chatRepository,
                wsManager = app.wsManager,
                onOpenChat = { convId, title, targetId, isGroup ->
                    val safeTitle = if (title.isBlank()) (if (isGroup) "群聊会话" else "私聊会话") else title
                    val encodedTitle = URLEncoder.encode(safeTitle, "UTF-8")
                    navController.navigate("chat/$convId/$encodedTitle/$targetId/$isGroup")
                },
                onLogout = {
                    app.authRepository.logout()
                    navController.navigate("auth") {
                        popUpTo("main") { inclusive = true }
                    }
                }
            )
        }

        composable(
            route = "chat/{convId}/{title}/{targetId}/{isGroup}",
            arguments = listOf(
                navArgument("convId") { type = NavType.StringType },
                navArgument("title") { type = NavType.StringType },
                navArgument("targetId") { type = NavType.LongType },
                navArgument("isGroup") { type = NavType.BoolType }
            )
        ) { backStackEntry ->
            val convId = backStackEntry.arguments?.getString("convId") ?: ""
            val rawTitle = backStackEntry.arguments?.getString("title") ?: ""
            val title = try { URLDecoder.decode(rawTitle, "UTF-8") } catch (_: Exception) { rawTitle }
            val targetId = backStackEntry.arguments?.getLong("targetId") ?: 0L
            val isGroup = backStackEntry.arguments?.getBoolean("isGroup") ?: false

            ChatScreen(
                conversationId = convId,
                title = title,
                targetId = targetId,
                isGroup = isGroup,
                chatRepo = app.chatRepository,
                authRepo = app.authRepository,
                onBack = { navController.popBackStack() }
            )
        }
    }
}
