package com.mim.im.ui.screens

import android.widget.Toast
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ExitToApp
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.data.model.AIConfigDto
import com.mim.im.network.WsManager
import com.mim.im.network.WsStatus
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.components.MimAvatar
import com.mim.im.ui.components.ServerConfigDialog
import com.mim.im.ui.theme.ColorPalette
import com.mim.im.ui.theme.MimTheme
import com.mim.im.ui.theme.ThemeMode
import kotlinx.coroutines.launch

@Composable
fun SettingsScreen(
    prefs: PreferencesHelper,
    wsManager: WsManager,
    onLogout: () -> Unit,
    modifier: Modifier = Modifier,
    chatRepo: ChatRepository? = null
) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()
    val scrollState = rememberScrollState()
    var showServerDialog by remember { mutableStateOf(false) }

    val currentPalette by prefs.paletteFlow.collectAsState()
    val currentThemeMode by prefs.themeModeFlow.collectAsState()
    val wsStatus by wsManager.status.collectAsState()
    val user = prefs.currentUser

    // Profile form state
    var isEditingProfile by remember { mutableStateOf(false) }
    var nicknameInput by remember(user) { mutableStateOf(user?.nickname ?: user?.username ?: "") }
    var signatureInput by remember(user) { mutableStateOf("") }
    var isSavingProfile by remember { mutableStateOf(false) }

    // AI Config state
    var aiProvider by remember { mutableStateOf("deepseek") }
    var aiBaseUrl by remember { mutableStateOf("https://api.deepseek.com/v1") }
    var aiApiKey by remember { mutableStateOf("") }
    var aiModel by remember { mutableStateOf("deepseek-chat") }
    var showApiKey by remember { mutableStateOf(false) }
    var isTestingAi by remember { mutableStateOf(false) }
    var isSavingAi by remember { mutableStateOf(false) }
    var aiTestLatency by remember { mutableStateOf<Long?>(null) }
    var aiTestError by remember { mutableStateOf<String?>(null) }

    // Privacy & Permissions state
    var searchMode by remember { mutableIntStateOf(2) } // 0: none, 1: account only, 2: nickname & account
    var needFriendConfirm by remember { mutableStateOf(true) }
    var isSavingPrivacy by remember { mutableStateOf(false) }

    // Load remote AI config & user profile on initial launch
    LaunchedEffect(Unit) {
        if (chatRepo != null) {
            val aiRes = chatRepo.getAIConfig()
            if (aiRes.isSuccess) {
                val cfg = aiRes.getOrNull()
                if (cfg != null) {
                    aiProvider = cfg.provider.ifBlank { "deepseek" }
                    aiBaseUrl = cfg.baseUrl.ifBlank { "https://api.deepseek.com/v1" }
                    aiApiKey = cfg.apiKey
                    aiModel = cfg.model.ifBlank { "deepseek-chat" }
                }
            }
        }
    }

    if (showServerDialog) {
        ServerConfigDialog(
            currentUrl = prefs.serverUrl,
            onDismiss = { showServerDialog = false },
            onConfirm = { newUrl ->
                prefs.serverUrl = newUrl
                showServerDialog = false
                wsManager.disconnect()
                wsManager.connect()
            }
        )
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .background(MimTheme.colors.background)
            .verticalScroll(scrollState)
            .padding(16.dp)
    ) {
        // --- 1. Profile Card & Editor ---
        Surface(
            shape = RoundedCornerShape(16.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(16.dp))
        ) {
            Column(modifier = Modifier.padding(16.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    MimAvatar(
                        name = nicknameInput.ifBlank { user?.nickname ?: user?.username ?: "Mim" },
                        size = 56.dp,
                        imageUrl = user?.avatar,
                        isOnline = wsStatus == WsStatus.CONNECTED
                    )
                    Spacer(modifier = Modifier.width(16.dp))
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = nicknameInput.ifBlank { user?.nickname ?: user?.username ?: "未知用户" },
                            style = MaterialTheme.typography.titleMedium.copy(
                                fontWeight = FontWeight.Bold,
                                color = MimTheme.colors.textPrimary
                            )
                        )
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = "@${user?.username ?: "unknown"}",
                            style = MaterialTheme.typography.bodySmall.copy(
                                color = MimTheme.colors.textMuted
                            )
                        )
                        Spacer(modifier = Modifier.height(4.dp))
                        Surface(
                            shape = RoundedCornerShape(6.dp),
                            color = MimTheme.colors.surface,
                            modifier = Modifier.border(0.5.dp, MimTheme.colors.border, RoundedCornerShape(6.dp))
                        ) {
                            Text(
                                text = "UID: #${user?.id ?: "0"}",
                                fontSize = 11.sp,
                                fontWeight = FontWeight.Medium,
                                color = MimTheme.colors.textMuted,
                                modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                            )
                        }
                    }

                    IconButton(
                        onClick = { isEditingProfile = !isEditingProfile }
                    ) {
                        Icon(
                            imageVector = if (isEditingProfile) Icons.Default.ExpandLess else Icons.Default.Edit,
                            contentDescription = "编辑个人资料",
                            tint = MimTheme.colors.accent,
                            modifier = Modifier.size(20.dp)
                        )
                    }
                }

                if (isEditingProfile) {
                    Spacer(modifier = Modifier.height(14.dp))
                    HorizontalDivider(color = MimTheme.colors.border)
                    Spacer(modifier = Modifier.height(12.dp))

                    Text(
                        text = "用户昵称",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Medium,
                        color = MimTheme.colors.textMuted
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    OutlinedTextField(
                        value = nicknameInput,
                        onValueChange = { nicknameInput = it },
                        singleLine = true,
                        placeholder = { Text("输入新昵称", color = MimTheme.colors.textMuted) },
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(10.dp)
                    )

                    Spacer(modifier = Modifier.height(10.dp))

                    Text(
                        text = "个性签名",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Medium,
                        color = MimTheme.colors.textMuted
                    )
                    Spacer(modifier = Modifier.height(4.dp))
                    OutlinedTextField(
                        value = signatureInput,
                        onValueChange = { signatureInput = it },
                        singleLine = true,
                        placeholder = { Text("分享一段个性签名...", color = MimTheme.colors.textMuted) },
                        modifier = Modifier.fillMaxWidth(),
                        shape = RoundedCornerShape(10.dp)
                    )

                    // Genshin character persona quick presets when Genshin theme is active
                    if (currentPalette == ColorPalette.GENSHIN) {
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            text = "✦ 原神角色一键换装预设",
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Bold,
                            color = Color(0xFFD97706)
                        )
                        Spacer(modifier = Modifier.height(6.dp))
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .horizontalScroll(rememberScrollState()),
                            horizontalArrangement = Arrangement.spacedBy(6.dp)
                        ) {
                            val genshinChars = listOf(
                                "钟离" to "欲买桂花同载酒，只可惜故人...",
                                "芙宁娜" to "尽情欢呼吧，为水神的名号！",
                                "胡桃" to "大丘丘病了二丘丘瞧~",
                                "雷电将军" to "浮世之梦，唯有永恒最接近天理。",
                                "纳西妲" to "最初的贤者，欢迎你的到来。",
                                "派蒙" to "前面的区域，以后再来探索吧！"
                            )
                            genshinChars.forEach { (name, quote) ->
                                Surface(
                                    shape = RoundedCornerShape(8.dp),
                                    color = Color(0xFFF59E0B).copy(alpha = 0.15f),
                                    modifier = Modifier
                                        .clip(RoundedCornerShape(8.dp))
                                        .clickable {
                                            nicknameInput = name
                                            signatureInput = quote
                                            Toast.makeText(context, "已套用角色预设「$name」", Toast.LENGTH_SHORT).show()
                                        }
                                        .border(1.dp, Color(0xFFD97706).copy(alpha = 0.3f), RoundedCornerShape(8.dp))
                                ) {
                                    Text(
                                        text = name,
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.Bold,
                                        color = Color(0xFFD97706),
                                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 4.dp)
                                    )
                                }
                            }
                        }
                    }

                    Spacer(modifier = Modifier.height(14.dp))

                    Button(
                        onClick = {
                            if (chatRepo != null) {
                                isSavingProfile = true
                                scope.launch {
                                    val res = chatRepo.updateProfile(nicknameInput.trim(), signatureInput.trim(), user?.avatar ?: "")
                                    isSavingProfile = false
                                    if (res.isSuccess) {
                                        Toast.makeText(context, "个人资料更新成功", Toast.LENGTH_SHORT).show()
                                        isEditingProfile = false
                                    } else {
                                        Toast.makeText(context, res.exceptionOrNull()?.message ?: "保存失败", Toast.LENGTH_SHORT).show()
                                    }
                                }
                            }
                        },
                        enabled = !isSavingProfile && nicknameInput.isNotBlank(),
                        shape = RoundedCornerShape(10.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text(if (isSavingProfile) "保存中..." else "保存个人资料", fontSize = 13.sp)
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // --- 2. AI Engine Configuration Card ---
        Text(
            text = "AI 智能引擎配置 (同步对齐 Web)",
            style = MaterialTheme.typography.labelLarge.copy(
                fontWeight = FontWeight.Bold,
                color = MimTheme.colors.textPrimary
            ),
            modifier = Modifier.padding(start = 4.dp, bottom = 8.dp)
        )

        Surface(
            shape = RoundedCornerShape(14.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(14.dp))
        ) {
            Column(modifier = Modifier.padding(14.dp)) {
                // Provider presets
                Text(
                    text = "快速供应商预设",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Medium,
                    color = MimTheme.colors.textMuted
                )
                Spacer(modifier = Modifier.height(6.dp))
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .horizontalScroll(rememberScrollState()),
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    val presets = listOf(
                        Triple("DeepSeek", "https://api.deepseek.com/v1", "deepseek-chat"),
                        Triple("Ollama (本地)", "http://10.0.2.2:11434/v1", "llama3"),
                        Triple("OpenAI", "https://api.openai.com/v1", "gpt-4o-mini"),
                        Triple("自定义", "", "")
                    )
                    presets.forEach { (name, url, model) ->
                        val isSelected = (aiProvider == name.lowercase() || (name == "DeepSeek" && aiProvider == "deepseek"))
                        Surface(
                            shape = RoundedCornerShape(8.dp),
                            color = if (isSelected) MimTheme.colors.accent.copy(alpha = 0.15f) else MimTheme.colors.surface,
                            modifier = Modifier
                                .clip(RoundedCornerShape(8.dp))
                                .clickable {
                                    aiProvider = name.lowercase()
                                    if (url.isNotBlank()) aiBaseUrl = url
                                    if (model.isNotBlank()) aiModel = model
                                    aiTestLatency = null
                                    aiTestError = null
                                }
                                .border(
                                    width = if (isSelected) 1.5.dp else 1.dp,
                                    color = if (isSelected) MimTheme.colors.accent else MimTheme.colors.border,
                                    shape = RoundedCornerShape(8.dp)
                                )
                        ) {
                            Text(
                                text = name,
                                fontSize = 11.sp,
                                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Normal,
                                color = if (isSelected) MimTheme.colors.accent else MimTheme.colors.textPrimary,
                                modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp)
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(10.dp))

                Text(
                    text = "API Base URL",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Medium,
                    color = MimTheme.colors.textMuted
                )
                Spacer(modifier = Modifier.height(4.dp))
                OutlinedTextField(
                    value = aiBaseUrl,
                    onValueChange = { aiBaseUrl = it },
                    singleLine = true,
                    placeholder = { Text("https://api.deepseek.com/v1", color = MimTheme.colors.textMuted) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(10.dp)
                )

                Spacer(modifier = Modifier.height(10.dp))

                Text(
                    text = "API Key",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Medium,
                    color = MimTheme.colors.textMuted
                )
                Spacer(modifier = Modifier.height(4.dp))
                OutlinedTextField(
                    value = aiApiKey,
                    onValueChange = { aiApiKey = it },
                    singleLine = true,
                    placeholder = { Text("sk-...", color = MimTheme.colors.textMuted) },
                    visualTransformation = if (showApiKey) VisualTransformation.None else PasswordVisualTransformation(),
                    trailingIcon = {
                        IconButton(onClick = { showApiKey = !showApiKey }) {
                            Icon(
                                imageVector = if (showApiKey) Icons.Default.VisibilityOff else Icons.Default.Visibility,
                                contentDescription = "切换显示",
                                tint = MimTheme.colors.textMuted,
                                modifier = Modifier.size(18.dp)
                            )
                        }
                    },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(10.dp)
                )

                Spacer(modifier = Modifier.height(10.dp))

                Text(
                    text = "模型名称 (Model)",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Medium,
                    color = MimTheme.colors.textMuted
                )
                Spacer(modifier = Modifier.height(4.dp))
                OutlinedTextField(
                    value = aiModel,
                    onValueChange = { aiModel = it },
                    singleLine = true,
                    placeholder = { Text("deepseek-chat / gpt-4o-mini", color = MimTheme.colors.textMuted) },
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(10.dp)
                )

                if (aiTestLatency != null) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "✓ 连接测试成功，延迟: ${aiTestLatency}ms",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFF16A34A)
                    )
                } else if (aiTestError != null) {
                    Spacer(modifier = Modifier.height(8.dp))
                    Text(
                        text = "✕ 测试失败: $aiTestError",
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFFEF4444)
                    )
                }

                Spacer(modifier = Modifier.height(14.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    OutlinedButton(
                        onClick = {
                            if (chatRepo != null) {
                                isTestingAi = true
                                aiTestLatency = null
                                aiTestError = null
                                scope.launch {
                                    val cfg = AIConfigDto(
                                        provider = aiProvider,
                                        baseUrl = aiBaseUrl.trim(),
                                        apiKey = aiApiKey.trim(),
                                        model = aiModel.trim()
                                    )
                                    val res = chatRepo.testAIConfig(cfg)
                                    isTestingAi = false
                                    if (res.isSuccess) {
                                        aiTestLatency = res.getOrDefault(0L)
                                    } else {
                                        aiTestError = res.exceptionOrNull()?.message ?: "连接失败"
                                    }
                                }
                            }
                        },
                        enabled = !isTestingAi && aiBaseUrl.isNotBlank(),
                        shape = RoundedCornerShape(10.dp),
                        modifier = Modifier.weight(1f)
                    ) {
                        Text(if (isTestingAi) "测试中..." else "测试连接", fontSize = 12.sp)
                    }

                    Button(
                        onClick = {
                            if (chatRepo != null) {
                                isSavingAi = true
                                scope.launch {
                                    val cfg = AIConfigDto(
                                        provider = aiProvider,
                                        baseUrl = aiBaseUrl.trim(),
                                        apiKey = aiApiKey.trim(),
                                        model = aiModel.trim()
                                    )
                                    val res = chatRepo.saveAIConfig(cfg)
                                    isSavingAi = false
                                    if (res.isSuccess) {
                                        Toast.makeText(context, "AI 引擎配置已成功保存", Toast.LENGTH_SHORT).show()
                                    } else {
                                        Toast.makeText(context, res.exceptionOrNull()?.message ?: "保存失败", Toast.LENGTH_SHORT).show()
                                    }
                                }
                            }
                        },
                        enabled = !isSavingAi && aiBaseUrl.isNotBlank(),
                        shape = RoundedCornerShape(10.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                        modifier = Modifier.weight(1f)
                    ) {
                        Text(if (isSavingAi) "保存中..." else "保存配置", fontSize = 12.sp)
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // --- 3. Privacy & Permission Settings Card ---
        Text(
            text = "隐私与权限控制",
            style = MaterialTheme.typography.labelLarge.copy(
                fontWeight = FontWeight.Bold,
                color = MimTheme.colors.textPrimary
            ),
            modifier = Modifier.padding(start = 4.dp, bottom = 8.dp)
        )

        Surface(
            shape = RoundedCornerShape(14.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(14.dp))
        ) {
            Column(modifier = Modifier.padding(14.dp)) {
                Text(
                    text = "搜索发现模式",
                    fontSize = 11.sp,
                    fontWeight = FontWeight.Medium,
                    color = MimTheme.colors.textMuted
                )
                Spacer(modifier = Modifier.height(6.dp))

                val searchModes = listOf(
                    0 to "完全隐身",
                    1 to "仅账号",
                    2 to "账号或昵称 (推荐)"
                )
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(6.dp)
                ) {
                    searchModes.forEach { (mode, title) ->
                        val isSelected = searchMode == mode
                        Surface(
                            shape = RoundedCornerShape(8.dp),
                            color = if (isSelected) MimTheme.colors.accent.copy(alpha = 0.15f) else MimTheme.colors.surface,
                            modifier = Modifier
                                .weight(1f)
                                .clip(RoundedCornerShape(8.dp))
                                .clickable { searchMode = mode }
                                .border(
                                    width = if (isSelected) 1.5.dp else 1.dp,
                                    color = if (isSelected) MimTheme.colors.accent else MimTheme.colors.border,
                                    shape = RoundedCornerShape(8.dp)
                                )
                        ) {
                            Text(
                                text = title,
                                fontSize = 10.sp,
                                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Normal,
                                color = if (isSelected) MimTheme.colors.accent else MimTheme.colors.textPrimary,
                                modifier = Modifier
                                    .padding(vertical = 8.dp)
                                    .fillMaxWidth(),
                                textAlign = androidx.compose.ui.text.style.TextAlign.Center
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(14.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.SpaceBetween
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = "添加好友需验证",
                            fontSize = 13.sp,
                            fontWeight = FontWeight.SemiBold,
                            color = MimTheme.colors.textPrimary
                        )
                        Text(
                            text = "开启后，他人添加好友必须经过您同意确认",
                            fontSize = 11.sp,
                            color = MimTheme.colors.textMuted
                        )
                    }
                    Switch(
                        checked = needFriendConfirm,
                        onCheckedChange = { needFriendConfirm = it },
                        colors = SwitchDefaults.colors(checkedThumbColor = Color.White, checkedTrackColor = MimTheme.colors.accent)
                    )
                }

                Spacer(modifier = Modifier.height(12.dp))

                Button(
                    onClick = {
                        if (chatRepo != null) {
                            isSavingPrivacy = true
                            scope.launch {
                                val res = chatRepo.updateConfig(searchMode, needFriendConfirm)
                                isSavingPrivacy = false
                                if (res.isSuccess) {
                                    Toast.makeText(context, "隐私权限设置已同步更新", Toast.LENGTH_SHORT).show()
                                } else {
                                    Toast.makeText(context, res.exceptionOrNull()?.message ?: "保存失败", Toast.LENGTH_SHORT).show()
                                }
                            }
                        }
                    },
                    enabled = !isSavingPrivacy,
                    shape = RoundedCornerShape(10.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = MimTheme.colors.accent),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text(if (isSavingPrivacy) "同步中..." else "更新隐私权限设置", fontSize = 12.sp)
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // --- 4. Theme Mode Section ---
        Text(
            text = "外观与模式",
            style = MaterialTheme.typography.labelLarge.copy(
                fontWeight = FontWeight.Bold,
                color = MimTheme.colors.textPrimary
            ),
            modifier = Modifier.padding(start = 4.dp, bottom = 8.dp)
        )

        Surface(
            shape = RoundedCornerShape(14.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(14.dp))
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(6.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                ThemeMode.entries.forEach { mode ->
                    val isSelected = currentThemeMode == mode
                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .clip(RoundedCornerShape(10.dp))
                            .background(
                                if (isSelected) MimTheme.colors.accent
                                else Color.Transparent
                            )
                            .clickable { prefs.themeMode = mode }
                            .padding(vertical = 10.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = mode.nameZh,
                            style = MaterialTheme.typography.bodySmall.copy(
                                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Normal,
                                color = if (isSelected) Color.White else MimTheme.colors.textPrimary
                            )
                        )
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // --- 5. Color Palettes Section ---
        Text(
            text = "高定调色板 (对齐桌面端)",
            style = MaterialTheme.typography.labelLarge.copy(
                fontWeight = FontWeight.Bold,
                color = MimTheme.colors.textPrimary
            ),
            modifier = Modifier.padding(start = 4.dp, bottom = 8.dp)
        )

        Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
            ColorPalette.entries.forEach { palette ->
                val isSelected = currentPalette == palette
                Surface(
                    shape = RoundedCornerShape(14.dp),
                    color = MimTheme.colors.surfaceElevated,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clip(RoundedCornerShape(14.dp))
                        .border(
                            width = if (isSelected) 1.5.dp else 1.dp,
                            color = if (isSelected) MimTheme.colors.accent else MimTheme.colors.border,
                            shape = RoundedCornerShape(14.dp)
                        )
                        .clickable { prefs.palette = palette }
                ) {
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(14.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        // Palette preview dot
                        Box(
                            modifier = Modifier
                                .size(28.dp)
                                .clip(CircleShape)
                                .background(palette.previewColor)
                                .border(2.dp, Color.White, CircleShape)
                        )

                        Spacer(modifier = Modifier.width(14.dp))

                        Column(modifier = Modifier.weight(1f)) {
                            Text(
                                text = palette.nameZh,
                                style = MaterialTheme.typography.bodyMedium.copy(
                                    fontWeight = FontWeight.Bold,
                                    color = MimTheme.colors.textPrimary
                                )
                            )
                            Text(
                                text = palette.nameEn,
                                style = MaterialTheme.typography.bodySmall.copy(
                                    color = MimTheme.colors.textMuted
                                )
                            )
                        }

                        if (isSelected) {
                            Box(
                                modifier = Modifier
                                    .size(24.dp)
                                    .clip(CircleShape)
                                    .background(MimTheme.colors.accent),
                                contentAlignment = Alignment.Center
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Check,
                                    contentDescription = "已选择",
                                    tint = Color.White,
                                    modifier = Modifier.size(16.dp)
                                )
                            }
                        }
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // --- 6. Network & Server Section ---
        Text(
            text = "网络与局域网服务",
            style = MaterialTheme.typography.labelLarge.copy(
                fontWeight = FontWeight.Bold,
                color = MimTheme.colors.textPrimary
            ),
            modifier = Modifier.padding(start = 4.dp, bottom = 8.dp)
        )

        Surface(
            shape = RoundedCornerShape(14.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(14.dp))
        ) {
            Column(modifier = Modifier.padding(14.dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(
                            text = "服务节点地址",
                            style = MaterialTheme.typography.bodyMedium.copy(
                                fontWeight = FontWeight.SemiBold,
                                color = MimTheme.colors.textPrimary
                            )
                        )
                        Spacer(modifier = Modifier.height(2.dp))
                        Text(
                            text = prefs.serverUrl,
                            style = MaterialTheme.typography.bodySmall.copy(
                                color = MimTheme.colors.textMuted
                            )
                        )
                    }

                    // WS status chip
                    val (statusText, statusBg, statusColor) = when (wsStatus) {
                        WsStatus.CONNECTED -> Triple("在线", Color(0x1F22C55E), Color(0xFF16A34A))
                        WsStatus.CONNECTING -> Triple("连接中", Color(0x1FF59E0B), Color(0xFFD97706))
                        WsStatus.DISCONNECTED -> Triple("未连接", Color(0x1FEF4444), Color(0xFFDC2626))
                    }

                    Box(
                        modifier = Modifier
                            .clip(RoundedCornerShape(8.dp))
                            .background(statusBg)
                            .padding(horizontal = 8.dp, vertical = 4.dp)
                    ) {
                        Text(
                            text = statusText,
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Bold,
                            color = statusColor
                        )
                    }
                }

                Spacer(modifier = Modifier.height(12.dp))

                OutlinedButton(
                    onClick = { showServerDialog = true },
                    shape = RoundedCornerShape(10.dp),
                    colors = ButtonDefaults.outlinedButtonColors(
                        contentColor = MimTheme.colors.accent
                    ),
                    border = androidx.compose.foundation.BorderStroke(1.dp, MimTheme.colors.border),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Icon(
                        imageVector = Icons.Default.Dns,
                        contentDescription = null,
                        modifier = Modifier.size(16.dp)
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(text = "配置服务器 IP 与端口", fontSize = 13.sp)
                }
            }
        }

        Spacer(modifier = Modifier.height(24.dp))

        // --- 7. Logout Button ---
        Surface(
            shape = RoundedCornerShape(14.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(14.dp))
                .border(1.dp, Color(0xFFEF4444).copy(alpha = 0.25f), RoundedCornerShape(14.dp))
                .clickable(onClick = onLogout)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(14.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.Center
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ExitToApp,
                    contentDescription = "退出登录",
                    tint = Color(0xFFEF4444),
                    modifier = Modifier.size(20.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "退出登录",
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFFEF4444)
                    )
                )
            }
        }

        Spacer(modifier = Modifier.height(32.dp))
    }
}
