package com.mim.im.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.data.model.UserDto
import com.mim.im.repository.AuthRepository
import com.mim.im.ui.components.MimAvatar
import com.mim.im.ui.components.ServerConfigDialog
import com.mim.im.ui.theme.MimTheme
import kotlinx.coroutines.launch

@Composable
fun AuthScreen(
    authRepo: AuthRepository,
    prefs: PreferencesHelper,
    onAuthSuccess: () -> Unit
) {
    val scope = rememberCoroutineScope()
    val scrollState = rememberScrollState()
    var selectedTab by remember { mutableIntStateOf(0) } // 0: Login, 1: Register

    var account by remember { mutableStateOf("alex") }
    var password by remember { mutableStateOf("123456") }
    var nickname by remember { mutableStateOf("") }
    var confirmPassword by remember { mutableStateOf("") }
    var showPassword by remember { mutableStateOf(false) }

    var isLoading by remember { mutableStateOf(false) }
    var errorMessage by remember { mutableStateOf<String?>(null) }
    var showServerConfig by remember { mutableStateOf(false) }
    var demoUsersWithTokens by remember { mutableStateOf<List<com.mim.im.data.model.DemoUserDto>>(emptyList()) }

    LaunchedEffect(Unit) {
        scope.launch {
            authRepo.getDemoUsersWithTokens().onSuccess {
                demoUsersWithTokens = it
            }
        }
    }

    if (showServerConfig) {
        ServerConfigDialog(
            currentUrl = prefs.serverUrl,
            onDismiss = { showServerConfig = false },
            onConfirm = { newUrl ->
                prefs.serverUrl = newUrl
                showServerConfig = false
            }
        )
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MimTheme.colors.background)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(scrollState)
                .padding(horizontal = 24.dp)
                .padding(top = 48.dp, bottom = 32.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            // Top Server Config Button
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.End
            ) {
                Surface(
                    shape = RoundedCornerShape(20.dp),
                    color = MimTheme.colors.surfaceElevated,
                    modifier = Modifier
                        .border(1.dp, MimTheme.colors.border, RoundedCornerShape(20.dp))
                        .clickable { showServerConfig = true }
                ) {
                    Row(
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Icon(
                            imageVector = Icons.Default.Dns,
                            contentDescription = "服务器配置",
                            tint = MimTheme.colors.accent,
                            modifier = Modifier.size(14.dp)
                        )
                        Spacer(modifier = Modifier.width(6.dp))
                        Text(
                            text = "服务节点",
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Medium,
                            color = MimTheme.colors.textPrimary
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            // --- Brand Logo & Header ---
            Box(
                modifier = Modifier
                    .size(64.dp)
                    .clip(RoundedCornerShape(18.dp))
                    .background(MimTheme.colors.accent),
                contentAlignment = Alignment.Center
            ) {
                Icon(
                    imageVector = Icons.Default.Shield,
                    contentDescription = "Mim",
                    tint = Color.White,
                    modifier = Modifier.size(36.dp)
                )
            }

            Spacer(modifier = Modifier.height(16.dp))

            Text(
                text = "Mim",
                style = MaterialTheme.typography.headlineMedium.copy(
                    fontWeight = FontWeight.Black,
                    color = MimTheme.colors.textPrimary,
                    letterSpacing = 1.sp
                )
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = "企业级端到端即时通讯系统",
                style = MaterialTheme.typography.bodyMedium.copy(
                    color = MimTheme.colors.textMuted
                )
            )

            Spacer(modifier = Modifier.height(28.dp))

            // --- Segmented Switcher Tab ---
            Surface(
                shape = RoundedCornerShape(12.dp),
                color = MimTheme.colors.surfaceElevated,
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
            ) {
                Row(
                    modifier = Modifier.padding(4.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    listOf("账号登录", "快速注册").forEachIndexed { index, title ->
                        val isSelected = selectedTab == index
                        Box(
                            modifier = Modifier
                                .weight(1f)
                                .clip(RoundedCornerShape(8.dp))
                                .background(
                                    if (isSelected) MimTheme.colors.accent
                                    else Color.Transparent
                                )
                                .clickable {
                                    selectedTab = index
                                    errorMessage = null
                                }
                                .padding(vertical = 10.dp),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = title,
                                style = MaterialTheme.typography.bodyMedium.copy(
                                    fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Medium,
                                    color = if (isSelected) Color.White else MimTheme.colors.textMuted
                                )
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            // --- Form Inputs Card ---
            Surface(
                shape = RoundedCornerShape(16.dp),
                color = MimTheme.colors.surfaceElevated,
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, MimTheme.colors.border, RoundedCornerShape(16.dp))
            ) {
                Column(modifier = Modifier.padding(16.dp)) {
                    // Account
                    OutlinedTextField(
                        value = account,
                        onValueChange = { account = it },
                        label = { Text("账号 / 用户名") },
                        leadingIcon = {
                            Icon(
                                imageVector = Icons.Default.Person,
                                contentDescription = null,
                                tint = MimTheme.colors.accent,
                                modifier = Modifier.size(20.dp)
                            )
                        },
                        singleLine = true,
                        shape = RoundedCornerShape(12.dp),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = MimTheme.colors.accent,
                            unfocusedBorderColor = MimTheme.colors.border,
                            focusedContainerColor = MimTheme.colors.surface,
                            unfocusedContainerColor = MimTheme.colors.surface,
                            focusedTextColor = MimTheme.colors.textPrimary,
                            unfocusedTextColor = MimTheme.colors.textPrimary,
                            focusedLabelColor = MimTheme.colors.accent,
                            unfocusedLabelColor = MimTheme.colors.textMuted
                        ),
                        modifier = Modifier.fillMaxWidth()
                    )

                    if (selectedTab == 1) {
                        Spacer(modifier = Modifier.height(12.dp))
                        OutlinedTextField(
                            value = nickname,
                            onValueChange = { nickname = it },
                            label = { Text("用户昵称") },
                            leadingIcon = {
                                Icon(
                                    imageVector = Icons.Default.Badge,
                                    contentDescription = null,
                                    tint = MimTheme.colors.accent,
                                    modifier = Modifier.size(20.dp)
                                )
                            },
                            singleLine = true,
                            shape = RoundedCornerShape(12.dp),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = MimTheme.colors.accent,
                                unfocusedBorderColor = MimTheme.colors.border,
                                focusedContainerColor = MimTheme.colors.surface,
                                unfocusedContainerColor = MimTheme.colors.surface,
                                focusedTextColor = MimTheme.colors.textPrimary,
                                unfocusedTextColor = MimTheme.colors.textPrimary,
                                focusedLabelColor = MimTheme.colors.accent,
                                unfocusedLabelColor = MimTheme.colors.textMuted
                            ),
                            modifier = Modifier.fillMaxWidth()
                        )
                    }

                    Spacer(modifier = Modifier.height(12.dp))

                    // Password
                    OutlinedTextField(
                        value = password,
                        onValueChange = { password = it },
                        label = { Text("登录密码") },
                        leadingIcon = {
                            Icon(
                                imageVector = Icons.Default.Lock,
                                contentDescription = null,
                                tint = MimTheme.colors.accent,
                                modifier = Modifier.size(20.dp)
                            )
                        },
                        trailingIcon = {
                            IconButton(onClick = { showPassword = !showPassword }) {
                                Icon(
                                    imageVector = if (showPassword) Icons.Default.Visibility else Icons.Default.VisibilityOff,
                                    contentDescription = if (showPassword) "隐藏密码" else "显示密码",
                                    tint = MimTheme.colors.textMuted,
                                    modifier = Modifier.size(20.dp)
                                )
                            }
                        },
                        singleLine = true,
                        visualTransformation = if (showPassword) VisualTransformation.None else PasswordVisualTransformation(),
                        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
                        shape = RoundedCornerShape(12.dp),
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = MimTheme.colors.accent,
                            unfocusedBorderColor = MimTheme.colors.border,
                            focusedContainerColor = MimTheme.colors.surface,
                            unfocusedContainerColor = MimTheme.colors.surface,
                            focusedTextColor = MimTheme.colors.textPrimary,
                            unfocusedTextColor = MimTheme.colors.textPrimary,
                            focusedLabelColor = MimTheme.colors.accent,
                            unfocusedLabelColor = MimTheme.colors.textMuted
                        ),
                        modifier = Modifier.fillMaxWidth()
                    )

                    if (selectedTab == 1) {
                        Spacer(modifier = Modifier.height(12.dp))
                        OutlinedTextField(
                            value = confirmPassword,
                            onValueChange = { confirmPassword = it },
                            label = { Text("确认密码") },
                            leadingIcon = {
                                Icon(
                                    imageVector = Icons.Default.LockReset,
                                    contentDescription = null,
                                    tint = MimTheme.colors.accent,
                                    modifier = Modifier.size(20.dp)
                                )
                            },
                            singleLine = true,
                            visualTransformation = if (showPassword) VisualTransformation.None else PasswordVisualTransformation(),
                            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
                            shape = RoundedCornerShape(12.dp),
                            colors = OutlinedTextFieldDefaults.colors(
                                focusedBorderColor = MimTheme.colors.accent,
                                unfocusedBorderColor = MimTheme.colors.border,
                                focusedContainerColor = MimTheme.colors.surface,
                                unfocusedContainerColor = MimTheme.colors.surface,
                                focusedTextColor = MimTheme.colors.textPrimary,
                                unfocusedTextColor = MimTheme.colors.textPrimary,
                                focusedLabelColor = MimTheme.colors.accent,
                                unfocusedLabelColor = MimTheme.colors.textMuted
                            ),
                            modifier = Modifier.fillMaxWidth()
                        )
                    }

                    if (errorMessage != null) {
                        Spacer(modifier = Modifier.height(12.dp))
                        Surface(
                            shape = RoundedCornerShape(8.dp),
                            color = Color(0xFFEF4444).copy(alpha = 0.1f),
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = errorMessage!!,
                                color = Color(0xFFDC2626),
                                style = MaterialTheme.typography.bodySmall.copy(fontWeight = FontWeight.Medium),
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp)
                            )
                        }
                    }

                    Spacer(modifier = Modifier.height(20.dp))

                    Button(
                        onClick = {
                            errorMessage = null
                            isLoading = true
                            scope.launch {
                                if (selectedTab == 0) {
                                    val res = authRepo.login(account.trim(), password.trim())
                                    isLoading = false
                                    res.onSuccess {
                                        onAuthSuccess()
                                    }.onFailure {
                                        errorMessage = it.message ?: "登录失败，请检查服务地址和账号密码"
                                    }
                                } else {
                                    if (password != confirmPassword) {
                                        isLoading = false
                                        errorMessage = "两次输入的密码不一致"
                                        return@launch
                                    }
                                    val res = authRepo.register(
                                        account.trim(),
                                        password.trim(),
                                        nickname.trim().ifEmpty { account.trim() }
                                    )
                                    isLoading = false
                                    res.onSuccess {
                                        val loginRes = authRepo.login(account.trim(), password.trim())
                                        loginRes.onSuccess { onAuthSuccess() }
                                    }.onFailure {
                                        errorMessage = it.message ?: "注册失败"
                                    }
                                }
                            }
                        },
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(48.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = MimTheme.colors.accent,
                            contentColor = Color.White
                        ),
                        enabled = !isLoading && account.isNotBlank() && password.isNotBlank()
                    ) {
                        if (isLoading) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(20.dp),
                                color = Color.White,
                                strokeWidth = 2.dp
                            )
                        } else {
                            Text(
                                text = if (selectedTab == 0) "安全登录" else "注册并登录",
                                fontWeight = FontWeight.Bold,
                                fontSize = 15.sp
                            )
                        }
                    }
                }
            }

            // --- Quick Demo Login Section ---
            if (selectedTab == 0) {
                val demoList = if (demoUsersWithTokens.isNotEmpty()) demoUsersWithTokens else listOf(
                    com.mim.im.data.model.DemoUserDto(UserDto(id = 1, account = "alex", nickname = "Alex Chen"), com.mim.im.data.model.TokenPairDto()),
                    com.mim.im.data.model.DemoUserDto(UserDto(id = 2, account = "sarah", nickname = "Sarah Jenkins"), com.mim.im.data.model.TokenPairDto()),
                    com.mim.im.data.model.DemoUserDto(UserDto(id = 3, account = "michael", nickname = "Michael Zhang"), com.mim.im.data.model.TokenPairDto()),
                    com.mim.im.data.model.DemoUserDto(UserDto(id = 4, account = "david", nickname = "David Miller"), com.mim.im.data.model.TokenPairDto())
                )

                Spacer(modifier = Modifier.height(24.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    HorizontalDivider(modifier = Modifier.weight(1f), color = MimTheme.colors.border)
                    Text(
                        text = " 快速体验演示成员 ",
                        style = MaterialTheme.typography.bodySmall.copy(color = MimTheme.colors.textMuted)
                    )
                    HorizontalDivider(modifier = Modifier.weight(1f), color = MimTheme.colors.border)
                }

                Spacer(modifier = Modifier.height(14.dp))

                LazyRow(
                    horizontalArrangement = Arrangement.spacedBy(10.dp),
                    modifier = Modifier.fillMaxWidth()
                ) {
                    items(demoList) { item ->
                        Surface(
                            shape = RoundedCornerShape(12.dp),
                            color = MimTheme.colors.surfaceElevated,
                            modifier = Modifier
                                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
                                .clickable {
                                    account = item.user.account
                                    password = "123456"
                                    if (item.token.accessToken.isNotBlank()) {
                                        // Instant zero-latency login using pre-minted token
                                        authRepo.loginWithToken(item.user, item.token.accessToken)
                                        onAuthSuccess()
                                    } else {
                                        scope.launch {
                                            isLoading = true
                                            val res = authRepo.login(item.user.account, "123456")
                                            isLoading = false
                                            res.onSuccess { onAuthSuccess() }
                                                .onFailure { errorMessage = it.message ?: "登录失败" }
                                        }
                                    }
                                }
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                MimAvatar(
                                    name = item.user.nickname,
                                    size = 28.dp
                                )
                                Spacer(modifier = Modifier.width(8.dp))
                                Column {
                                    Text(
                                        text = item.user.nickname,
                                        style = MaterialTheme.typography.bodySmall.copy(
                                            fontWeight = FontWeight.Bold,
                                            color = MimTheme.colors.textPrimary
                                        )
                                    )
                                    Text(
                                        text = "一键免密",
                                        fontSize = 10.sp,
                                        color = MimTheme.colors.accent
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
