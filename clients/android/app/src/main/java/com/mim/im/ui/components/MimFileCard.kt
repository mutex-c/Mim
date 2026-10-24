package com.mim.im.ui.components

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.ui.theme.MimTheme
import java.io.File

data class FileTypeStyle(
    val badgeColor: Color,
    val label: String
)

fun getFileTypeStyle(fileName: String): FileTypeStyle {
    val ext = fileName.substringAfterLast('.', "").lowercase()
    return when (ext) {
        "pdf" -> FileTypeStyle(Color(0xFFDC2626), "PDF")
        "doc", "docx" -> FileTypeStyle(Color(0xFF2563EB), "DOC")
        "xls", "xlsx" -> FileTypeStyle(Color(0xFF059669), "XLS")
        "ppt", "pptx" -> FileTypeStyle(Color(0xFFEA580C), "PPT")
        "zip", "rar", "7z", "tar", "gz" -> FileTypeStyle(Color(0xFFD97706), "ZIP")
        "txt", "md", "json", "go", "kt", "ts", "rs" -> FileTypeStyle(Color(0xFF0D9488), ext.uppercase())
        else -> FileTypeStyle(Color(0xFF64748B), if (ext.isNotEmpty()) ext.take(4).uppercase() else "FILE")
    }
}

@Composable
fun MimFileCard(
    fileName: String,
    fileSize: Long,
    isDownloaded: Boolean,
    isDownloading: Boolean,
    downloadProgress: Float = 0f,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    val style = getFileTypeStyle(fileName)
    val sizeText = when {
        fileSize < 1024 -> "$fileSize B"
        fileSize < 1024 * 1024 -> String.format("%.1f KB", fileSize / 1024.0)
        else -> String.format("%.1f MB", fileSize / (1024.0 * 1024.0))
    }

    Surface(
        shape = RoundedCornerShape(12.dp),
        color = MimTheme.colors.surfaceElevated,
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .border(1.dp, MimTheme.colors.border, RoundedCornerShape(12.dp))
            .clickable(onClick = onClick)
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            // File Icon Box
            Box(
                modifier = Modifier
                    .size(46.dp)
                    .clip(RoundedCornerShape(10.dp))
                    .background(style.badgeColor.copy(alpha = 0.12f)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = style.label,
                    color = style.badgeColor,
                    fontWeight = FontWeight.Black,
                    fontSize = 12.sp
                )
            }

            Spacer(modifier = Modifier.width(12.dp))

            // File Name & Info
            Column(
                modifier = Modifier.weight(1f)
            ) {
                Text(
                    text = fileName,
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontWeight = FontWeight.SemiBold,
                        color = MimTheme.colors.textPrimary
                    ),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
                Spacer(modifier = Modifier.height(3.dp))
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(
                        text = sizeText,
                        style = MaterialTheme.typography.bodySmall.copy(
                            color = MimTheme.colors.textMuted
                        )
                    )
                    if (isDownloaded) {
                        Spacer(modifier = Modifier.width(6.dp))
                        Text(
                            text = "· 已下载",
                            style = MaterialTheme.typography.bodySmall.copy(
                                color = Color(0xFF16A34A),
                                fontWeight = FontWeight.Medium
                            )
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.width(8.dp))

            // Action / Status Indicator
            Box(
                modifier = Modifier
                    .size(34.dp)
                    .clip(CircleShape)
                    .background(
                        if (isDownloaded) MimTheme.colors.accentSubtle
                        else MimTheme.colors.surface
                    ),
                contentAlignment = Alignment.Center
            ) {
                when {
                    isDownloading -> {
                        CircularProgressIndicator(
                            progress = { downloadProgress },
                            modifier = Modifier.size(20.dp),
                            strokeWidth = 2.dp,
                            color = MimTheme.colors.accent
                        )
                    }
                    isDownloaded -> {
                        Icon(
                            imageVector = Icons.Default.FolderOpen,
                            contentDescription = "打开",
                            tint = MimTheme.colors.accent,
                            modifier = Modifier.size(18.dp)
                        )
                    }
                    else -> {
                        Icon(
                            imageVector = Icons.Default.Download,
                            contentDescription = "下载",
                            tint = MimTheme.colors.textMuted,
                            modifier = Modifier.size(18.dp)
                        )
                    }
                }
            }
        }
    }
}
