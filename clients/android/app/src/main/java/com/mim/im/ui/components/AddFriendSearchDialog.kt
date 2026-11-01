package com.mim.im.ui.components

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Search
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
fun AddFriendSearchDialog(
    authRepo: AuthRepository,
    myUserId: String,
    friendIds: Set<Long>,
    onDismiss: () -> Unit,
    onFriendAddedOrRequested: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()

    var searchKeyword by remember { mutableStateOf("") }
    var searchResults by remember { mutableStateOf<List<UserDto>>(emptyList()) }
    var isSearching by remember { mutableStateOf(false) }
    var hasSearched by remember { mutableStateOf(false) }

    var selectedUserForReq by remember { mutableStateOf<UserDto?>(null) }
    var verifyMessage by remember { mutableStateOf("你好，我想添加你为好友") }
    var isSendingReq by remember { mutableStateOf(false) }

    fun doSearch() {
        val kw = searchKeyword.trim()
        if (kw.isBlank()) {
            Toast.makeText(context, "请输入搜索关键词", Toast.LENGTH_SHORT).show()
            return
        }
        isSearching = true
        hasSearched = true
        scope.launch {
            val res = authRepo.searchUsers(kw)
            isSearching = false
            if (res.isSuccess) {
                searchResults = res.getOrDefault(emptyList())
            } else {
                searchResults = emptyList()
                Toast.makeText(context, "搜索失败，请重试", Toast.LENGTH_SHORT).show()
            }
        }
    }

    Dialog(onDismissRequest = onDismiss) {
        Surface(
            shape = RoundedCornerShape(18.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .heightIn(max = 560.dp)
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(18.dp))
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(20.dp)
            ) {
                // Header
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = "添加好友",
                        style = MaterialTheme.typography.titleLarge.copy(
                            fontWeight = FontWeight.Bold,
                            color = MimTheme.colors.textPrimary
                        )
                    )
                    IconButton(onClick = onDismiss, modifier = Modifier.size(32.dp)) {
                        Icon(Icons.Default.Close, contentDescription = "关闭", tint = MimTheme.colors.textMuted)
                    }
                }

                Spacer(modifier = Modifier.height(14.dp))

                // Search input + Search button
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    OutlinedTextField(
                        value = searchKeyword,
                        onValueChange = { searchKeyword = it },
                        placeholder = { Text("输入用户名或账号搜索...", color = MimTheme.colors.textMuted, fontSize = 13.sp) },
                        leadingIcon = { Icon(Icons.Default.Search, contentDescription = null, tint = MimTheme.colors.accent) },
                        shape = RoundedCornerShape(12.dp),
                        singleLine = true,
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = MimTheme.colors.accent,
                            unfocusedBorderColor = MimTheme.colors.border,
                            focusedContainerColor = MimTheme.colors.inputBackground,
                            unfocusedContainerColor = MimTheme.colors.inputBackground,
                            focusedTextColor = MimTheme.colors.textPrimary,
                            unfocusedTextColor = MimTheme.colors.textPrimary
                        ),
                        modifier = Modifier.weight(1f)
                    )

                    Button(
                        onClick = { doSearch() },
                        enabled = !isSearching && searchKeyword.isNotBlank(),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                        modifier = Modifier.height(50.dp)
                    ) {
                        if (isSearching) {
                            CircularProgressIndicator(modifier = Modifier.size(16.dp), color = Color.White, strokeWidth = 2.dp)
                        } else {
                            Text("搜索", color = Color.White, fontWeight = FontWeight.Bold)
                        }
                    }
                }

                Spacer(modifier = Modifier.height(14.dp))

                // If user selected someone to add
                if (selectedUserForReq != null) {
                    val target = selectedUserForReq!!
                    Surface(
                        shape = RoundedCornerShape(12.dp),
                        color = MimTheme.colors.background,
                        modifier = Modifier
                            .fillMaxWidth()
                            .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
                            .padding(12.dp)
                    ) {
                        Column {
                            Row(verticalAlignment = Alignment.CenterVertically) {
                                MimAvatar(name = target.displayName, size = 40.dp)
                                Spacer(modifier = Modifier.width(10.dp))
                                Column(modifier = Modifier.weight(1f)) {
                                    Text(
                                        text = target.displayName,
                                        style = MaterialTheme.typography.bodyMedium.copy(
                                            fontWeight = FontWeight.Bold,
                                            color = MimTheme.colors.textPrimary
                                        )
                                    )
                                    Text(
                                        text = "@${target.account}",
                                        style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.textMuted)
                                    )
                                }
                                IconButton(
                                    onClick = { selectedUserForReq = null },
                                    modifier = Modifier.size(24.dp)
                                ) {
                                    Icon(Icons.Default.Close, contentDescription = "取消选择", tint = MimTheme.colors.textMuted)
                                }
                            }

                            Spacer(modifier = Modifier.height(10.dp))

                            OutlinedTextField(
                                value = verifyMessage,
                                onValueChange = { verifyMessage = it },
                                placeholder = { Text("输入申请附言...", color = MimTheme.colors.textMuted, fontSize = 12.sp) },
                                shape = RoundedCornerShape(8.dp),
                                maxLines = 2,
                                colors = OutlinedTextFieldDefaults.colors(
                                    focusedBorderColor = MimTheme.colors.accent,
                                    unfocusedBorderColor = MimTheme.colors.border,
                                    focusedContainerColor = MimTheme.colors.surfaceElevated,
                                    unfocusedContainerColor = MimTheme.colors.surfaceElevated,
                                    focusedTextColor = MimTheme.colors.textPrimary,
                                    unfocusedTextColor = MimTheme.colors.textPrimary
                                ),
                                modifier = Modifier.fillMaxWidth()
                            )

                            Spacer(modifier = Modifier.height(10.dp))

                            Button(
                                onClick = {
                                    if (!isSendingReq) {
                                        isSendingReq = true
                                        scope.launch {
                                            val res = authRepo.sendFriendRequest(target.effectiveId, verifyMessage.trim())
                                            isSendingReq = false
                                            if (res.isSuccess) {
                                                Toast.makeText(context, res.getOrNull() ?: "申请已发送", Toast.LENGTH_SHORT).show()
                                                selectedUserForReq = null
                                                onFriendAddedOrRequested()
                                            } else {
                                                Toast.makeText(context, res.exceptionOrNull()?.message ?: "发送失败", Toast.LENGTH_SHORT).show()
                                            }
                                        }
                                    }
                                },
                                enabled = !isSendingReq,
                                shape = RoundedCornerShape(8.dp),
                                colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                                modifier = Modifier.fillMaxWidth().height(38.dp)
                            ) {
                                if (isSendingReq) {
                                    CircularProgressIndicator(modifier = Modifier.size(16.dp), color = Color.White, strokeWidth = 2.dp)
                                } else {
                                    Text("发送申请", color = Color.White, fontWeight = FontWeight.SemiBold)
                                }
                            }
                        }
                    }
                } else {
                    // Search results list
                    if (isSearching) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(160.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            CircularProgressIndicator(color = MimTheme.colors.accent, strokeWidth = 2.dp)
                        }
                    } else if (searchResults.isEmpty()) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(140.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = if (hasSearched) "未找到匹配用户，请确认账号拼写" else "搜索好友账号或昵称发起添加",
                                style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.textMuted)
                            )
                        }
                    } else {
                        LazyColumn(
                            verticalArrangement = Arrangement.spacedBy(8.dp),
                            modifier = Modifier
                                .fillMaxWidth()
                                .heightIn(max = 300.dp)
                        ) {
                            items(searchResults) { user ->
                                val isSelf = user.effectiveId.toString() == myUserId
                                val isAlreadyFriend = friendIds.contains(user.effectiveId)

                                Surface(
                                    shape = RoundedCornerShape(10.dp),
                                    color = MimTheme.colors.background,
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(10.dp))
                                        .padding(10.dp)
                                ) {
                                    Row(
                                        modifier = Modifier.fillMaxWidth(),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        MimAvatar(name = user.displayName, size = 38.dp)
                                        Spacer(modifier = Modifier.width(10.dp))
                                        Column(modifier = Modifier.weight(1f)) {
                                            Text(
                                                text = user.displayName,
                                                style = MaterialTheme.typography.bodyMedium.copy(
                                                    fontWeight = FontWeight.Bold,
                                                    color = MimTheme.colors.textPrimary
                                                )
                                            )
                                            Text(
                                                text = "@${user.account}",
                                                style = MaterialTheme.typography.bodySmall.copy(
                                                    color = MimTheme.colors.textMuted,
                                                    fontSize = 11.sp
                                                )
                                            )
                                        }

                                        when {
                                            isSelf -> {
                                                Surface(
                                                    shape = RoundedCornerShape(6.dp),
                                                    color = MimTheme.colors.surfaceElevated,
                                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                                                ) {
                                                    Text("自己", fontSize = 11.sp, color = MimTheme.colors.textMuted)
                                                }
                                            }
                                            isAlreadyFriend -> {
                                                Surface(
                                                    shape = RoundedCornerShape(6.dp),
                                                    color = Color(0xFF22C55E).copy(alpha = 0.12f),
                                                    modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                                                ) {
                                                    Text("已是好友", fontSize = 11.sp, color = Color(0xFF22C55E), fontWeight = FontWeight.SemiBold)
                                                }
                                            }
                                            else -> {
                                                Button(
                                                    onClick = { selectedUserForReq = user },
                                                    shape = RoundedCornerShape(6.dp),
                                                    contentPadding = PaddingValues(horizontal = 12.dp, vertical = 4.dp),
                                                    colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                                                    modifier = Modifier.height(30.dp)
                                                ) {
                                                    Text("+ 好友", fontSize = 11.sp, color = Color.White)
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
