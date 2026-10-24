package com.mim.im.ui.components

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.ui.theme.MimTheme

@Composable
fun MimVoiceBubble(
    isPlaying: Boolean,
    progress: Float,
    isSelf: Boolean,
    onTogglePlay: () -> Unit,
    modifier: Modifier = Modifier
) {
    val infiniteTransition = rememberInfiniteTransition(label = "waveform")
    val animatedPhase by infiniteTransition.animateFloat(
        initialValue = 0f,
        targetValue = 1f,
        animationSpec = infiniteRepeatable(
            animation = tween(1000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "phase"
    )

    val contentColor = if (isSelf) MimTheme.colors.bubbleSelfText else MimTheme.colors.bubblePeerText
    val barBaseColor = contentColor.copy(alpha = 0.35f)
    val barActiveColor = contentColor

    Row(
        modifier = modifier
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onTogglePlay)
            .padding(horizontal = 8.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(10.dp)
    ) {
        // Play/Pause circle button
        Box(
            modifier = Modifier
                .size(32.dp)
                .clip(CircleShape)
                .background(contentColor.copy(alpha = 0.12f)),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = if (isPlaying) Icons.Default.Pause else Icons.Default.PlayArrow,
                contentDescription = if (isPlaying) "Pause" else "Play",
                tint = contentColor,
                modifier = Modifier.size(18.dp)
            )
        }

        // Waveform bars (8 bars)
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(3.dp),
            modifier = Modifier.width(90.dp).height(24.dp)
        ) {
            val barHeights = listOf(10.dp, 16.dp, 22.dp, 14.dp, 20.dp, 18.dp, 12.dp, 8.dp)
            barHeights.forEachIndexed { index, height ->
                val barProgress = index.toFloat() / barHeights.size
                val isBarActive = isPlaying && (progress >= barProgress)

                val effectiveHeight = if (isPlaying) {
                    val dynamicFactor = 0.6f + 0.4f * kotlin.math.sin((animatedPhase * 2 * Math.PI + index).toFloat()).coerceIn(0f, 1f)
                    height * dynamicFactor
                } else {
                    height
                }

                Box(
                    modifier = Modifier
                        .width(3.5.dp)
                        .height(effectiveHeight)
                        .clip(RoundedCornerShape(2.dp))
                        .background(if (isBarActive) barActiveColor else barBaseColor)
                )
            }
        }
    }
}
