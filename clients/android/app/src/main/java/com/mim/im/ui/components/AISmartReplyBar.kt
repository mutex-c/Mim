package com.mim.im.ui.components

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.Send
import androidx.compose.material.icons.filled.AutoAwesome
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.model.AIReplyItem
import com.mim.im.ui.theme.MimTheme

@Composable
fun AISmartReplyBar(
    suggestions: List<AIReplyItem>,
    isLoading: Boolean,
    onSelect: (String) -> Unit,
    onSend: (String) -> Unit,
    onRefresh: () -> Unit,
    onDismiss: () -> Unit,
    modifier: Modifier = Modifier
) {
    if (suggestions.isEmpty() && !isLoading) return

    Surface(
        shape = RoundedCornerShape(12.dp),
        color = MimTheme.colors.surfaceElevated.copy(alpha = 0.95f),
        modifier = modifier
            .fillMaxWidth()
            .padding(horizontal = 12.dp, vertical = 4.dp)
            .border(1.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(8.dp)
        ) {
            // Header
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(22.dp)
                            .clip(RoundedCornerShape(6.dp))
                            .background(MimTheme.colors.accent.copy(alpha = 0.15f)),
                        contentAlignment = Alignment.Center
                    ) {
                        Icon(
                            imageVector = Icons.Default.AutoAwesome,
                            contentDescription = null,
                            tint = MimTheme.colors.accent,
                            modifier = Modifier.size(13.dp)
                        )
                    }
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = "AI 智能帮回",
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.textPrimary
                    )
                }

                Row(verticalAlignment = Alignment.CenterVertically) {
                    IconButton(
                        onClick = onRefresh,
                        enabled = !isLoading,
                        modifier = Modifier.size(24.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Refresh,
                            contentDescription = "换一批",
                            tint = MimTheme.colors.textMuted,
                            modifier = Modifier.size(14.dp)
                        )
                    }
                    Spacer(modifier = Modifier.width(2.dp))
                    IconButton(
                        onClick = onDismiss,
                        modifier = Modifier.size(24.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Close,
                            contentDescription = "关闭",
                            tint = MimTheme.colors.textMuted,
                            modifier = Modifier.size(14.dp)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(6.dp))

            if (isLoading) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 8.dp),
                    horizontalArrangement = Arrangement.Center,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(14.dp),
                        strokeWidth = 2.dp,
                        color = MimTheme.colors.accent
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Text(
                        text = "正在思考智能回复...",
                        fontSize = 11.sp,
                        color = MimTheme.colors.textMuted
                    )
                }
            } else {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    suggestions.forEach { item ->
                        Surface(
                            shape = RoundedCornerShape(10.dp),
                            color = MimTheme.colors.surface,
                            modifier = Modifier
                                .widthIn(max = 240.dp)
                                .clip(RoundedCornerShape(10.dp))
                                .border(1.dp, MimTheme.colors.border.copy(alpha = 0.6f), RoundedCornerShape(10.dp))
                                .clickable { onSelect(item.text) }
                        ) {
                            Column(
                                modifier = Modifier.padding(horizontal = 8.dp, vertical = 6.dp)
                            ) {
                                Row(
                                    modifier = Modifier.fillMaxWidth(),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.SpaceBetween
                                ) {
                                    val badgeLabel = item.label ?: when (item.type) {
                                        "polite", "friendly" -> "亲和友好"
                                        "professional" -> "严谨商务"
                                        "concise" -> "简明确认"
                                        else -> "推荐回复"
                                    }
                                    val badgeColor = when (item.type) {
                                        "polite", "friendly" -> Color(0xFF10B981)
                                        "professional" -> Color(0xFF3B82F6)
                                        "concise" -> Color(0xFF8B5CF6)
                                        else -> MimTheme.colors.accent
                                    }

                                    Box(
                                        modifier = Modifier
                                            .clip(RoundedCornerShape(4.dp))
                                            .background(badgeColor.copy(alpha = 0.12f))
                                            .padding(horizontal = 5.dp, vertical = 1.dp)
                                    ) {
                                        Text(
                                            text = badgeLabel,
                                            fontSize = 9.sp,
                                            fontWeight = FontWeight.Bold,
                                            color = badgeColor
                                        )
                                    }

                                    Spacer(modifier = Modifier.width(8.dp))

                                    Icon(
                                        imageVector = Icons.AutoMirrored.Filled.Send,
                                        contentDescription = "直接发送",
                                        tint = MimTheme.colors.accent,
                                        modifier = Modifier
                                            .size(13.dp)
                                            .clip(CircleShape)
                                            .clickable { onSend(item.text) }
                                    )
                                }

                                Spacer(modifier = Modifier.height(4.dp))

                                Text(
                                    text = item.text,
                                    fontSize = 11.sp,
                                    lineHeight = 15.sp,
                                    color = MimTheme.colors.textPrimary,
                                    maxLines = 2,
                                    overflow = TextOverflow.Ellipsis
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}
