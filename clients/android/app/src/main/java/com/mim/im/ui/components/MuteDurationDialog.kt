package com.mim.im.ui.components

import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import com.mim.im.ui.theme.MimTheme

@Composable
fun MuteDurationDialog(
    memberName: String,
    onDismiss: () -> Unit,
    onSelectDuration: (durationSeconds: Int) -> Unit
) {
    val durations = listOf(
        Pair("10 分钟", 10 * 60),
        Pair("1 小时", 60 * 60),
        Pair("24 小时 (1 天)", 24 * 60 * 60),
        Pair("解除禁言", 0)
    )

    Dialog(onDismissRequest = onDismiss) {
        Surface(
            shape = RoundedCornerShape(16.dp),
            color = MimTheme.colors.surfaceElevated,
            modifier = Modifier
                .fillMaxWidth()
                .border(1.dp, MimTheme.colors.border, RoundedCornerShape(16.dp))
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(20.dp)
            ) {
                Text(
                    text = "设置禁言: $memberName",
                    style = MaterialTheme.typography.titleMedium.copy(
                        fontWeight = FontWeight.Bold,
                        color = MimTheme.colors.textPrimary
                    )
                )

                Spacer(modifier = Modifier.height(14.dp))

                durations.forEach { (label, secs) ->
                    val isUnmute = (secs == 0)
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clip(RoundedCornerShape(8.dp))
                            .clickable {
                                onSelectDuration(secs)
                            }
                            .padding(vertical = 12.dp, horizontal = 8.dp)
                    ) {
                        Text(
                            text = label,
                            style = MaterialTheme.typography.bodyMedium.copy(
                                fontWeight = FontWeight.Medium,
                                color = if (isUnmute) Color(0xFF22C55E) else MimTheme.colors.textPrimary
                            )
                        )
                    }
                    HorizontalDivider(color = MimTheme.colors.border.copy(alpha = 0.5f))
                }

                Spacer(modifier = Modifier.height(8.dp))

                OutlinedButton(
                    onClick = onDismiss,
                    shape = RoundedCornerShape(10.dp),
                    modifier = Modifier.fillMaxWidth(),
                    border = androidx.compose.foundation.BorderStroke(0.5.dp, MimTheme.colors.border)
                ) {
                    Text("取消", color = MimTheme.colors.textSecondary)
                }
            }
        }
    }
}
