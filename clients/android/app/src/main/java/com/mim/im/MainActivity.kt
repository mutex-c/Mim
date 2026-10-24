package com.mim.im

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.navigation.NavType
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.navArgument
import com.mim.im.ui.screens.AuthScreen
import com.mim.im.ui.screens.ChatScreen
import com.mim.im.ui.screens.MainScreen
import com.mim.im.ui.theme.MimTheme
import java.net.URLDecoder
import java.net.URLEncoder

class MainActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        enableEdgeToEdge()

        val app = application as MimApplication

        setContent {
            val palette by app.preferencesHelper.paletteFlow.collectAsState()
            val themeMode by app.preferencesHelper.themeModeFlow.collectAsState()

            MimTheme(palette = palette, themeMode = themeMode) {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MimTheme.colors.background
                ) {
                    MimNavHost(app = app)
                }
            }
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
                    val encodedTitle = URLEncoder.encode(title, "UTF-8")
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
                onBack = { navController.popBackStack() }
            )
        }
    }
}
