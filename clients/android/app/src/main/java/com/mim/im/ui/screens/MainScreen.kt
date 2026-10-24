package com.mim.im.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.local.ConversationEntity
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.data.model.UserDto
import com.mim.im.network.WsManager
import com.mim.im.network.WsStatus
import com.mim.im.repository.AuthRepository
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.components.MimAvatar
import com.mim.im.ui.components.ServerConfigDialog
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*

@Composable
fun MainScreen(
    prefs: PreferencesHelper,
    authRepo: AuthRepository,
    chatRepo: ChatRepository,
    wsManager: WsManager,
    onOpenChat: (convId: String, title: String, targetId: Long, isGroup: Boolean) -> Unit,
    onLogout: () -> Unit
) {
    val scope = rememberCoroutineScope()
    var selectedTab by remember { mutableIntStateOf(0) } // 0: Chats, 1: Contacts, 2: Settings
    var showServerConfig by remember { mutableStateOf(false) }

    val conversations by chatRepo.getConversations().collectAsState(initial = emptyList())
    val wsStatus by wsManager.status.collectAsState()
    val user = prefs.currentUser

    LaunchedEffect(Unit) {
        chatRepo.refreshConversations()
    }

    if (showServerConfig) {
        ServerConfigDialog(
            currentUrl = prefs.serverUrl,
            onDismiss = { showServerConfig = false },
            onConfirm = { newUrl ->
                prefs.serverUrl = newUrl
                showServerConfig = false
                wsManager.disconnect()
                wsManager.connect()
            }
        )
    }

    Scaffold(
        containerColor = MimTheme.colors.background,
        topBar = {
            Surface(
                color = MimTheme.colors.background,
                modifier = Modifier
                    .fillMaxWidth()
                    .statusBarsPadding()
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // Current User Avatar
                    MimAvatar(
                        name = user?.nickname ?: user?.username ?: "Mim",
                        size = 36.dp,
                        imageUrl = user?.avatar,
                        isOnline = wsStatus == WsStatus.CONNECTED
                    )

                    Spacer(modifier = Modifier.width(12.dp))

                    Column(modifier = Modifier.weight(1f)) {
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Text(
                                text = when (selectedTab) {
                                    0 -> "消息会话"
                                    1 -> "通讯录"
                                    else -> "系统设置"
                                },
                                style = MaterialTheme.typography.titleMedium.copy(
                                    fontWeight = FontWeight.Bold,
                                    color = MimTheme.colors.textPrimary
                                )
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            // WS status pill
                            val (dotColor, statusTitle) = when (wsStatus) {
                                WsStatus.CONNECTED -> Color(0xFF22C55E) to "在线"
                                WsStatus.CONNECTING -> Color(0xFFF59E0B) to "连接中"
                                WsStatus.DISCONNECTED -> Color(0xFFEF4444) to "离线"
                            }
                            Surface(
                                shape = RoundedCornerShape(10.dp),
                                color = dotColor.copy(alpha = 0.12f),
                                modifier = Modifier.padding(top = 1.dp)
                            ) {
                                Row(
                                    modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                                    verticalAlignment = Alignment.CenterVertically
                                ) {
                                    Box(
                                        modifier = Modifier
                                            .size(6.dp)
                                            .clip(CircleShape)
                                            .background(dotColor)
                                    )
                                    Spacer(modifier = Modifier.width(4.dp))
                                    Text(
                                        text = statusTitle,
                                        fontSize = 10.sp,
                                        fontWeight = FontWeight.Bold,
                                        color = dotColor
                                    )
                                }
                            }
                        }
                    }

                    if (selectedTab == 0) {
                        IconButton(
                            onClick = { scope.launch { chatRepo.refreshConversations() } },
                            modifier = Modifier.size(36.dp)
                        ) {
                            Icon(
                                imageVector = Icons.Default.Refresh,
                                contentDescription = "刷新",
                                tint = MimTheme.colors.textMuted,
                                modifier = Modifier.size(20.dp)
                            )
                        }
                    }
                    IconButton(
                        onClick = { showServerConfig = true },
                        modifier = Modifier.size(36.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Dns,
                            contentDescription = "服务器节点",
                            tint = MimTheme.colors.textMuted,
                            modifier = Modifier.size(20.dp)
                        )
                    }
                }
            }
        },
        bottomBar = {
            Surface(
                color = MimTheme.colors.surfaceElevated,
                modifier = Modifier
                    .fillMaxWidth()
                    .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(0.dp))
            ) {
                NavigationBar(
                    containerColor = Color.Transparent,
                    tonalElevation = 0.dp,
                    modifier = Modifier.height(64.dp)
                ) {
                    val totalUnread = conversations.sumOf { it.unreadCount }

                    NavigationBarItem(
                        selected = selectedTab == 0,
                        onClick = { selectedTab = 0 },
                        icon = {
                            BadgedBox(
                                badge = {
                                    if (totalUnread > 0) {
                                        Badge(
                                            containerColor = MimTheme.colors.accent,
                                            contentColor = Color.White
                                        ) {
                                            Text(totalUnread.toString())
                                        }
                                    }
                                }
                            ) {
                                Icon(
                                    imageVector = if (selectedTab == 0) Icons.Default.ChatBubble else Icons.Default.ChatBubbleOutline,
                                    contentDescription = "消息"
                                )
                            }
                        },
                        label = { Text("消息", fontSize = 11.sp, fontWeight = if (selectedTab == 0) FontWeight.Bold else FontWeight.Normal) },
                        colors = NavigationBarItemDefaults.colors(
                            selectedIconColor = MimTheme.colors.accent,
                            selectedTextColor = MimTheme.colors.accent,
                            unselectedIconColor = MimTheme.colors.textMuted,
                            unselectedTextColor = MimTheme.colors.textMuted,
                            indicatorColor = MimTheme.colors.accentSubtle
                        )
                    )

                    NavigationBarItem(
                        selected = selectedTab == 1,
                        onClick = { selectedTab = 1 },
                        icon = {
                            Icon(
                                imageVector = if (selectedTab == 1) Icons.Default.People else Icons.Default.PeopleOutline,
                                contentDescription = "通讯录"
                            )
                        },
                        label = { Text("联系人", fontSize = 11.sp, fontWeight = if (selectedTab == 1) FontWeight.Bold else FontWeight.Normal) },
                        colors = NavigationBarItemDefaults.colors(
                            selectedIconColor = MimTheme.colors.accent,
                            selectedTextColor = MimTheme.colors.accent,
                            unselectedIconColor = MimTheme.colors.textMuted,
                            unselectedTextColor = MimTheme.colors.textMuted,
                            indicatorColor = MimTheme.colors.accentSubtle
                        )
                    )

                    NavigationBarItem(
                        selected = selectedTab == 2,
                        onClick = { selectedTab = 2 },
                        icon = {
                            Icon(
                                imageVector = if (selectedTab == 2) Icons.Default.Settings else Icons.Default.Tune,
                                contentDescription = "设置"
                            )
                        },
                        label = { Text("设置", fontSize = 11.sp, fontWeight = if (selectedTab == 2) FontWeight.Bold else FontWeight.Normal) },
                        colors = NavigationBarItemDefaults.colors(
                            selectedIconColor = MimTheme.colors.accent,
                            selectedTextColor = MimTheme.colors.accent,
                            unselectedIconColor = MimTheme.colors.textMuted,
                            unselectedTextColor = MimTheme.colors.textMuted,
                            indicatorColor = MimTheme.colors.accentSubtle
                        )
                    )
                }
            }
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            when (selectedTab) {
                0 -> ConversationListView(
                    conversations = conversations,
                    onConversationClick = { conv ->
                        val targetId = conv.targetId.toLongOrNull() ?: 0L
                        val isGroup = conv.convType == "group"
                        onOpenChat(conv.id, conv.title, targetId, isGroup)
                    }
                )
                1 -> ContactsView(
                    authRepo = authRepo,
                    onChatWithUser = { u ->
                        val myId = prefs.currentUser?.id?.toLongOrNull() ?: 0L
                        val targetId = u.id
                        val convId = if (myId < targetId) "${myId}_${targetId}" else "${targetId}_${myId}"
                        onOpenChat(convId, u.nickname.ifEmpty { u.account }, targetId, false)
                    }
                )
                2 -> SettingsScreen(
                    prefs = prefs,
                    wsManager = wsManager,
                    onLogout = onLogout
                )
            }
        }
    }
}

