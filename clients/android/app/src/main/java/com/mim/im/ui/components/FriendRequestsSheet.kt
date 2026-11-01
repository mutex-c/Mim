package com.mim.im.ui.components

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
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
import com.mim.im.data.model.FriendRequestDto
import com.mim.im.repository.AuthRepository
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun FriendRequestsSheet(
    authRepo: AuthRepository,
    onDismiss: () -> Unit,
    onRequestHandled: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    var requests by remember { mutableStateOf<List<FriendRequestDto>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }

    fun refreshList() {
        scope.launch {
            isLoading = true
            val res = authRepo.getFriendRequests()
            isLoading = false
            if (res.isSuccess) {
                requests = res.getOrDefault(emptyList())
            }
        }
    }

    val sheetState = rememberModalBottomSheetState(skipPartiallyExpanded = true)

    LaunchedEffect(Unit) {
        refreshList()
    }

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
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp, vertical = 12.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "新的朋友",
                    style = MaterialTheme.typography.titleLarge.copy(
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.textPrimary
                    )
                )
                IconButton(onClick = onDismiss) {
                    Icon(
                        imageVector = Icons.Default.Close,
                        contentDescription = "关闭",
                        tint = MimTheme.colors.textMuted
                    )
                }
            }

            HorizontalDivider(color = MimTheme.colors.border)

            if (isLoading) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(180.dp),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(color = MimTheme.colors.accent, strokeWidth = 2.dp)
                }
            } else if (requests.isEmpty()) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(180.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "暂无好友申请",
                        style = MaterialTheme.typography.bodyMedium.copy(
                            color = MimTheme.colors.textMuted
                        )
                    )
                }
            } else {
                LazyColumn(
                    modifier = Modifier
                        .fillMaxWidth()
                        .heightIn(max = 420.dp)
                ) {
                    items(requests) { req ->
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 20.dp, vertical = 12.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            MimAvatar(
                                name = req.effectiveNickname,
                                size = 44.dp
                            )
                            Spacer(modifier = Modifier.width(12.dp))

                            Column(modifier = Modifier.weight(1f)) {
                                Text(
                                    text = req.effectiveNickname,
                                    style = MaterialTheme.typography.bodyLarge.copy(
                                        fontWeight = FontWeight.SemiBold,
                                        color = MimTheme.colors.textPrimary
                                    )
                                )
                                Text(
                                    text = if (req.effectiveMessage.isNotBlank()) req.effectiveMessage else "请求添加您为好友",
                                    style = MaterialTheme.typography.bodySmall.copy(
                                        color = MimTheme.colors.textSecondary
                                    ),
                                    maxLines = 1
                                )
                            }

                            Spacer(modifier = Modifier.width(8.dp))

                            when (req.status) {
                                1 -> { // Pending
                                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                                        OutlinedButton(
                                            onClick = {
                                                scope.launch {
                                                    val res = authRepo.handleFriendRequest(req.id, 3) // Reject
                                                    if (res.isSuccess) {
                                                        Toast.makeText(context, "已拒绝", Toast.LENGTH_SHORT).show()
                                                        refreshList()
                                                        onRequestHandled()
                                                    } else {
                                                        Toast.makeText(context, res.exceptionOrNull()?.message ?: "操作失败", Toast.LENGTH_SHORT).show()
                                                    }
                                                }
                                            },
                                            shape = RoundedCornerShape(8.dp),
                                            contentPadding = PaddingValues(horizontal = 10.dp, vertical = 4.dp),
                                            modifier = Modifier.height(32.dp),
                                            border = androidx.compose.foundation.BorderStroke(0.5.dp, MimTheme.colors.border)
                                        ) {
                                            Text("拒绝", fontSize = 12.sp, color = MimTheme.colors.textMuted)
                                        }

                                        Button(
                                            onClick = {
                                                scope.launch {
                                                    val res = authRepo.handleFriendRequest(req.id, 2) // Accept
                                                    if (res.isSuccess) {
                                                        Toast.makeText(context, "已添加为好友", Toast.LENGTH_SHORT).show()
                                                        refreshList()
                                                        onRequestHandled()
                                                    } else {
                                                        Toast.makeText(context, res.exceptionOrNull()?.message ?: "操作失败", Toast.LENGTH_SHORT).show()
                                                    }
                                                }
                                            },
                                            shape = RoundedCornerShape(8.dp),
                                            contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                                            modifier = Modifier.height(32.dp),
                                            colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent)
                                        ) {
                                            Text("同意", fontSize = 12.sp, color = Color.White)
                                        }
                                    }
                                }
                                2 -> {
                                    Text(
                                        text = "已同意",
                                        style = MaterialTheme.typography.bodySmall.copy(
                                            color = MimTheme.colors.textMuted
                                        )
                                    )
                                }
                                3 -> {
                                    Text(
                                        text = "已拒绝",
                                        style = MaterialTheme.typography.bodySmall.copy(
                                            color = MimTheme.colors.textMuted
                                        )
                                    )
                                }
                            }
                        }
                        HorizontalDivider(
                            modifier = Modifier.padding(start = 76.dp),
                            color = MimTheme.colors.border.copy(alpha = 0.5f)
                        )
                    }
                }
            }
        }
    }
}
