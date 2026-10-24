package com.mim.im.ui.screens

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import android.content.Intent
import android.net.Uri
import android.webkit.MimeTypeMap
import android.widget.Toast
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.Send
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
import androidx.core.content.FileProvider
import coil.compose.AsyncImage
import com.mim.im.data.local.MessageEntity
import com.mim.im.data.model.ConvType
import com.mim.im.data.model.GroupInfoDto
import com.mim.im.data.model.GroupMemberDto
import com.mim.im.repository.AuthRepository
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.components.*
import com.mim.im.ui.theme.MimTheme
import com.mim.im.utils.AudioPlayerHelper
import com.mim.im.utils.AudioRecorderHelper
import kotlinx.coroutines.launch
import java.io.File
import java.io.FileOutputStream
import java.text.SimpleDateFormat
import java.util.*

@Composable
fun ChatScreen(
    conversationId: String,
    title: String,
    targetId: Long,
    isGroup: Boolean,
    chatRepo: ChatRepository,
    authRepo: AuthRepository,
    onBack: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val listState = rememberLazyListState()

    val recorderHelper = remember { AudioRecorderHelper(context) }
    val playerHelper = remember { AudioPlayerHelper(context) }

    DisposableEffect(Unit) {
        onDispose {
            playerHelper.release()
        }
    }

    val messages by chatRepo.getMessages(conversationId).collectAsState(initial = emptyList())
    var inputText by remember { mutableStateOf("") }
    var isVoiceInputMode by remember { mutableStateOf(false) }
    var showAttachmentSheet by remember { mutableStateOf(false) }
    var showGroupSettings by remember { mutableStateOf(false) }
    var replyingToMessage by remember { mutableStateOf<MessageEntity?>(null) }

    // Group info state for mute checking
    var groupInfo by remember { mutableStateOf<GroupInfoDto?>(null) }
    var myMemberInfo by remember { mutableStateOf<GroupMemberDto?>(null) }

    LaunchedEffect(conversationId, isGroup) {
        if (isGroup) {
            val infoRes = chatRepo.getGroupInfo(targetId)
            if (infoRes.isSuccess) groupInfo = infoRes.getOrNull()

            val membersRes = chatRepo.getGroupMembers(targetId)
            if (membersRes.isSuccess) {
                myMemberInfo = membersRes.getOrDefault(emptyList()).find { it.userId.toString() == chatRepo.currentOwnerId }
            }
        }
    }

    val isUserMutedInGroup = remember(groupInfo, myMemberInfo) {
        if (!isGroup) false
        else {
            val role = myMemberInfo?.role ?: 3
            val isStaff = (role == 1 || role == 2)
            if (isStaff) false
            else (groupInfo?.isAllMuted == 1 || (myMemberInfo?.isMuted == true))
        }
    }

    LaunchedEffect(messages.size) {
        if (messages.isNotEmpty()) {
            listState.animateScrollToItem(messages.size - 1)
        }
    }

    LaunchedEffect(conversationId) {
        chatRepo.markConversationRead(conversationId)
    }

    val filePicker = rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
        uri?.let {
            scope.launch {
                val tempFile = copyUriToTempFile(context, uri)
                if (tempFile != null) {
                    val mimeType = context.contentResolver.getType(uri) ?: "application/octet-stream"
                    chatRepo.sendFileMessage(
                        conversationId = conversationId,
                        convType = if (isGroup) ConvType.GROUP else ConvType.SINGLE,
                        targetId = targetId,
                        file = tempFile,
                        mimeType = mimeType,
                        quotedMsgId = replyingToMessage?.msgId?.toLongOrNull() ?: 0L,
                        quotedContent = replyingToMessage?.content ?: ""
                    )
                    replyingToMessage = null
                }
            }
        }
    }

    val imagePicker = rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
        uri?.let {
            scope.launch {
                val tempFile = copyUriToTempFile(context, uri)
                if (tempFile != null) {
                    val mimeType = context.contentResolver.getType(uri) ?: "image/jpeg"
                    chatRepo.sendFileMessage(
                        conversationId = conversationId,
                        convType = if (isGroup) ConvType.GROUP else ConvType.SINGLE,
                        targetId = targetId,
                        file = tempFile,
                        mimeType = mimeType,
                        quotedMsgId = replyingToMessage?.msgId?.toLongOrNull() ?: 0L,
                        quotedContent = replyingToMessage?.content ?: ""
                    )
                    replyingToMessage = null
                }
            }
        }
    }

    Scaffold(
        containerColor = MimTheme.colors.background,
        topBar = {
            Surface(
                color = MimTheme.colors.surfaceElevated,
                modifier = Modifier
                    .fillMaxWidth()
                    .statusBarsPadding()
                    .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(0.dp))
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 8.dp, vertical = 8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    IconButton(onClick = onBack) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "返回",
                            tint = MimTheme.colors.textPrimary
                        )
                    }

                    MimAvatar(
                        name = title,
                        size = 38.dp,
                        isGroup = isGroup
                    )

                    Spacer(modifier = Modifier.width(10.dp))

                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = title,
                            style = MaterialTheme.typography.titleMedium.copy(
                                fontWeight = FontWeight.Bold,
                                color = MimTheme.colors.textPrimary
                            ),
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis
                        )
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Box(
                                modifier = Modifier
                                    .size(6.dp)
                                    .clip(CircleShape)
                                    .background(Color(0xFF22C55E))
                            )
                            Spacer(modifier = Modifier.width(4.dp))
                            Text(
                                text = if (isGroup) "群聊安全协同" else "端到端加密安全会话",
                                style = MaterialTheme.typography.bodySmall.copy(
                                    color = MimTheme.colors.textMuted,
                                    fontSize = 11.sp
                                )
                            )
                        }
                    }

                    if (isGroup) {
                        IconButton(onClick = { showGroupSettings = true }) {
                            Icon(
                                imageVector = Icons.Default.MoreVert,
                                contentDescription = "群设置",
                                tint = MimTheme.colors.textPrimary
                            )
                        }
                    }
                }
            }
        },
        bottomBar = {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .navigationBarsPadding()
                    .imePadding()
            ) {
                // 1. Quoted Message Banner (if active)
                replyingToMessage?.let { quoted ->
                    QuoteReplyBanner(
                        quotedMessage = quoted,
                        onDismiss = { replyingToMessage = null }
                    )
                }

                // 2. Muted Warning Bar or Input Bar
                if (isUserMutedInGroup) {
                    Surface(
                        color = MimTheme.colors.surfaceElevated,
                        modifier = Modifier
                            .fillMaxWidth()
                            .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(0.dp))
                            .padding(16.dp)
                    ) {
                        Box(contentAlignment = Alignment.Center) {
                            Text(
                                text = if (groupInfo?.isAllMuted == 1) "全员禁言中，仅管理员可发言" else "您已被管理员禁言",
                                style = MaterialTheme.typography.bodyMedium.copy(
                                    color = MimTheme.colors.textMuted,
                                    fontWeight = FontWeight.Medium
                                )
                            )
                        }
                    }
                } else {
                    Surface(
                        color = MimTheme.colors.surfaceElevated,
                        modifier = Modifier
                            .fillMaxWidth()
                            .border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(0.dp))
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 8.dp, vertical = 8.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            // Mic / Keyboard toggle
                            IconButton(
                                onClick = { isVoiceInputMode = !isVoiceInputMode },
                                modifier = Modifier.size(38.dp)
                            ) {
                                Icon(
                                    imageVector = if (isVoiceInputMode) Icons.Default.Keyboard else Icons.Default.Mic,
                                    contentDescription = "切换语音/文字",
                                    tint = MimTheme.colors.accent,
                                    modifier = Modifier.size(22.dp)
                                )
                            }

                            if (isVoiceInputMode) {
                                // Hold to talk voice bar
                                VoiceRecordBar(
                                    recorderHelper = recorderHelper,
                                    onVoiceRecorded = { recordedFile, _ ->
                                        scope.launch {
                                            chatRepo.sendFileMessage(
                                                conversationId = conversationId,
                                                convType = if (isGroup) ConvType.GROUP else ConvType.SINGLE,
                                                targetId = targetId,
                                                file = recordedFile,
                                                mimeType = "audio/mp4",
                                                quotedMsgId = replyingToMessage?.msgId?.toLongOrNull() ?: 0L,
                                                quotedContent = replyingToMessage?.content ?: ""
                                            )
                                            replyingToMessage = null
                                        }
                                    },
                                    modifier = Modifier
                                        .weight(1f)
                                        .padding(horizontal = 4.dp)
                                )
                            } else {
                                OutlinedTextField(
                                    value = inputText,
                                    onValueChange = { inputText = it },
                                    placeholder = { Text("输入消息...", color = MimTheme.colors.textMuted) },
                                    shape = RoundedCornerShape(22.dp),
                                    colors = OutlinedTextFieldDefaults.colors(
                                        focusedBorderColor = MimTheme.colors.accent,
                                        unfocusedBorderColor = MimTheme.colors.border,
                                        focusedContainerColor = MimTheme.colors.inputBackground,
                                        unfocusedContainerColor = MimTheme.colors.inputBackground,
                                        focusedTextColor = MimTheme.colors.textPrimary,
                                        unfocusedTextColor = MimTheme.colors.textPrimary
                                    ),
                                    maxLines = 4,
                                    modifier = Modifier
                                        .weight(1f)
                                        .padding(horizontal = 4.dp)
                                )
                            }

                            IconButton(
                                onClick = { showAttachmentSheet = !showAttachmentSheet },
                                modifier = Modifier.size(38.dp)
                            ) {
                                Icon(
                                    imageVector = if (showAttachmentSheet) Icons.Default.Close else Icons.Default.AddCircleOutline,
                                    contentDescription = "多媒体附件",
                                    tint = MimTheme.colors.accent,
                                    modifier = Modifier.size(24.dp)
                                )
                            }

                            if (!isVoiceInputMode) {
                                val canSend = inputText.isNotBlank()
                                Box(
                                    modifier = Modifier
                                        .size(38.dp)
                                        .clip(CircleShape)
                                        .background(if (canSend) MimTheme.colors.accent else MimTheme.colors.surface)
                                        .clickable(enabled = canSend) {
                                            val text = inputText.trim()
                                            if (text.isNotEmpty()) {
                                                scope.launch {
                                                    chatRepo.sendTextMessage(
                                                        conversationId = conversationId,
                                                        convType = if (isGroup) ConvType.GROUP else ConvType.SINGLE,
                                                        targetId = targetId,
                                                        text = text,
                                                        quotedMsgId = replyingToMessage?.msgId?.toLongOrNull() ?: 0L,
                                                        quotedContent = replyingToMessage?.content ?: ""
                                                    )
                                                }
                                                inputText = ""
                                                replyingToMessage = null
                                            }
                                        },
                                    contentAlignment = Alignment.Center
                                ) {
                                    Icon(
                                        imageVector = Icons.AutoMirrored.Filled.Send,
                                        contentDescription = "发送",
                                        tint = if (canSend) Color.White else MimTheme.colors.textMuted,
                                        modifier = Modifier.size(18.dp)
                                    )
                                }
                            }
                        }
                    }
                }

                // Attachment Drawer
                if (showAttachmentSheet) {
                    Surface(
                        color = MimTheme.colors.surface,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(16.dp),
                            horizontalArrangement = Arrangement.SpaceEvenly
                        ) {
                            AttachmentOptionItem(
                                icon = Icons.Default.Image,
                                label = "相册图片",
                                onClick = {
                                    showAttachmentSheet = false
                                    imagePicker.launch("image/*")
                                }
                            )
                            AttachmentOptionItem(
                                icon = Icons.Default.Folder,
                                label = "本地文件",
                                onClick = {
                                    showAttachmentSheet = false
                                    filePicker.launch("*/*")
                                }
                            )
                        }
                    }
                }
            }
        }
    ) { padding ->
        LazyColumn(
            state = listState,
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(horizontal = 12.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(messages, key = { it.msgId }) { msg ->
                val myId = chatRepo.currentOwnerId
                val isSelf = msg.senderId == myId

                MessageBubbleItem(
                    msg = msg,
                    isSelf = isSelf,
                    playerHelper = playerHelper,
                    onQuote = { entity -> replyingToMessage = entity },
                    onDownload = { entity, onProgress, onSuccess, onError ->
                        scope.launch {
                            chatRepo.downloadFile(entity, onProgress, onSuccess, onError)
                        }
                    },
                    onRevoke = { entity ->
                        scope.launch {
                            val success = chatRepo.revokeMessage(entity)
                            if (!success) {
                                Toast.makeText(context, "撤回失败（超出时限或已被撤回）", Toast.LENGTH_SHORT).show()
                            }
                        }
                    },
                    onScrollToMessage = { targetMsgId ->
                        val index = messages.indexOfFirst { it.msgId == targetMsgId }
                        if (index >= 0) {
                            scope.launch { listState.animateScrollToItem(index) }
                        }
                    }
                )
            }
        }
    }

    // Group Settings Bottom Sheet
    if (showGroupSettings && isGroup) {
        GroupSettingsSheet(
            groupId = targetId,
            currentOwnerId = chatRepo.currentOwnerId,
            chatRepo = chatRepo,
            authRepo = authRepo,
            onDismiss = { showGroupSettings = false },
            onGroupDisbandedOrQuit = {
                showGroupSettings = false
                onBack()
            }
        )
    }
}