@Composable
fun ConversationListView(
    conversations: List<ConversationEntity>,
    onConversationClick: (ConversationEntity) -> Unit
) {
    if (conversations.isEmpty()) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(24.dp),
            contentAlignment = Alignment.Center
        ) {
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                Box(
                    modifier = Modifier
                        .size(64.dp)
                        .clip(CircleShape)
                        .background(MimTheme.colors.surfaceElevated)
                        .border(1.dp, MimTheme.colors.border, CircleShape),
                    contentAlignment = Alignment.Center
                ) {
                    Icon(
                        imageVector = Icons.Default.ChatBubbleOutline,
                        contentDescription = null,
                        tint = MimTheme.colors.textMuted,
                        modifier = Modifier.size(28.dp)
                    )
                }
                Spacer(modifier = Modifier.height(14.dp))
                Text(
                    text = "暂无活跃会话",
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.textPrimary
                    )
                )
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "可在「联系人」中点击任意成员发起实时对话",
                    style = MaterialTheme.typography.bodySmall.copy(
                        color = MimTheme.colors.textMuted
                    )
                )
            }
        }
        return
    }

    LazyColumn(modifier = Modifier.fillMaxSize()) {
        items(conversations, key = { it.id }) { conv ->
            ConversationRowItem(conversation = conv, onClick = { onConversationClick(conv) })
            HorizontalDivider(
                color = MimTheme.colors.border.copy(alpha = 0.5f),
                thickness = 0.5.dp,
                modifier = Modifier.padding(start = 72.dp)
            )
        }
    }
}

