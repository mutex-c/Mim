package com.mim.im.ui.components

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
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
import androidx.compose.ui.window.Dialog
import com.mim.im.data.model.UserDto
import com.mim.im.repository.AuthRepository
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch

@Composable
fun AddFriendDialog(
    targetUser: UserDto,
    authRepo: AuthRepository,
    onDismiss: () -> Unit,
    onSuccess: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var message by remember { mutableStateOf("你好，我想添加你为好友") }
    var isSending by remember { mutableStateOf(false) }

    Dialog(onDismissRequest = onDismiss) {
        Surface(
            shape = RoundedCornerShape(16.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(16.dp))
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(20.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                MimAvatar(name = targetUser.nickname.ifBlank { targetUser.account }, size = 52.dp)
                Spacer(modifier = Modifier.height(10.dp))
                Text(
                    text = targetUser.nickname.ifBlank { targetUser.account },
                    style = MaterialTheme.typography.titleMedium.copy(
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.textPrimary
                    )
                )
                Text(
                    text = "@${targetUser.account}",
                    style = MaterialTheme.typography.bodySmall.copy(
                        color = MimTheme.colors.textMuted
                    )
                )

                Spacer(modifier = Modifier.height(16.dp))

                OutlinedTextField(
                    value = message,
                    onValueChange = { message = it },
                    label = { Text("验证消息", color = MimTheme.colors.textMuted) },
                    placeholder = { Text("输入打招呼的话...", color = MimTheme.colors.textMuted) },
                    shape = RoundedCornerShape(12.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = MimTheme.colors.accent,
                        unfocusedBorderColor = MimTheme.colors.border,
                        focusedContainerColor = MimTheme.colors.inputBackground,
                        unfocusedContainerColor = MimTheme.colors.inputBackground,
                        focusedTextColor = MimTheme.colors.textPrimary,
                        unfocusedTextColor = MimTheme.colors.textPrimary
                    ),
                    maxLines = 3,
                    modifier = Modifier.fillMaxWidth()
                )

                Spacer(modifier = Modifier.height(20.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    OutlinedButton(
                        onClick = onDismiss,
                        shape = RoundedCornerShape(10.dp),
                        modifier = Modifier.weight(1f),
                        border = androidx.compose.foundation.BorderStroke(1.dp, MimTheme.colors.border)
                    ) {
                        Text("取消", color = MimTheme.colors.textSecondary)
                    }

                    Button(
                        onClick = {
                            if (!isSending) {
                                isSending = true
                                scope.launch {
                                    val res = authRepo.sendFriendRequest(targetUser.id, message.trim())
                                    isSending = false
                                    if (res.isSuccess) {
                                        Toast.makeText(context, res.getOrNull() ?: "好友申请已发送", Toast.LENGTH_SHORT).show()
                                        onSuccess()
                                    } else {
                                        Toast.makeText(context, res.exceptionOrNull()?.message ?: "发送失败", Toast.LENGTH_SHORT).show()
                                    }
                                }
                            }
                        },
                        enabled = !isSending,
                        shape = RoundedCornerShape(10.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = MimTheme.colors.accent
                        ),
                        modifier = Modifier.weight(1f)
                    ) {
                        if (isSending) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(16.dp),
                                color = Color.White,
                                strokeWidth = 2.dp
                            )
                        } else {
                            Text("发送申请", color = Color.White)
                        }
                    }
                }
            }
        }
    }
}
