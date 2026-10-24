package com.mim.im.ui.components

import android.Manifest
import android.content.pm.PackageManager
import android.widget.Toast
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.scaleIn
import androidx.compose.animation.scaleOut
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Mic
import androidx.compose.material.icons.filled.MicOff
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.scale
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.core.content.ContextCompat
import com.mim.im.ui.theme.MimTheme
import com.mim.im.utils.AudioRecorderHelper
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import java.io.File

@Composable
fun VoiceRecordBar(
    recorderHelper: AudioRecorderHelper,
    onVoiceRecorded: (file: File, durationMs: Long) -> Unit,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    var isRecording by remember { mutableStateOf(false) }
    var isWillCancel by remember { mutableStateOf(false) }
    var recordSeconds by remember { mutableStateOf(0) }

    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        if (!isGranted) {
            Toast.makeText(context, "需要麦克风权限才能发送语音", Toast.LENGTH_SHORT).show()
        }
    }

    // Timer effect while recording
    LaunchedEffect(isRecording) {
        if (isRecording) {
            recordSeconds = 0
            while (isActive && isRecording) {
                delay(1000)
                recordSeconds++
            }
        }
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .height(44.dp)
            .clip(RoundedCornerShape(22.dp))
            .background(
                if (isRecording) {
                    if (isWillCancel) Color(0xFFEF4444).copy(alpha = 0.15f)
                    else MimTheme.colors.primary.copy(alpha = 0.15f)
                } else {
                    MimTheme.colors.cardBackground
                }
            )
            .pointerInput(Unit) {
                awaitEachGesture {
                    val down = awaitFirstDown(requireUnconsumed = false)
                    
                    // Check permission
                    val hasPerm = ContextCompat.checkSelfPermission(
                        context,
                        Manifest.permission.RECORD_AUDIO
                    ) == PackageManager.PERMISSION_GRANTED

                    if (!hasPerm) {
                        permissionLauncher.launch(Manifest.permission.RECORD_AUDIO)
                        return@awaitEachGesture
                    }

                    // Start recording
                    val started = recorderHelper.startRecording()
                    if (!started) {
                        Toast.makeText(context, "无法启动录音", Toast.LENGTH_SHORT).show()
                        return@awaitEachGesture
                    }

                    isRecording = true
                    isWillCancel = false

                    val pointerId = down.id
                    while (true) {
                        val event = awaitPointerEvent()
                        val change = event.changes.firstOrNull { it.id == pointerId }
                        if (change == null || !change.pressed) {
                            // Released!
                            break
                        }

                        // Drag up beyond threshold (e.g. -120px) cancels
                        val currentY = change.position.y
                        isWillCancel = (currentY < -120f)
                    }

                    // Done recording
                    isRecording = false
                    if (isWillCancel) {
                        recorderHelper.cancelRecording()
                        Toast.makeText(context, "已取消发送", Toast.LENGTH_SHORT).show()
                    } else {
                        val res = recorderHelper.stopRecording()
                        if (res == null) {
                            Toast.makeText(context, "说话时间太短", Toast.LENGTH_SHORT).show()
                        } else {
                            onVoiceRecorded(res.first, res.second)
                        }
                    }
                }
            },
        contentAlignment = Alignment.Center
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Icon(
                imageVector = if (isWillCancel) Icons.Default.MicOff else Icons.Default.Mic,
                contentDescription = null,
                tint = if (isWillCancel) Color(0xFFEF4444) else MimTheme.colors.primary,
                modifier = Modifier.size(18.dp)
            )
            Text(
                text = if (isRecording) {
                    if (isWillCancel) "松开手指，取消发送"
                    else "松开发送 (${String.format("%02d:%02d", recordSeconds / 60, recordSeconds % 60)})"
                } else {
                    "按住 说话"
                },
                style = MaterialTheme.typography.bodyMedium.copy(
                    fontWeight = FontWeight.Medium,
                    color = if (isWillCancel) Color(0xFFEF4444)
                    else if (isRecording) MimTheme.colors.primary
                    else MimTheme.colors.textSecondary
                )
            )
        }
    }
}
