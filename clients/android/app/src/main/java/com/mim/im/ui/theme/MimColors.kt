package com.mim.im.ui.theme

import androidx.compose.ui.graphics.Color

enum class ColorPalette(val id: String, val nameZh: String, val nameEn: String, val previewColor: Color) {
    PINE("pine", "松风墨翠", "Nordic Forest Pine", Color(0xFF1E5243)),
    PETROL("petrol", "沧浪深黛", "Baltic Sea Petrol", Color(0xFF1E4555)),
    OBSIDIAN("obsidian", "黑曜暖琥", "Obsidian & Amber", Color(0xFF8C4C1B));

    companion object {
        fun fromId(id: String): ColorPalette = entries.firstOrNull { it.id == id } ?: PINE
    }
}

enum class ThemeMode(val id: String, val nameZh: String) {
    SYSTEM("system", "跟随系统"),
    LIGHT("light", "白天模式"),
    DARK("dark", "暗黑模式");

    companion object {
        fun fromId(id: String): ThemeMode = entries.firstOrNull { it.id == id } ?: SYSTEM
    }
}

data class MimColors(
    val background: Color,
    val surface: Color,
    val surfaceElevated: Color,
    val border: Color,
    val textPrimary: Color,
    val textMuted: Color,
    val accent: Color,
    val accentSubtle: Color,
    val bubbleSelf: Color,
    val bubbleSelfText: Color,
    val bubblePeer: Color,
    val bubblePeerText: Color,
    val inputBackground: Color,
    val isDark: Boolean
)

val PineLightColors = MimColors(
    background = Color(0xFFE7E6DE),
    surface = Color(0xFFF4F3EC),
    surfaceElevated = Color(0xFFFFFFFF),
    border = Color(0xFFDCDAD0),
    textPrimary = Color(0xFF191B1A),
    textMuted = Color(0xFF646A66),
    accent = Color(0xFF1E5243),
    accentSubtle = Color(0x1F1E5243),
    bubbleSelf = Color(0xFF1E5243),
    bubbleSelfText = Color(0xFFFFFFFF),
    bubblePeer = Color(0xFFFFFFFF),
    bubblePeerText = Color(0xFF191B1A),
    inputBackground = Color(0xFFFFFFFF),
    isDark = false
)

val PineDarkColors = MimColors(
    background = Color(0xFF141715),
    surface = Color(0xFF151816),
    surfaceElevated = Color(0xFF202623),
    border = Color(0xFF2A332E),
    textPrimary = Color(0xFFE6EBE8),
    textMuted = Color(0xFF8B958F),
    accent = Color(0xFF2E6E5A),
    accentSubtle = Color(0x2834D399),
    bubbleSelf = Color(0xFF235445),
    bubbleSelfText = Color(0xFFF0FDF4),
    bubblePeer = Color(0xFF232926),
    bubblePeerText = Color(0xFFE6EBE8),
    inputBackground = Color(0xFF1A201D),
    isDark = true
)

val PetrolLightColors = MimColors(
    background = Color(0xFFE2E8EA),
    surface = Color(0xFFEEF2F4),
    surfaceElevated = Color(0xFFFFFFFF),
    border = Color(0xFFD4DCDE),
    textPrimary = Color(0xFF161C20),
    textMuted = Color(0xFF5E6B72),
    accent = Color(0xFF1E4555),
    accentSubtle = Color(0x1F1E4555),
    bubbleSelf = Color(0xFF1E4555),
    bubbleSelfText = Color(0xFFFFFFFF),
    bubblePeer = Color(0xFFFFFFFF),
    bubblePeerText = Color(0xFF161C20),
    inputBackground = Color(0xFFFFFFFF),
    isDark = false
)

val PetrolDarkColors = MimColors(
    background = Color(0xFF111619),
    surface = Color(0xFF141B1F),
    surfaceElevated = Color(0xFF1D262C),
    border = Color(0xFF27343C),
    textPrimary = Color(0xFFE4EBF0),
    textMuted = Color(0xFF82939E),
    accent = Color(0xFF2A5F75),
    accentSubtle = Color(0x2838BDF8),
    bubbleSelf = Color(0xFF1E4555),
    bubbleSelfText = Color(0xFFF0F7FA),
    bubblePeer = Color(0xFF1D262C),
    bubblePeerText = Color(0xFFE4EBF0),
    inputBackground = Color(0xFF182126),
    isDark = true
)

val ObsidianLightColors = MimColors(
    background = Color(0xFFE8E5E0),
    surface = Color(0xFFF3F0EB),
    surfaceElevated = Color(0xFFFFFFFF),
    border = Color(0xFFDDD8CF),
    textPrimary = Color(0xFF1C1917),
    textMuted = Color(0xFF6B6661),
    accent = Color(0xFF8C4C1B),
    accentSubtle = Color(0x1F8C4C1B),
    bubbleSelf = Color(0xFF8C4C1B),
    bubbleSelfText = Color(0xFFFFFFFF),
    bubblePeer = Color(0xFFFFFFFF),
    bubblePeerText = Color(0xFF1C1917),
    inputBackground = Color(0xFFFFFFFF),
    isDark = false
)

val ObsidianDarkColors = MimColors(
    background = Color(0xFF151413),
    surface = Color(0xFF1A1816),
    surfaceElevated = Color(0xFF24221F),
    border = Color(0xFF332F2A),
    textPrimary = Color(0xFFECE8E1),
    textMuted = Color(0xFF948E85),
    accent = Color(0xFFB45309),
    accentSubtle = Color(0x28F59E0B),
    bubbleSelf = Color(0xFF78350F),
    bubbleSelfText = Color(0xFFFFFBEB),
    bubblePeer = Color(0xFF24221F),
    bubblePeerText = Color(0xFFECE8E1),
    inputBackground = Color(0xFF1F1D1A),
    isDark = true
)

fun getMimColors(palette: ColorPalette, darkTheme: Boolean): MimColors {
    return when (palette) {
        ColorPalette.PINE -> if (darkTheme) PineDarkColors else PineLightColors
        ColorPalette.PETROL -> if (darkTheme) PetrolDarkColors else PetrolLightColors
        ColorPalette.OBSIDIAN -> if (darkTheme) ObsidianDarkColors else ObsidianLightColors
    }
}
