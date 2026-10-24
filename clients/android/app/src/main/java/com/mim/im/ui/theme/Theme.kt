package com.mim.im.ui.theme

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color

val LocalMimColors = staticCompositionLocalOf { PineLightColors }

object MimTheme {
    val colors: MimColors
        @Composable
        @ReadOnlyComposable
        get() = LocalMimColors.current
}

@Composable
fun MimTheme(
    palette: ColorPalette = ColorPalette.PINE,
    themeMode: ThemeMode = ThemeMode.SYSTEM,
    content: @Composable () -> Unit
) {
    val isSystemDark = isSystemInDarkTheme()
    val isDark = when (themeMode) {
        ThemeMode.SYSTEM -> isSystemDark
        ThemeMode.LIGHT -> false
        ThemeMode.DARK -> true
    }

    val mimColors = getMimColors(palette, isDark)

    val materialColors = if (isDark) {
        darkColorScheme(
            primary = mimColors.accent,
            onPrimary = Color.White,
            primaryContainer = mimColors.bubbleSelf,
            onPrimaryContainer = mimColors.bubbleSelfText,
            background = mimColors.background,
            surface = mimColors.surface,
            onSurface = mimColors.textPrimary,
            surfaceVariant = mimColors.surfaceElevated,
            onSurfaceVariant = mimColors.textMuted
        )
    } else {
        lightColorScheme(
            primary = mimColors.accent,
            onPrimary = Color.White,
            primaryContainer = mimColors.accentSubtle,
            onPrimaryContainer = mimColors.accent,
            background = mimColors.background,
            surface = mimColors.surface,
            onSurface = mimColors.textPrimary,
            surfaceVariant = mimColors.surfaceElevated,
            onSurfaceVariant = mimColors.textMuted
        )
    }

    CompositionLocalProvider(
        LocalMimColors provides mimColors
    ) {
        MaterialTheme(
            colorScheme = materialColors,
            content = content
        )
    }
}
