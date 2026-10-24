package com.mim.im.ui.screens

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
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
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.local.PreferencesHelper
import com.mim.im.network.WsManager
import com.mim.im.network.WsStatus
import com.mim.im.ui.components.MimAvatar
import com.mim.im.ui.components.ServerConfigDialog
import com.mim.im.ui.theme.ColorPalette
import com.mim.im.ui.theme.MimTheme
import com.mim.im.ui.theme.ThemeMode

@Composable
fun SettingsScreen(
    prefs: PreferencesHelper,
    wsManager: WsManager,
    onLogout: () -> Unit,
    modifier: Modifier = Modifier
) {
    val scrollState = rememberScrollState()
    var showServerDialog by remember { mutableStateOf(false) }

    val currentPalette by prefs.paletteFlow.collectAsState()
    val currentThemeMode by prefs.themeModeFlow.collectAsState()
    val wsStatus by wsManager.status.collectAsState()
    val user = prefs.currentUser

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
        // --- Profile Card ---
        Surface(
            shape = RoundedCornerShape(16.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(16.dp))
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                MimAvatar(
                    name = user?.nickname ?: user?.username ?: "Mim",
                    size = 56.dp,
                    imageUrl = user?.avatar,
                    isOnline = wsStatus == WsStatus.CONNECTED
                )
                Spacer(modifier = Modifier.width(16.dp))
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = user?.nickname ?: user?.username ?: "未知用户",
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
            }
        }

        Spacer(modifier = Modifier.height(20.dp))

        // --- Theme Mode Section ---
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

        // --- Color Palettes Section ---
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

        // --- Network & Server Section ---
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
                    border = ButtonDefaults.outlinedButtonBorder.copy(
                        brush = androidx.compose.ui.graphics.SolidColor(MimTheme.colors.border)
                    ),
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

        // --- Logout Button ---
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
                horizontalArrangement = Arrangement.Center,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(
                    imageVector = Icons.AutoMirrored.Filled.ExitToApp,
                    contentDescription = "退出登录",
                    tint = Color(0xFFDC2626),
                    modifier = Modifier.size(18.dp)
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "退出当前账号",
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontWeight = FontWeight.Bold,
                        color = Color(0xFFDC2626)
                    )
                )
            }
        }

        Spacer(modifier = Modifier.height(32.dp))
    }
}
