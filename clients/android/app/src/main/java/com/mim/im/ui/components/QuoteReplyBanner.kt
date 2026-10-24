package com.mim.im.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.FormatQuote
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.local.MessageEntity
import com.mim.im.ui.theme.MimTheme

@Composable
fun QuoteReplyBanner(
    quotedMessage: MessageEntity,
    onDismiss: () -> Unit,
    modifier: Modifier = Modifier
) {
    val displayContent = when (quotedMessage.contentType) {
        "image" -> "[图片]"
        "voice" -> "[语音留言]"
        "file" -> "[文件] ${quotedMessage.fileName ?: ""}"
        else -> quotedMessage.content
    }

    Row(
        modifier = modifier
            .fillMaxWidth()
            .background(MimTheme.colors.surfaceElevated)
            .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(0.dp))
            .padding(horizontal = 12.dp, vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        // Accent vertical bar
        Box(
            modifier = Modifier
                .width(3.dp)
                .height(30.dp)
                .clip(RoundedCornerShape(1.5.dp))
                .background(MimTheme.colors.accent)
        )

        Spacer(modifier = Modifier.width(8.dp))

        Column(modifier = Modifier.weight(1f)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    imageVector = Icons.Default.FormatQuote,
                    contentDescription = null,
                    tint = MimTheme.colors.accent,
                    modifier = Modifier.size(14.dp)
                )
                Spacer(modifier = Modifier.width(4.dp))
                Text(
                    text = "回复消息",
                    style = MaterialTheme.typography.bodySmall.copy(
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.accent,
                        fontSize = 11.sp
                    )
                )
            }
            Text(
                text = displayContent,
                style = MaterialTheme.typography.bodySmall.copy(
                    color = MimTheme.colors.textSecondary,
                    fontSize = 12.sp
                ),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis
            )
        }

        IconButton(
            onClick = onDismiss,
            modifier = Modifier.size(28.dp)
        ) {
            Icon(
                imageVector = Icons.Default.Close,
                contentDescription = "取消回复",
                tint = MimTheme.colors.textMuted,
                modifier = Modifier.size(16.dp)
            )
        }
    }
}

@Composable
fun InBubbleQuoteView(
    quotedContent: String,
    isSelf: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val accentColor = if (isSelf) MimTheme.colors.bubbleSelfText.copy(alpha = 0.8f) else MimTheme.colors.accent
    val bg = if (isSelf) Color.Black.copy(alpha = 0.1f) else MimTheme.colors.background.copy(alpha = 0.6f)
    val textColor = if (isSelf) MimTheme.colors.bubbleSelfText.copy(alpha = 0.85f) else MimTheme.colors.textSecondary

    Row(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(6.dp))
            .clickable(onClick = onClick)
            .background(bg)
            .padding(horizontal = 8.dp, vertical = 5.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .width(2.5.dp)
                .height(20.dp)
                .clip(RoundedCornerShape(1.dp))
                .background(accentColor)
        )
        Spacer(modifier = Modifier.width(6.dp))
        Text(
            text = quotedContent,
            style = MaterialTheme.typography.bodySmall.copy(
                color = textColor,
                fontSize = 11.sp,
                lineHeight = 14.sp
            ),
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.weight(1f)
        )
    }
}