@Composable
fun AttachmentOptionItem(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    onClick: () -> Unit
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier
            .clip(RoundedCornerShape(12.dp))
            .clickable(onClick = onClick)
            .padding(12.dp)
    ) {
        Box(
            modifier = Modifier
                .size(52.dp)
                .clip(CircleShape)
                .background(MimTheme.colors.surfaceElevated)
                .border(1.dp, MimTheme.colors.border, CircleShape),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                tint = MimTheme.colors.accent,
                modifier = Modifier.size(24.dp)
            )
        }
        Spacer(modifier = Modifier.height(6.dp))
        Text(
            text = label,
            style = MaterialTheme.typography.bodySmall.copy(
                fontWeight = FontWeight.Medium,
                color = MimTheme.colors.textPrimary
            )
        )
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun MessageBubbleItem(
    msg: MessageEntity,
    isSelf: Boolean,
    playerHelper: AudioPlayerHelper,
    onQuote: (MessageEntity) -> Unit,
    onDownload: (MessageEntity, (Float) -> Unit, (File) -> Unit, (String) -> Unit) -> Unit,
    onRevoke: (MessageEntity) -> Unit,
    onScrollToMessage: (String) -> Unit
) {
    val context = LocalContext.current
    var showMenu by remember { mutableStateOf(false) }

    if (msg.contentType == "revoked") {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 4.dp),
            contentAlignment = Alignment.Center
        ) {
            Surface(
                shape = RoundedCornerShape(12.dp),
                color = MimTheme.colors.surfaceElevated.copy(alpha = 0.8f),
                modifier = Modifier.border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
            ) {
                Text(
                    text = if (isSelf) "你撤回了一条消息" else "对方撤回了一条消息",
                    style = MaterialTheme.typography.bodySmall.copy(
                        color = MimTheme.colors.textMuted,
                        fontSize = 11.sp
                    ),
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
                )
            }
        }
        return
    }

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 2.dp),
        horizontalArrangement = if (isSelf) Arrangement.End else Arrangement.Start
    ) {
        if (!isSelf) {
            MimAvatar(
                name = msg.senderName.ifBlank { msg.senderId.takeLast(2) },
                size = 34.dp
            )
            Spacer(modifier = Modifier.width(8.dp))
        }

        Box {
            if (msg.contentType == "file") {
                // Standalone QQ-style File Card
                var isDownloading by remember { mutableStateOf(false) }
                var downloadProgress by remember { mutableFloatStateOf(0f) }
                val isDownloaded = msg.isDownloaded || (!msg.localFilePath.isNullOrEmpty() && File(msg.localFilePath).exists())

                Box(
                    modifier = Modifier
                        .widthIn(max = 280.dp)
                        .combinedClickable(
                            onClick = {},
                            onLongClick = { showMenu = true }
                        )
                ) {
                    Column {
                        if (!msg.quotedContent.isNullOrBlank()) {
                            InBubbleQuoteView(
                                quotedContent = msg.quotedContent,
                                isSelf = isSelf,
                                onClick = { onScrollToMessage(msg.quotedMsgId) }
                            )
                            Spacer(modifier = Modifier.height(4.dp))
                        }

                        MimFileCard(
                            fileName = msg.fileName ?: "未知文件",
                            fileSize = msg.fileSize,
                            isDownloaded = isDownloaded,
                            isDownloading = isDownloading,
                            downloadProgress = downloadProgress,
                            onClick = {
                                if (isDownloaded && !msg.localFilePath.isNullOrEmpty()) {
                                    openDownloadedFile(context, File(msg.localFilePath))
                                } else if (!isDownloading) {
                                    isDownloading = true
                                    onDownload(
                                        msg,
                                        { progress -> downloadProgress = progress },
                                        { _ -> isDownloading = false },
                                        { err ->
                                            isDownloading = false
                                            Toast.makeText(context, err, Toast.LENGTH_SHORT).show()
                                        }
                                    )
                                }
                            }
                        )
                    }
                }
            } else {
                Surface(
                    shape = RoundedCornerShape(
                        topStart = if (isSelf) 16.dp else 4.dp,
                        topEnd = if (isSelf) 4.dp else 16.dp,
                        bottomStart = 16.dp,
                        bottomEnd = 16.dp
                    ),
                    color = if (isSelf) MimTheme.colors.bubbleSelf else MimTheme.colors.bubblePeer,
                    modifier = Modifier
                        .widthIn(max = 280.dp)
                        .border(
                            width = if (isSelf) 0.dp else 1.dp,
                            color = if (isSelf) Color.Transparent else MimTheme.colors.border,
                            shape = RoundedCornerShape(
                                topStart = if (isSelf) 16.dp else 4.dp,
                                topEnd = if (isSelf) 4.dp else 16.dp,
                                bottomStart = 16.dp,
                                bottomEnd = 16.dp
                            )
                        )
                        .combinedClickable(
                            onClick = {},
                            onLongClick = { showMenu = true }
                        )
                ) {
                    Column(modifier = Modifier.padding(10.dp)) {
                        // In-Bubble Quoted Preview
                        if (!msg.quotedContent.isNullOrBlank()) {
                            InBubbleQuoteView(
                                quotedContent = msg.quotedContent,
                                isSelf = isSelf,
                                onClick = { onScrollToMessage(msg.quotedMsgId) }
                            )
                            Spacer(modifier = Modifier.height(6.dp))
                        }

                        when (msg.contentType) {
                            "text" -> {
                                Text(
                                    text = msg.content,
                                    style = MaterialTheme.typography.bodyMedium.copy(
                                        color = if (isSelf) MimTheme.colors.bubbleSelfText else MimTheme.colors.bubblePeerText,
                                        lineHeight = 20.sp
                                    )
                                )
                            }
                            "voice" -> {
                                val currentPlayingId by playerHelper.currentPlayingMsgId.collectAsState()
                                val isPlaying by playerHelper.isPlaying.collectAsState()
                                val progress by playerHelper.playbackProgress.collectAsState()
                                val isThisPlaying = (currentPlayingId == msg.msgId && isPlaying)

                                MimVoiceBubble(
                                    isPlaying = isThisPlaying,
                                    progress = if (currentPlayingId == msg.msgId) progress else 0f,
                                    isSelf = isSelf,
                                    onTogglePlay = {
                                        val mediaSource = msg.localFilePath ?: msg.mediaUrl ?: ""
                                        if (mediaSource.isNotBlank()) {
                                            playerHelper.playOrToggle(msg.msgId, mediaSource)
                                        }
                                    }
                                )
                            }
                            "image" -> {
                                AsyncImage(
                                    model = msg.mediaUrl ?: msg.localFilePath,
                                    contentDescription = "图片",
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .heightIn(max = 200.dp)
                                        .clip(RoundedCornerShape(8.dp))
                                )
                            }
                        }

                        Spacer(modifier = Modifier.height(4.dp))

                        // Timestamp & status
                        Row(
                            modifier = Modifier.align(Alignment.End),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Text(
                                text = SimpleDateFormat("HH:mm", Locale.getDefault()).format(Date(msg.timestamp)),
                                fontSize = 10.sp,
                                color = if (isSelf) MimTheme.colors.bubbleSelfText.copy(alpha = 0.7f) else MimTheme.colors.textMuted
                            )
                            if (isSelf) {
                                Spacer(modifier = Modifier.width(4.dp))
                                when (msg.status) {
                                    "sending" -> CircularProgressIndicator(
                                        modifier = Modifier.size(10.dp),
                                        strokeWidth = 1.dp,
                                        color = MimTheme.colors.bubbleSelfText.copy(alpha = 0.7f)
                                    )
                                    "failed" -> Icon(
                                        Icons.Default.Error,
                                        contentDescription = "失败",
                                        tint = Color(0xFFEF4444),
                                        modifier = Modifier.size(12.dp)
                                    )
                                    else -> Icon(
                                        Icons.Default.Done,
                                        contentDescription = "已送达",
                                        tint = MimTheme.colors.bubbleSelfText.copy(alpha = 0.85f),
                                        modifier = Modifier.size(12.dp)
                                    )
                                }
                            }
                        }
                    }
                }
            }

            DropdownMenu(
                expanded = showMenu,
                onDismissRequest = { showMenu = false },
                modifier = Modifier
                    .background(MimTheme.colors.surfaceElevated)
                    .border(1.dp, MimTheme.colors.border, RoundedCornerShape(8.dp))
            ) {
                // 1. Reply / Quote
                DropdownMenuItem(
                    text = { Text("引用回复", color = MimTheme.colors.textPrimary) },
                    onClick = {
                        showMenu = false
                        onQuote(msg)
                    },
                    leadingIcon = { Icon(Icons.Default.Reply, contentDescription = null, tint = MimTheme.colors.accent) }
                )

                // 2. Copy Text
                if (msg.contentType == "text" && msg.content.isNotBlank()) {
                    DropdownMenuItem(
                        text = { Text("复制内容", color = MimTheme.colors.textPrimary) },
                        onClick = {
                            showMenu = false
                            val clipboard = context.getSystemService(Context.CLIPBOARD_SERVICE) as ClipboardManager
                            clipboard.setPrimaryClip(ClipData.newPlainText("Mim Message", msg.content))
                            Toast.makeText(context, "已复制", Toast.LENGTH_SHORT).show()
                        },
                        leadingIcon = { Icon(Icons.Default.ContentCopy, contentDescription = null, tint = MimTheme.colors.accent) }
                    )
                }

                // 3. Revoke (Self only)
                if (isSelf) {
                    DropdownMenuItem(
                        text = { Text("撤回消息", color = Color(0xFFDC2626)) },
                        onClick = {
                            showMenu = false
                            onRevoke(msg)
                        },
                        leadingIcon = { Icon(Icons.Default.Delete, contentDescription = null, tint = Color(0xFFDC2626)) }
                    )
                }
            }
        }
    }
}

