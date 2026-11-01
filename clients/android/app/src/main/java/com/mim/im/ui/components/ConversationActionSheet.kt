package com.mim.im.ui.components

import android.widget.Toast
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.local.ConversationEntity
import com.mim.im.repository.AuthRepository
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ConversationActionSheet(
    conv: ConversationEntity,
    chatRepo: ChatRepository,
    authRepo: AuthRepository,
    onDismiss: () -> Unit,
    onActionCompleted: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val isGroup = conv.convType == "group"
    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)

    ModalBottomSheet(
        onDismissRequest = onDismiss,
        sheetState = sheetState,
        containerColor = MimTheme.colors.surfaceElevated,
        shape = RoundedCornerShape(topStart = 20.dp, topEnd = 20.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 32.dp)
        ) {
            // Header
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp, vertical = 14.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                MimAvatar(name = conv.title, size = 42.dp, isGroup = isGroup)
                Spacer(modifier = Modifier.width(12.dp))
                Column {
                    Text(
                        text = conv.title,
                        style = MaterialTheme.typography.titleMedium.copy(
                            fontWeight = FontWeight.Bold,
                            color = MimTheme.colors.textPrimary
                        )
                    )
                    Text(
                        text = if (isGroup) "群聊会话" else "私聊会话",
                        style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.textMuted)
                    )
                }
            }

            HorizontalDivider(color = MimTheme.colors.border)

            // 1. Pin / Unpin
            ActionRowItem(
                icon = Icons.Default.PushPin,
                label = if (conv.isPinned) "取消置顶" else "置顶会话",
                iconTint = MimTheme.colors.accent,
                onClick = {
                    scope.launch {
                        val res = chatRepo.pinConversation(conv.id, !conv.isPinned)
                        onDismiss()
                        if (res.isSuccess) {
                            Toast.makeText(context, if (!conv.isPinned) "已置顶" else "已取消置顶", Toast.LENGTH_SHORT).show()
                            onActionCompleted()
                        } else {
                            Toast.makeText(context, "操作失败", Toast.LENGTH_SHORT).show()
                        }
                    }
                }
            )

            // 2. Mute / Unmute
            ActionRowItem(
                icon = if (conv.isMuted) Icons.Default.Notifications else Icons.Default.NotificationsOff,
                label = if (conv.isMuted) "开启新消息通知" else "消息免打扰",
                iconTint = MimTheme.colors.accent,
                onClick = {
                    scope.launch {
                        val res = chatRepo.muteConversation(conv.id, !conv.isMuted)
                        onDismiss()
                        if (res.isSuccess) {
                            Toast.makeText(context, if (!conv.isMuted) "已开启免打扰" else "已恢复通知", Toast.LENGTH_SHORT).show()
                            onActionCompleted()
                        } else {
                            Toast.makeText(context, "操作失败", Toast.LENGTH_SHORT).show()
                        }
                    }
                }
            )

            // Clear Chat History
            ActionRowItem(
                icon = Icons.Default.CleaningServices,
                label = "清空聊天记录",
                iconTint = MimTheme.colors.textSecondary,
                onClick = {
                    scope.launch {
                        val res = chatRepo.clearMessages(conv.id)
                        onDismiss()
                        if (res.isSuccess) {
                            Toast.makeText(context, "聊天记录已清空", Toast.LENGTH_SHORT).show()
                            onActionCompleted()
                        } else {
                            Toast.makeText(context, "清空失败", Toast.LENGTH_SHORT).show()
                        }
                    }
                }
            )

            // Voice & Video Call (only for single chats)
            if (!isGroup) {
                ActionRowItem(
                    icon = Icons.Default.Phone,
                    label = "语音通话",
                    iconTint = Color(0xFF16A34A),
                    onClick = {
                        val targetId = conv.targetId.toLongOrNull() ?: 0L
                        if (targetId > 0L) {
                            onDismiss()
                            chatRepo.initiateCall(targetId, conv.title, false)
                        }
                    }
                )

                ActionRowItem(
                    icon = Icons.Default.Videocam,
                    label = "视频通话",
                    iconTint = MimTheme.colors.accent,
                    onClick = {
                        val targetId = conv.targetId.toLongOrNull() ?: 0L
                        if (targetId > 0L) {
                            onDismiss()
                            chatRepo.initiateCall(targetId, conv.title, true)
                        }
                    }
                )
            }

            // 4. Delete Conversation
            ActionRowItem(
                icon = Icons.Default.Delete,
                label = "删除会话",
                iconTint = Color(0xFFDC2626),
                labelColor = Color(0xFFDC2626),
                onClick = {
                    scope.launch {
                        val res = chatRepo.deleteConversation(conv.id)
                        onDismiss()
                        if (res.isSuccess) {
                            Toast.makeText(context, "会话已删除", Toast.LENGTH_SHORT).show()
                            onActionCompleted()
                        } else {
                            Toast.makeText(context, "删除失败", Toast.LENGTH_SHORT).show()
                        }
                    }
                }
            )

            // 5. Delete Friend (only for single chats)
            if (!isGroup) {
                ActionRowItem(
                    icon = Icons.Default.PersonRemove,
                    label = "删除联系人",
                    iconTint = Color(0xFFDC2626),
                    labelColor = Color(0xFFDC2626),
                    onClick = {
                        val friendId = conv.targetId.toLongOrNull() ?: 0L
                        if (friendId > 0L) {
                            scope.launch {
                                val res = authRepo.deleteFriend(friendId)
                                if (res.isSuccess) {
                                    chatRepo.deleteConversation(conv.id)
                                    onDismiss()
                                    Toast.makeText(context, "已删除联系人", Toast.LENGTH_SHORT).show()
                                    onActionCompleted()
                                } else {
                                    Toast.makeText(context, "删除失败", Toast.LENGTH_SHORT).show()
                                }
                            }
                        }
                    }
                )
            }
        }
    }
}

@Composable
private fun ActionRowItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    iconTint: Color,
    labelColor: Color = MimTheme.colors.textPrimary,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 20.dp, vertical = 14.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(imageVector = icon, contentDescription = null, tint = iconTint, modifier = Modifier.size(22.dp))
        Spacer(modifier = Modifier.width(16.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.bodyLarge.copy(
                fontWeight = FontWeight.Medium,
                color = labelColor
            )
        )
    }
}
