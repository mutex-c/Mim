package com.mim.im.ui.components

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
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
import com.mim.im.data.model.GroupInfoDto
import com.mim.im.data.model.UserDto
import com.mim.im.repository.AuthRepository
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch

@Composable
fun NewGroupDialog(
    authRepo: AuthRepository,
    chatRepo: ChatRepository,
    onDismiss: () -> Unit,
    onGroupCreated: (groupId: Long, title: String) -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()

    var friends by remember { mutableStateOf<List<UserDto>>(emptyList()) }
    var searchQuery by remember { mutableStateOf("") }
    var groupName by remember { mutableStateOf("") }
    val selectedUserIds = remember { mutableStateListOf<Long>() }
    var isCreating by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        val res = authRepo.getFriends()
        if (res.isSuccess) {
            friends = res.getOrDefault(emptyList())
        }
    }

    val filteredFriends = remember(friends, searchQuery) {
        if (searchQuery.isBlank()) friends
        else friends.filter { it.nickname.contains(searchQuery, true) || it.account.contains(searchQuery, true) }
    }

    Dialog(onDismissRequest = onDismiss) {
        Surface(
            shape = RoundedCornerShape(18.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .heightIn(max = 580.dp)
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(18.dp))
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp)
            ) {
                // Header
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Text(
                        text = "发起群聊",
                        style = MaterialTheme.typography.titleLarge.copy(
                            fontWeight = FontWeight.Bold,
                            color = MimTheme.colors.textPrimary
                        )
                    )
                    IconButton(onClick = onDismiss, modifier = Modifier.size(32.dp)) {
                        Icon(Icons.Default.Close, contentDescription = "关闭", tint = MimTheme.colors.textMuted)
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                // Group Name Input
                OutlinedTextField(
                    value = groupName,
                    onValueChange = { groupName = it },
                    placeholder = { Text("群聊名称 (选填，默认按成员命名)", color = MimTheme.colors.textMuted, fontSize = 13.sp) },
                    shape = RoundedCornerShape(12.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = MimTheme.colors.accent,
                        unfocusedBorderColor = MimTheme.colors.border,
                        focusedContainerColor = MimTheme.colors.inputBackground,
                        unfocusedContainerColor = MimTheme.colors.inputBackground,
                        focusedTextColor = MimTheme.colors.textPrimary,
                        unfocusedTextColor = MimTheme.colors.textPrimary
                    ),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )

                Spacer(modifier = Modifier.height(10.dp))

                // Selected contacts shuttle tray
                if (selectedUserIds.isNotEmpty()) {
                    LazyRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(vertical = 4.dp)
                    ) {
                        items(selectedUserIds, key = { it }) { uid ->
                            val user = friends.find { it.id == uid }
                            Box(
                                modifier = Modifier
                                    .size(38.dp)
                                    .clip(CircleShape)
                                    .background(MimTheme.colors.surface)
                                    .clickable { selectedUserIds.remove(uid) },
                                contentAlignment = Alignment.Center
                            ) {
                                MimAvatar(name = user?.nickname ?: uid.toString(), size = 38.dp)
                            }
                        }
                    }
                    Spacer(modifier = Modifier.height(8.dp))
                }

                // Search contacts
                OutlinedTextField(
                    value = searchQuery,
                    onValueChange = { searchQuery = it },
                    leadingIcon = { Icon(Icons.Default.Search, contentDescription = null, tint = MimTheme.colors.textMuted) },
                    placeholder = { Text("搜索好友", color = MimTheme.colors.textMuted, fontSize = 13.sp) },
                    shape = RoundedCornerShape(10.dp),
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = MimTheme.colors.accent,
                        unfocusedBorderColor = MimTheme.colors.border,
                        focusedContainerColor = MimTheme.colors.inputBackground,
                        unfocusedContainerColor = MimTheme.colors.inputBackground,
                        focusedTextColor = MimTheme.colors.textPrimary,
                        unfocusedTextColor = MimTheme.colors.textPrimary
                    ),
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )

                Spacer(modifier = Modifier.height(10.dp))

                // Friends List with Checkbox
                LazyColumn(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxWidth()
                ) {
                    items(filteredFriends, key = { it.id }) { friend ->
                        val isSelected = selectedUserIds.contains(friend.id)
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .clip(RoundedCornerShape(8.dp))
                                .clickable {
                                    if (isSelected) selectedUserIds.remove(friend.id)
                                    else selectedUserIds.add(friend.id)
                                }
                                .padding(vertical = 8.dp, horizontal = 4.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            // Checkmark circle
                            Box(
                                modifier = Modifier
                                    .size(20.dp)
                                    .clip(CircleShape)
                                    .border(
                                        width = 1.5.dp,
                                        color = if (isSelected) MimTheme.colors.accent else MimTheme.colors.border,
                                        shape = CircleShape
                                    )
                                    .background(if (isSelected) MimTheme.colors.accent else Color.Transparent),
                                contentAlignment = Alignment.Center
                            ) {
                                if (isSelected) {
                                    Icon(
                                        imageVector = Icons.Default.Check,
                                        contentDescription = null,
                                        tint = Color.White,
                                        modifier = Modifier.size(14.dp)
                                    )
                                }
                            }

                            Spacer(modifier = Modifier.width(12.dp))

                            MimAvatar(name = friend.nickname.ifBlank { friend.account }, size = 36.dp)

                            Spacer(modifier = Modifier.width(10.dp))

                            Column(modifier = Modifier.weight(1f)) {
                                Text(
                                    text = friend.nickname.ifBlank { friend.account },
                                    style = MaterialTheme.typography.bodyMedium.copy(
                                        fontWeight = FontWeight.Medium,
                                        color = MimTheme.colors.textPrimary
                                    )
                                )
                                Text(
                                    text = "@${friend.account}",
                                    style = MaterialTheme.typography.bodySmall.copy(
                                        color = MimTheme.colors.textMuted
                                    )
                                )
                            }
                        }
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                // Bottom Action Button
                Button(
                    onClick = {
                        if (selectedUserIds.isEmpty()) {
                            Toast.makeText(context, "请至少选择一位群成员", Toast.LENGTH_SHORT).show()
                            return@Button
                        }
                        isCreating = true
                        scope.launch {
                            val defaultName = if (groupName.isNotBlank()) {
                                groupName.trim()
                            } else {
                                val names = friends.filter { selectedUserIds.contains(it.id) }.map { it.nickname.ifBlank { it.account } }
                                (names.take(3).joinToString("、") + "的群聊").take(30)
                            }

                            val res = chatRepo.createGroup(defaultName, selectedUserIds.toList())
                            isCreating = false
                            if (res.isSuccess) {
                                val group = res.getOrNull()!!
                                Toast.makeText(context, "群聊创建成功", Toast.LENGTH_SHORT).show()
                                onGroupCreated(group.id, group.title)
                            } else {
                                Toast.makeText(context, res.exceptionOrNull()?.message ?: "创建失败", Toast.LENGTH_SHORT).show()
                            }
                        }
                    },
                    enabled = !isCreating && selectedUserIds.isNotEmpty(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(44.dp)
                ) {
                    if (isCreating) {
                        CircularProgressIndicator(modifier = Modifier.size(18.dp), color = Color.White, strokeWidth = 2.dp)
                    } else {
                        Text(
                            text = if (selectedUserIds.isEmpty()) "请选择成员" else "完成创建 (${selectedUserIds.size})",
                            fontWeight = FontWeight.Bold,
                            color = Color.White
                        )
                    }
                }
            }
        }
    }
}