private fun openDownloadedFile(context: Context, file: File) {
    try {
        if (!file.exists()) {
            Toast.makeText(context, "文件不存在", Toast.LENGTH_SHORT).show()
            return
        }

        val uri = FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            file
        )

        val extension = file.extension.lowercase(Locale.getDefault())
        val mimeType = MimeTypeMap.getSingleton().getMimeTypeFromExtension(extension) ?: "*/*"

        val intent = Intent(Intent.ACTION_VIEW).apply {
            setDataAndType(uri, mimeType)
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }

        context.startActivity(Intent.createChooser(intent, "打开文件"))
    } catch (_: Exception) {
        Toast.makeText(context, "未找到能够打开该格式文件的应用程序", Toast.LENGTH_SHORT).show()
    }
}

private fun copyUriToTempFile(context: Context, uri: Uri): File? {
    return try {
        val resolver = context.contentResolver
        var fileName = "mim_upload_${System.currentTimeMillis()}"
        val cursor = resolver.query(uri, null, null, null, null)
        cursor?.use {
            if (it.moveToFirst()) {
                val nameIndex = it.getColumnIndex(android.provider.OpenableColumns.DISPLAY_NAME)
                if (nameIndex != -1) {
                    fileName = it.getString(nameIndex)
                }
            }
        }

        val tempFile = File(context.cacheDir, fileName)
        resolver.openInputStream(uri)?.use { input ->
            FileOutputStream(tempFile).use { output ->
                input.copyTo(output)
            }
        }
        tempFile
    } catch (_: Exception) {
        null
    }
}
