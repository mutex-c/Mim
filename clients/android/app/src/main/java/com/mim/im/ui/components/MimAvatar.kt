package com.mim.im.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Group
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.mim.im.ui.theme.MimTheme
import kotlin.math.abs

// Bespoke solid grounding colors for avatars (anti-washed out, anti-pastel)
val AvatarPalette = listOf(
    Color(0xFF1E5243), // Forest Pine
    Color(0xFF1E4555), // Baltic Petrol
    Color(0xFF8C4C1B), // Terracotta Amber
    Color(0xFF2E4057), // Deep Slate Blue
    Color(0xFF5B3758), // Velvet Plum
    Color(0xFF4A5859), // Cool Pewter
    Color(0xFF7A3E2D), // Rich Ochre
    Color(0xFF2C5E43)  // Alpine Fir
)

fun getAvatarColor(key: String): Color {
    if (key.isBlank()) return AvatarPalette[0]
    val hash = abs(key.hashCode())
    return AvatarPalette[hash % AvatarPalette.size]
}

@Composable
fun MimAvatar(
    name: String,
    modifier: Modifier = Modifier,
    size: Dp = 44.dp,
    imageUrl: String? = null,
    isGroup: Boolean = false,
    isOnline: Boolean? = null
) {
    val initial = if (name.isNotBlank()) {
        val trimmed = name.trim()
        if (trimmed.length >= 2 && trimmed[0].isLetter() && trimmed[1].isLetter()) {
            trimmed.substring(0, 2).uppercase()
        } else {
            trimmed.take(1).uppercase()
        }
    } else {
        "M"
    }

    val shape = if (isGroup) RoundedCornerShape(size * 0.28f) else CircleShape
    val avatarBg = getAvatarColor(name)
    val fontSize = (size.value * 0.38f).sp

    Box(
        modifier = modifier.size(size),
        contentAlignment = Alignment.Center
    ) {
        if (!imageUrl.isNullOrBlank()) {
            AsyncImage(
                model = imageUrl,
                contentDescription = name,
                modifier = Modifier
                    .fillMaxSize()
                    .clip(shape)
            )
        } else {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .clip(shape)
                    .background(avatarBg),
                contentAlignment = Alignment.Center
            ) {
                if (isGroup) {
                    Icon(
                        imageVector = Icons.Default.Group,
                        contentDescription = "群聊",
                        tint = Color.White.copy(alpha = 0.95f),
                        modifier = Modifier.size(size * 0.52f)
                    )
                } else {
                    Text(
                        text = initial,
                        color = Color.White,
                        fontSize = fontSize,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }

        // Online dot status
        if (isOnline == true) {
            val dotSize = size * 0.28f
            val dotOffset = size * 0.04f
            Box(
                modifier = Modifier
                    .size(dotSize)
                    .align(Alignment.BottomEnd)
                    .offset(x = -dotOffset, y = -dotOffset)
                    .clip(CircleShape)
                    .background(Color(0xFF22C55E))
                    .border(1.5.dp, MimTheme.colors.surface, CircleShape)
            )
        }
    }
}
