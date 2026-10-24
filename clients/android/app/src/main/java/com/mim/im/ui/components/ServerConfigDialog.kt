package com.mim.im.ui.components

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun ServerConfigDialog(
    currentUrl: String,
    onDismiss: () -> Unit,
    onConfirm: (String) -> Unit
) {
    var urlText by remember { mutableStateOf(currentUrl) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text(text = "服务器网络配置") },
        text = {
            Column(modifier = Modifier.fillMaxWidth()) {
                Text(
                    text = "请输入 Mim 后端服务地址（模拟器请使用 10.0.2.2:8080，真机请使用局域网 IP 如 192.168.x.x:8080）：",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(12.dp))
                OutlinedTextField(
                    value = urlText,
                    onValueChange = { urlText = it },
                    label = { Text("服务地址 (HTTP/HTTPS)") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )
                Spacer(modifier = Modifier.height(8.dp))
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    OutlinedButton(
                        onClick = { urlText = "http://10.0.2.2:8080" },
                        modifier = Modifier.weight(1f)
                    ) {
                        Text("模拟器默认")
                    }
                    OutlinedButton(
                        onClick = { urlText = "http://127.0.0.1:8080" },
                        modifier = Modifier.weight(1f)
                    ) {
                        Text("Localhost")
                    }
                }
            }
        },
        confirmButton = {
            Button(
                onClick = {
                    val formatted = if (!urlText.startsWith("http://") && !urlText.startsWith("https://")) {
                        "http://$urlText"
                    } else {
                        urlText
                    }
                    onConfirm(formatted.trimEnd('/'))
                }
            ) {
                Text("保存并生效")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("取消")
            }
        }
    )
}