@Composable
fun ConversationRowItem(
    conversation: ConversationEntity,
    onClick: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        val isGroup = conversation.convType == "group"
        MimAvatar(
            name = conversation.title,
            size = 46.dp,
            isGroup = isGroup
        )

        Spacer(modifier = Modifier.width(12.dp))

        Column(modifier = Modifier.weight(1f)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = conversation.title,
                    style = MaterialTheme.typography.bodyLarge.copy(
                        fontWeight = FontWeight.SemiBold,
                        color = MimTheme.colors.textPrimary
                    ),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = formatTime(conversation.lastMessageTime),
                    style = MaterialTheme.typography.bodySmall.copy(
                        color = MimTheme.colors.textMuted,
                        fontSize = 11.sp
                    )
                )
            }

            Spacer(modifier = Modifier.height(4.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                val preview = formatLastMessage(conversation.lastMessage)
                Text(
                    text = preview,
                    style = MaterialTheme.typography.bodyMedium.copy(
                        color = if (conversation.unreadCount > 0) MimTheme.colors.textPrimary else MimTheme.colors.textMuted,
                        fontSize = 13.sp
                    ),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.weight(1f)
                )

                if (conversation.unreadCount > 0) {
                    Spacer(modifier = Modifier.width(8.dp))
                    Box(
                        modifier = Modifier
                            .clip(CircleShape)
                            .background(MimTheme.colors.accent)
                            .padding(horizontal = 6.dp, vertical = 2.dp)
                    ) {
                        Text(
                            text = if (conversation.unreadCount > 99) "99+" else conversation.unreadCount.toString(),
                            color = Color.White,
                            fontSize = 10.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }
                }
            }
        }
    }
}

