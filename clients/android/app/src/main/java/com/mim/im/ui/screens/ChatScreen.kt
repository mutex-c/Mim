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
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.components.MimAvatar
import com.mim.im.ui.components.MimFileCard
import com.mim.im.ui.theme.MimTheme
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
    onBack: () -> Unit
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val listState = rememberLazyListState()

    val messages by chatRepo.getMessages(conversationId).collectAsState(initial = emptyList())
    var inputText by remember { mutableStateOf("") }
    var showAttachmentSheet by remember { mutableStateOf(false) }

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
                        mimeType = mimeType
                    )
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
                        mimeType = mimeType
                    )
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
                                text = if (isGroup) "群聊成员协同" else "端到端加密安全会话",
                                style = MaterialTheme.typography.bodySmall.copy(
                                    color = MimTheme.colors.textMuted,
                                    fontSize = 11.sp
                                )
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
                        IconButton(
                            onClick = { showAttachmentSheet = !showAttachmentSheet },
                            modifier = Modifier.size(40.dp)
                        ) {
                            Icon(
                                imageVector = if (showAttachmentSheet) Icons.Default.Close else Icons.Default.AddCircleOutline,
                                contentDescription = "多媒体附件",
                                tint = MimTheme.colors.accent,
                                modifier = Modifier.size(24.dp)
                            )
                        }

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

                        val canSend = inputText.isNotBlank()
                        Box(
                            modifier = Modifier
                                .size(40.dp)
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
                                                text = text
                                            )
                                        }
                                        inputText = ""
                                    }
                                },
                            contentAlignment = Alignment.Center
                        ) {
                            Icon(
                                imageVector = Icons.AutoMirrored.Filled.Send,
                                contentDescription = "发送",
                                tint = if (canSend) Color.White else MimTheme.colors.textMuted,
                                modifier = Modifier.size(20.dp)
                            )
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
                    }
                )
            }
        }
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
    onDownload: (MessageEntity, (Float) -> Unit, (File) -> Unit, (String) -> Unit) -> Unit,
    onRevoke: (MessageEntity) -> Unit
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
                name = msg.senderId.takeLast(2),
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
