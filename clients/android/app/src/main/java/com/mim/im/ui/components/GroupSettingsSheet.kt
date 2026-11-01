package com.mim.im.ui.components

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
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
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.mim.im.data.model.GroupInfoDto
import com.mim.im.data.model.GroupMemberDto
import com.mim.im.data.model.UserDto
import com.mim.im.repository.AuthRepository
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GroupSettingsSheet(
    groupId: Long,
    currentOwnerId: String,
    chatRepo: ChatRepository,
    authRepo: AuthRepository,
    onDismiss: () -> Unit,
    onGroupDisbandedOrQuit: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()

    var groupInfo by remember { mutableStateOf<GroupInfoDto?>(null) }
    var members by remember { mutableStateOf<List<GroupMemberDto>>(emptyList()) }
    var isLoading by remember { mutableStateOf(true) }

    // Dialog states
    var selectedMemberForAction by remember { mutableStateOf<GroupMemberDto?>(null) }
    var memberToMute by remember { mutableStateOf<GroupMemberDto?>(null) }
    var showInviteDialog by remember { mutableStateOf(false) }
    var showEditNoticeDialog by remember { mutableStateOf(false) }
    var showDisbandConfirm by remember { mutableStateOf(false) }
    var showQuitConfirm by remember { mutableStateOf(false) }

    fun refreshData() {
        scope.launch {
            isLoading = true
            val infoRes = chatRepo.getGroupInfo(groupId)
            val membersRes = chatRepo.getGroupMembers(groupId)
            isLoading = false
            if (infoRes.isSuccess) groupInfo = infoRes.getOrNull()
            if (membersRes.isSuccess) members = membersRes.getOrDefault(emptyList())
        }
    }

    LaunchedEffect(groupId) {
        refreshData()
    }

    val myMember = members.find { it.userId.toString() == currentOwnerId }
    val myRole = myMember?.role ?: 3
    val isOwner = myRole == 1
    val isAdmin = myRole == 2
    val isOwnerOrAdmin = isOwner || isAdmin
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
                .verticalScroll(rememberScrollState())
                .padding(bottom = 32.dp)
        ) {
            // Top Bar
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Text(
                    text = "群聊设置",
                    style = MaterialTheme.typography.titleLarge.copy(
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.textPrimary
                    )
                )
                IconButton(onClick = onDismiss) {
                    Icon(Icons.Default.Close, contentDescription = "关闭", tint = MimTheme.colors.textMuted)
                }
            }

            HorizontalDivider(color = MimTheme.colors.border)

            if (isLoading && groupInfo == null) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(200.dp),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator(color = MimTheme.colors.accent, strokeWidth = 2.dp)
                }
            } else {
                val info = groupInfo ?: GroupInfoDto(id = groupId, title = "群聊 $groupId")

                // Group Profile Header
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 20.dp, vertical = 16.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    MimAvatar(name = info.title, size = 52.dp, isGroup = true)
                    Spacer(modifier = Modifier.width(14.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = info.title,
                            style = MaterialTheme.typography.titleMedium.copy(
                                fontWeight = FontWeight.Bold,
                                color = MimTheme.colors.textPrimary
                            )
                        )
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = "群号: ${info.id} · ${members.size} 位成员",
                            style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.textMuted)
                        )
                    }
                }

                HorizontalDivider(color = MimTheme.colors.border.copy(alpha = 0.5f))

                // Group Announcement / Notice Card
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 20.dp, vertical = 12.dp)
                ) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Text(
                            text = "群公告",
                            style = MaterialTheme.typography.bodyMedium.copy(
                                fontWeight = FontWeight.SemiBold,
                                color = MimTheme.colors.textPrimary
                            )
                        )
                        if (isOwnerOrAdmin) {
                            TextButton(
                                onClick = { showEditNoticeDialog = true },
                                contentPadding = PaddingValues(horizontal = 8.dp, vertical = 0.dp)
                            ) {
                                Text("编辑", color = MimTheme.colors.accent, fontSize = 12.sp)
                            }
                        }
                    }
                    Text(
                        text = info.notice.ifBlank { "暂无群公告" },
                        style = MaterialTheme.typography.bodySmall.copy(
                            color = if (info.notice.isNotBlank()) MimTheme.colors.textSecondary else MimTheme.colors.textMuted,
                            lineHeight = 18.sp
                        )
                    )
                }

                HorizontalDivider(color = MimTheme.colors.border.copy(alpha = 0.5f))

                // All Mute Switch (Owner only)
                if (isOwner) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 20.dp, vertical = 10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.SpaceBetween
                    ) {
                        Column {
                            Text(
                                text = "全员禁言",
                                style = MaterialTheme.typography.bodyMedium.copy(
                                    fontWeight = FontWeight.SemiBold,
                                    color = MimTheme.colors.textPrimary
                                )
                            )
                            Text(
                                text = "开启后仅群主和管理员可发言",
                                style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.textMuted)
                            )
                        }

                        Switch(
                            checked = info.isAllMuted == 1,
                            onCheckedChange = { checked ->
                                scope.launch {
                                    val res = chatRepo.updateGroup(
                                        groupId = groupId,
                                        title = info.title,
                                        notice = info.notice,
                                        isAllMuted = if (checked) 1 else 0
                                    )
                                    if (res.isSuccess) {
                                        refreshData()
                                    } else {
                                        Toast.makeText(context, "设置失败", Toast.LENGTH_SHORT).show()
                                    }
                                }
                            },
                            colors = SwitchDefaults.colors(
                                checkedThumbColor = Color.White,
                                checkedTrackColor = MimTheme.colors.accent
                            )
                        )
                    }
                    HorizontalDivider(color = MimTheme.colors.border.copy(alpha = 0.5f))
                }

                // Member Grid Section
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 12.dp)
                ) {
                    Text(
                        text = "群成员 (${members.size})",
                        style = MaterialTheme.typography.bodyMedium.copy(
                            fontWeight = FontWeight.SemiBold,
                            color = MimTheme.colors.textPrimary
                        ),
                        modifier = Modifier.padding(horizontal = 20.dp, vertical = 4.dp)
                    )

                    Spacer(modifier = Modifier.height(8.dp))

                    LazyRow(
                        contentPadding = PaddingValues(horizontal = 20.dp),
                        horizontalArrangement = Arrangement.spacedBy(14.dp),
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        // Invite Button
                        if (isOwnerOrAdmin) {
                            item {
                                Column(
                                    horizontalAlignment = Alignment.CenterHorizontally,
                                    modifier = Modifier
                                        .clip(RoundedCornerShape(8.dp))
                                        .clickable { showInviteDialog = true }
                                        .padding(4.dp)
                                ) {
                                    Box(
                                        modifier = Modifier
                                            .size(44.dp)
                                            .clip(CircleShape)
                                            .border(1.dp, MimTheme.colors.accent, CircleShape)
                                            .background(MimTheme.colors.accent.copy(alpha = 0.1f)),
                                        contentAlignment = Alignment.Center
                                    ) {
                                        Icon(Icons.Default.Add, contentDescription = "邀请", tint = MimTheme.colors.accent)
                                    }
                                    Spacer(modifier = Modifier.height(4.dp))
                                    Text(
                                        text = "邀请",
                                        style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.accent, fontSize = 11.sp)
                                    )
                                }
                            }
                        }

                        // Member items
                        items(members, key = { it.userId }) { member ->
                            val canManage = when {
                                isOwner && member.userId.toString() != currentOwnerId -> true
                                isAdmin && member.role == 3 -> true
                                else -> false
                            }

                            Column(
                                horizontalAlignment = Alignment.CenterHorizontally,
                                modifier = Modifier
                                    .clip(RoundedCornerShape(8.dp))
                                    .clickable(enabled = canManage) {
                                        selectedMemberForAction = member
                                    }
                                    .padding(4.dp)
                                    .width(52.dp)
                            ) {
                                Box {
                                    MimAvatar(name = member.nickname.ifBlank { member.account }, size = 44.dp)

                                    // Role badge
                                    when (member.role) {
                                        1 -> { // Owner
                                            Surface(
                                                shape = CircleShape,
                                                color = Color(0xFFF59E0B),
                                                modifier = Modifier
                                                    .size(16.dp)
                                                    .align(Alignment.BottomEnd)
                                            ) {
                                                Box(contentAlignment = Alignment.Center) {
                                                    Text("👑", fontSize = 9.sp)
                                                }
                                            }
                                        }
                                        2 -> { // Admin
                                            Surface(
                                                shape = CircleShape,
                                                color = MimTheme.colors.accent,
                                                modifier = Modifier
                                                    .size(16.dp)
                                                    .align(Alignment.BottomEnd)
                                            ) {
                                                Box(contentAlignment = Alignment.Center) {
                                                    Text("🛡️", fontSize = 9.sp)
                                                }
                                            }
                                        }
                                    }
                                }

                                Spacer(modifier = Modifier.height(4.dp))

                                Text(
                                    text = member.nickname.ifBlank { member.account },
                                    style = MaterialTheme.typography.bodySmall.copy(
                                        color = MimTheme.colors.textPrimary,
                                        fontSize = 11.sp
                                    ),
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis
                                )
                            }
                        }
                    }
                }

                HorizontalDivider(color = MimTheme.colors.border.copy(alpha = 0.5f))

                Spacer(modifier = Modifier.height(20.dp))

                // Danger zone action buttons
                Box(modifier = Modifier.padding(horizontal = 20.dp)) {
                    if (isOwner) {
                        Button(
                            onClick = { showDisbandConfirm = true },
                            colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFDC2626)),
                            shape = RoundedCornerShape(12.dp),
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(46.dp)
                        ) {
                            Text("解散群聊", color = Color.White, fontWeight = FontWeight.Bold)
                        }
                    } else {
                        Button(
                            onClick = { showQuitConfirm = true },
                            colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFDC2626).copy(alpha = 0.15f)),
                            shape = RoundedCornerShape(12.dp),
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(46.dp)
                        ) {
                            Text("退出群聊", color = Color(0xFFDC2626), fontWeight = FontWeight.Bold)
                        }
                    }
                }
            }
        }
    }

    // Member Action Sheet / Dialog
    selectedMemberForAction?.let { member ->
        AlertDialog(
            onDismissRequest = { selectedMemberForAction = null },
            title = { Text(member.nickname.ifBlank { member.account }, color = MimTheme.colors.textPrimary) },
            text = {
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    // Mute / Unmute option
                    TextButton(
                        onClick = {
                            selectedMemberForAction = null
                            memberToMute = member
                        },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text(if (member.isMuted) "解除禁言" else "设置禁言时长", color = MimTheme.colors.accent)
                    }

                    // Set/Remove Admin (Owner only)
                    if (isOwner) {
                        if (member.role == 3) {
                            TextButton(
                                onClick = {
                                    scope.launch {
                                        val res = chatRepo.setGroupRole(groupId, member.userId, 2)
                                        selectedMemberForAction = null
                                        if (res.isSuccess) {
                                            Toast.makeText(context, "已设为管理员", Toast.LENGTH_SHORT).show()
                                            refreshData()
                                        }
                                    }
                                },
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text("设为管理员", color = MimTheme.colors.accent)
                            }
                        } else if (member.role == 2) {
                            TextButton(
                                onClick = {
                                    scope.launch {
                                        val res = chatRepo.setGroupRole(groupId, member.userId, 3)
                                        selectedMemberForAction = null
                                        if (res.isSuccess) {
                                            Toast.makeText(context, "已取消管理员", Toast.LENGTH_SHORT).show()
                                            refreshData()
                                        }
                                    }
                                },
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Text("取消管理员", color = MimTheme.colors.textSecondary)
                            }
                        }

                        // Transfer Owner
                        TextButton(
                            onClick = {
                                scope.launch {
                                    val res = chatRepo.transferGroupOwner(groupId, member.userId)
                                    selectedMemberForAction = null
                                    if (res.isSuccess) {
                                        Toast.makeText(context, "群主已转让", Toast.LENGTH_SHORT).show()
                                        refreshData()
                                    }
                                }
                            },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text("转让群主", color = Color(0xFFF59E0B))
                        }
                    }

                    // Kick member
                    TextButton(
                        onClick = {
                            scope.launch {
                                val res = chatRepo.kickMember(groupId, member.userId)
                                selectedMemberForAction = null
                                if (res.isSuccess) {
                                    Toast.makeText(context, "已移出群聊", Toast.LENGTH_SHORT).show()
                                    refreshData()
                                }
                            }
                        },
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text("移出群聊", color = Color(0xFFDC2626))
                    }
                }
            },
            confirmButton = {},
            dismissButton = {
                TextButton(onClick = { selectedMemberForAction = null }) {
                    Text("取消", color = MimTheme.colors.textMuted)
                }
            },
            containerColor = MimTheme.colors.surfaceElevated
        )
    }

    // Mute Duration Dialog
    memberToMute?.let { member ->
        MuteDurationDialog(
            memberName = member.nickname.ifBlank { member.account },
            onDismiss = { memberToMute = null },
            onSelectDuration = { secs ->
                scope.launch {
                    val res = chatRepo.muteMember(groupId, member.userId, secs)
                    memberToMute = null
                    if (res.isSuccess) {
                        Toast.makeText(context, if (secs == 0) "已解除禁言" else "已禁言", Toast.LENGTH_SHORT).show()
                        refreshData()
                    } else {
                        Toast.makeText(context, "操作失败", Toast.LENGTH_SHORT).show()
                    }
                }
            }
        )
    }

    // Disband Confirm Dialog
    if (showDisbandConfirm) {
        AlertDialog(
            onDismissRequest = { showDisbandConfirm = false },
            title = { Text("确认解散群聊？", color = MimTheme.colors.textPrimary) },
            text = { Text("解散后群成员将被清空且无法恢复。", color = MimTheme.colors.textSecondary) },
            confirmButton = {
                Button(
                    onClick = {
                        scope.launch {
                            val res = chatRepo.disbandGroup(groupId)
                            showDisbandConfirm = false
                            if (res.isSuccess) {
                                Toast.makeText(context, "群聊已解散", Toast.LENGTH_SHORT).show()
                                onDismiss()
                                onGroupDisbandedOrQuit()
                            } else {
                                Toast.makeText(context, "解散失败", Toast.LENGTH_SHORT).show()
                            }
                        }
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFDC2626))
                ) {
                    Text("确认解散", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { showDisbandConfirm = false }) {
                    Text("取消", color = MimTheme.colors.textMuted)
                }
            },
            containerColor = MimTheme.colors.surfaceElevated
        )
    }

    // Quit Confirm Dialog
    if (showQuitConfirm) {
        AlertDialog(
            onDismissRequest = { showQuitConfirm = false },
            title = { Text("退出群聊？", color = MimTheme.colors.textPrimary) },
            text = { Text("退出后将不再接收该群的消息通知。", color = MimTheme.colors.textSecondary) },
            confirmButton = {
                Button(
                    onClick = {
                        scope.launch {
                            val res = chatRepo.quitGroup(groupId)
                            showQuitConfirm = false
                            if (res.isSuccess) {
                                Toast.makeText(context, "已退出群聊", Toast.LENGTH_SHORT).show()
                                onDismiss()
                                onGroupDisbandedOrQuit()
                            } else {
                                Toast.makeText(context, "退出失败", Toast.LENGTH_SHORT).show()
                            }
                        }
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFDC2626))
                ) {
                    Text("确认退出", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { showQuitConfirm = false }) {
                    Text("取消", color = MimTheme.colors.textMuted)
                }
            },
            containerColor = MimTheme.colors.surfaceElevated
        )
    }

    // Edit Notice Dialog
    if (showEditNoticeDialog) {
        var noticeText by remember { mutableStateOf(groupInfo?.notice ?: "") }
        AlertDialog(
            onDismissRequest = { showEditNoticeDialog = false },
            title = { Text("编辑群公告", color = MimTheme.colors.textPrimary) },
            text = {
                OutlinedTextField(
                    value = noticeText,
                    onValueChange = { noticeText = it },
                    placeholder = { Text("输入群公告内容...", color = MimTheme.colors.textMuted) },
                    modifier = Modifier.fillMaxWidth(),
                    maxLines = 5,
                    colors = OutlinedTextFieldDefaults.colors(
                        focusedBorderColor = MimTheme.colors.accent,
                        unfocusedBorderColor = MimTheme.colors.border,
                        focusedContainerColor = MimTheme.colors.inputBackground,
                        unfocusedContainerColor = MimTheme.colors.inputBackground,
                        focusedTextColor = MimTheme.colors.textPrimary,
                        unfocusedTextColor = MimTheme.colors.textPrimary
                    )
                )
            },
            confirmButton = {
                Button(
                    onClick = {
                        scope.launch {
                            val res = chatRepo.updateGroup(
                                groupId = groupId,
                                title = groupInfo?.title ?: "",
                                notice = noticeText.trim(),
                                isAllMuted = groupInfo?.isAllMuted ?: 0
                            )
                            showEditNoticeDialog = false
                            if (res.isSuccess) {
                                Toast.makeText(context, "公告已更新", Toast.LENGTH_SHORT).show()
                                refreshData()
                            } else {
                                Toast.makeText(context, "更新失败", Toast.LENGTH_SHORT).show()
                            }
                        }
                    },
                    colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent)
                ) {
                    Text("保存", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { showEditNoticeDialog = false }) {
                    Text("取消", color = MimTheme.colors.textMuted)
                }
            },
            containerColor = MimTheme.colors.surfaceElevated
        )
    }

    // Invite Friends Dialog
    if (showInviteDialog) {
        var friendList by remember { mutableStateOf<List<UserDto>>(emptyList()) }
        val selectedInvitees = remember { mutableStateListOf<Long>() }

        LaunchedEffect(Unit) {
            val res = authRepo.getFriends()
            if (res.isSuccess) {
                val existingIds = members.map { it.userId }.toSet()
                friendList = res.getOrDefault(emptyList()).filter { !existingIds.contains(it.id) }
            }
        }

        AlertDialog(
            onDismissRequest = { showInviteDialog = false },
            title = { Text("邀请好友加入群聊", color = MimTheme.colors.textPrimary) },
            text = {
                if (friendList.isEmpty()) {
                    Text("暂无可邀请的好友", color = MimTheme.colors.textMuted)
                } else {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .heightIn(max = 260.dp)
                            .verticalScroll(rememberScrollState())
                    ) {
                        friendList.forEach { f ->
                            val isChecked = selectedInvitees.contains(f.id)
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clip(RoundedCornerShape(8.dp))
                                    .clickable {
                                        if (isChecked) selectedInvitees.remove(f.id)
                                        else selectedInvitees.add(f.id)
                                    }
                                    .padding(vertical = 6.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Checkbox(
                                    checked = isChecked,
                                    onCheckedChange = { ch ->
                                        if (ch) selectedInvitees.add(f.id) else selectedInvitees.remove(f.id)
                                    },
                                    colors = CheckboxDefaults.colors(checkedColor = MimTheme.colors.accent)
                                )
                                MimAvatar(name = f.nickname.ifBlank { f.account }, size = 32.dp)
                                Spacer(modifier = Modifier.width(8.dp))
                                Text(
                                    text = f.nickname.ifBlank { f.account },
                                    color = MimTheme.colors.textPrimary,
                                    fontSize = 13.sp
                                )
                            }
                        }
                    }
                }
            },
            confirmButton = {
                Button(
                    onClick = {
                        if (selectedInvitees.isNotEmpty()) {
                            scope.launch {
                                val res = chatRepo.inviteGroup(groupId, selectedInvitees.toList())
                                showInviteDialog = false
                                if (res.isSuccess) {
                                    Toast.makeText(context, "邀请已发送", Toast.LENGTH_SHORT).show()
                                    refreshData()
                                } else {
                                    Toast.makeText(context, "邀请失败", Toast.LENGTH_SHORT).show()
                                }
                            }
                        }
                    },
                    enabled = selectedInvitees.isNotEmpty(),
                    colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent)
                ) {
                    Text("确认邀请 (${selectedInvitees.size})", color = Color.White)
                }
            },
            dismissButton = {
                TextButton(onClick = { showInviteDialog = false }) {
                    Text("取消", color = MimTheme.colors.textMuted)
                }
            },
            containerColor = MimTheme.colors.surfaceElevated
        )
    }
}