fun formatLastMessage(raw: String): String {
    if (raw.isBlank()) return "暂无新消息"
    return when {
        raw.startsWith("/api/file/download") -> "[文件附件]"
        raw.startsWith("data:audio") || raw.startsWith("audio:") -> "[语音留言]"
        raw.startsWith("data:image") || raw.endsWith(".png") || raw.endsWith(".jpg") -> "[图片]"
        raw.contains("撤回了一条消息") -> "[消息已被撤回]"
        else -> raw
    }
}

@Composable
fun ContactsView(
    authRepo: AuthRepository,
    onChatWithUser: (UserDto) -> Unit
) {
    val scope = rememberCoroutineScope()
    var friends by remember { mutableStateOf<List<UserDto>>(emptyList()) }
    var searchKeyword by remember { mutableStateOf("") }
    var searchResults by remember { mutableStateOf<List<UserDto>>(emptyList()) }

    LaunchedEffect(Unit) {
        authRepo.getFriends().onSuccess { friends = it }
        if (friends.isEmpty()) {
            authRepo.getDemoUsers().onSuccess { friends = it }
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        OutlinedTextField(
            value = searchKeyword,
            onValueChange = {
                searchKeyword = it
                if (it.isNotBlank()) {
                    scope.launch {
                        authRepo.searchUsers(it).onSuccess { res -> searchResults = res }
                    }
                } else {
                    searchResults = emptyList()
                }
            },
            placeholder = { Text("搜索用户昵称或唯一账号...", color = MimTheme.colors.textMuted) },
            leadingIcon = {
                Icon(
                    imageVector = Icons.Default.Search,
                    contentDescription = "搜索",
                    tint = MimTheme.colors.accent
                )
            },
            singleLine = true,
            shape = RoundedCornerShape(12.dp),
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

        Spacer(modifier = Modifier.height(16.dp))

        val displayList = if (searchKeyword.isNotBlank()) searchResults else friends

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = if (searchKeyword.isNotBlank()) "搜索结果 (${displayList.size})" else "全部联系人 (${displayList.size})",
                style = MaterialTheme.typography.labelLarge.copy(
                    fontWeight = FontWeight.Bold,
                    color = MimTheme.colors.textPrimary
                )
            )
        }

        Spacer(modifier = Modifier.height(8.dp))

        LazyColumn(
            modifier = Modifier.fillMaxSize(),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(displayList) { user ->
                Surface(
                    shape = RoundedCornerShape(12.dp),
                    color = MimTheme.colors.surfaceElevated,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(12.dp))
                        .border(1.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
                        .clickable { onChatWithUser(user) }
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(12.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        MimAvatar(
                            name = user.nickname.ifEmpty { user.account },
                            size = 40.dp
                        )

                        Spacer(modifier = Modifier.width(12.dp))

                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = user.nickname.ifEmpty { user.account },
                                style = MaterialTheme.typography.bodyMedium.copy(
                                    fontWeight = FontWeight.Bold,
                                    color = MimTheme.colors.textPrimary
                                )
                            )
                            Spacer(modifier = Modifier.height(2.dp))
                            Text(
                                text = "@${user.account} · ID #${user.id}",
                                style = MaterialTheme.typography.bodySmall.copy(
                                    color = MimTheme.colors.textMuted,
                                    fontSize = 11.sp
                                )
                            )
                        }

                        Button(
                            onClick = { onChatWithUser(user) },
                            shape = RoundedCornerShape(8.dp),
                            contentPadding = PaddingValues(horizontal = 12.dp, vertical = 6.dp),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = MimTheme.colors.accent,
                                contentColor = Color.White
                            ),
                            modifier = Modifier.height(34.dp)
                        ) {
                            Text("发消息", fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                        }
                    }
                }
            }
        }
    }
}

private fun formatTime(timeMs: Long): String {
    if (timeMs <= 0) return ""
    val now = Calendar.getInstance()
    val msgTime = Calendar.getInstance().apply { timeInMillis = timeMs }

    return if (now.get(Calendar.DATE) == msgTime.get(Calendar.DATE)) {
        SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date(timeMs))
    } else {
        SimpleDateFormat("MM-dd HH:mm", Locale.getDefault()).format(Date(timeMs))
    }
}
