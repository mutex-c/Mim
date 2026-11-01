package com.mim.im.ui.components

import androidx.compose.animation.core.*
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.*
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.mim.im.data.model.CallSessionState
import com.mim.im.repository.ChatRepository
import com.mim.im.ui.theme.MimTheme
import com.mim.im.webrtc.WebRtcManager
import io.getstream.webrtc.android.compose.VideoRenderer
import io.getstream.webrtc.android.compose.VideoScalingType
import org.webrtc.RendererCommon

@Composable
fun ActiveCallOverlay(
    callState: CallSessionState,
    webRtcManager: WebRtcManager? = null,
    chatRepository: ChatRepository? = null,
    onEndCall: () -> Unit
) {
    var isMicMuted by remember { mutableStateOf(false) }
    var isSpeakerOn by remember { mutableStateOf(true) }
    var isCameraOff by remember { mutableStateOf(false) }
    var isSwappedVideo by remember { mutableStateOf(false) }
    var isLocalMirrored by remember { mutableStateOf(true) }

    val localTrack by (webRtcManager?.localVideoTrack?.collectAsState() ?: remember { mutableStateOf(null) })
    val remoteTrack by (webRtcManager?.remoteVideoTrack?.collectAsState() ?: remember { mutableStateOf(null) })
    val isPeerCameraOff by (chatRepository?.isPeerCameraOff?.collectAsState() ?: remember { mutableStateOf(false) })

    val rendererEvents = remember {
        object : RendererCommon.RendererEvents {
            override fun onFirstFrameRendered() {}
            override fun onFrameResolutionChanged(videoWidth: Int, videoHeight: Int, rotation: Int) {}
        }
    }

    val infiniteTransition = rememberInfiniteTransition(label = "pulse")
    val pulseScale by infiniteTransition.animateFloat(
        initialValue = 1f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(
            animation = tween(1200, easing = FastOutSlowInEasing),
            repeatMode = RepeatMode.Reverse
        ),
        label = "scale"
    )

    val minutes = callState.durationSeconds / 60
    val seconds = callState.durationSeconds % 60
    val timeFormatted = String.format("%02d:%02d", minutes, seconds)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(
                        Color(0xFF0F172A),
                        Color(0xFF020617),
                        Color(0xFF000000)
                    )
                )
            )
            .statusBarsPadding()
            .navigationBarsPadding()
    ) {
        // --- Top Bar ---
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 16.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Column {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .clip(CircleShape)
                            .background(if (callState.isConnected) Color(0xFF22C55E) else Color(0xFFF59E0B))
                    )
                    Spacer(modifier = Modifier.width(6.dp))
                    Text(
                        text = if (callState.isVideo) "视频通话" else "语音通话",
                        fontSize = 13.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )
                    Spacer(modifier = Modifier.width(8.dp))
                    Surface(
                        shape = RoundedCornerShape(6.dp),
                        color = Color.White.copy(alpha = 0.1f)
                    ) {
                        Text(
                            text = "端到端加密",
                            fontSize = 10.sp,
                            color = Color(0xFF94A3B8),
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                        )
                    }
                }
                Spacer(modifier = Modifier.height(2.dp))
                Text(
                    text = callState.peerName,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    color = Color.White
                )
            }

            // Duration timer badge
            Surface(
                shape = RoundedCornerShape(20.dp),
                color = Color.White.copy(alpha = 0.12f),
                modifier = Modifier.border(0.5.dp, Color.White.copy(alpha = 0.2f), RoundedCornerShape(20.dp))
            ) {
                Text(
                    text = if (callState.isConnected) timeFormatted else "等待接听...",
                    fontSize = 13.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = if (callState.isConnected) Color.White else Color(0xFFF59E0B),
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp)
                )
            }
        }

        // --- Center Stage ---
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(bottom = 120.dp),
            contentAlignment = Alignment.Center
        ) {
            if (callState.isVideo) {
                // Video Call Stage
                Box(
                    modifier = Modifier
                        .fillMaxWidth(0.92f)
                        .fillMaxHeight(0.78f)
                        .clip(RoundedCornerShape(24.dp))
                        .background(Color(0xFF1E293B))
                        .border(1.dp, Color.White.copy(alpha = 0.1f), RoundedCornerShape(24.dp)),
                    contentAlignment = Alignment.Center
                ) {
                    val rTrack = remoteTrack
                    val lTrack = localTrack

                    // 1. Primary Canvas (Big Stage)
                    if (!isSwappedVideo) {
                        // Default: Show Remote Video
                        if (isPeerCameraOff) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                MimAvatar(name = callState.peerName, size = 96.dp)
                                Spacer(modifier = Modifier.height(16.dp))
                                Text(
                                    text = callState.peerName,
                                    fontSize = 20.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = Color.White
                                )
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(
                                    text = "对方已关闭摄像头",
                                    fontSize = 14.sp,
                                    color = Color(0xFF94A3B8)
                                )
                            }
                        } else if (rTrack != null && webRtcManager != null) {
                            VideoRenderer(
                                videoTrack = rTrack,
                                eglBaseContext = webRtcManager.eglBase.eglBaseContext,
                                videoScalingType = VideoScalingType.SCALE_ASPECT_FILL,
                                modifier = Modifier.fillMaxSize(),
                                rendererEvents = rendererEvents
                            )
                        } else {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                MimAvatar(name = callState.peerName, size = 96.dp)
                                Spacer(modifier = Modifier.height(16.dp))
                                Text(
                                    text = callState.peerName,
                                    fontSize = 20.sp,
                                    fontWeight = FontWeight.Bold,
                                    color = Color.White
                                )
                                Spacer(modifier = Modifier.height(6.dp))
                                Text(
                                    text = if (callState.isConnected) "正在建立 WebRTC 视频流..." else "对方正在连线中...",
                                    fontSize = 12.sp,
                                    color = Color(0xFF94A3B8)
                                )
                            }
                        }
                    } else {
                        // Swapped: Show Local Camera on Big Stage
                        if (isCameraOff) {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                Icon(
                                    imageVector = Icons.Default.VideocamOff,
                                    contentDescription = "摄像头已关闭",
                                    tint = Color(0xFF94A3B8),
                                    modifier = Modifier.size(48.dp)
                                )
                                Spacer(modifier = Modifier.height(12.dp))
                                Text(
                                    text = "本地摄像头已关闭",
                                    fontSize = 16.sp,
                                    color = Color(0xFF94A3B8)
                                )
                            }
                        } else if (lTrack != null && webRtcManager != null) {
                            VideoRenderer(
                                videoTrack = lTrack,
                                eglBaseContext = webRtcManager.eglBase.eglBaseContext,
                                videoScalingType = VideoScalingType.SCALE_ASPECT_FILL,
                                modifier = Modifier
                                    .fillMaxSize()
                                    .graphicsLayer { scaleX = if (isLocalMirrored) -1f else 1f },
                                rendererEvents = rendererEvents
                            )
                        } else {
                            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                CircularProgressIndicator(
                                    modifier = Modifier.size(32.dp),
                                    color = MimTheme.colors.accent,
                                    strokeWidth = 2.5.dp
                                )
                                Spacer(modifier = Modifier.height(10.dp))
                                Text("开启本地摄像头...", fontSize = 13.sp, color = Color.White)
                            }
                        }
                    }

                    // 2. Secondary Floating PiP Card (Top-Right)
                    // Click to swap positions like WeChat!
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(16.dp)
                            .size(width = 110.dp, height = 150.dp)
                            .clip(RoundedCornerShape(16.dp))
                            .background(Color(0xFF0F172A))
                            .border(1.5.dp, Color.White.copy(alpha = 0.45f), RoundedCornerShape(16.dp)),
                        contentAlignment = Alignment.Center
                    ) {
                        if (!isSwappedVideo) {
                            // Default: Small window shows Local Camera
                            if (isCameraOff) {
                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                    Icon(
                                        imageVector = Icons.Default.VideocamOff,
                                        contentDescription = "摄像头已关闭",
                                        tint = Color(0xFF94A3B8),
                                        modifier = Modifier.size(24.dp)
                                    )
                                    Spacer(modifier = Modifier.height(4.dp))
                                    Text(
                                        text = "已关摄像头",
                                        fontSize = 10.sp,
                                        color = Color(0xFF94A3B8)
                                    )
                                }
                            } else if (lTrack != null && webRtcManager != null) {
                                VideoRenderer(
                                    videoTrack = lTrack,
                                    eglBaseContext = webRtcManager.eglBase.eglBaseContext,
                                    videoScalingType = VideoScalingType.SCALE_ASPECT_FILL,
                                    modifier = Modifier
                                        .fillMaxSize()
                                        .graphicsLayer { scaleX = if (isLocalMirrored) -1f else 1f },
                                    rendererEvents = rendererEvents
                                )
                            } else {
                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.size(22.dp),
                                        color = MimTheme.colors.accent,
                                        strokeWidth = 2.dp
                                    )
                                    Spacer(modifier = Modifier.height(6.dp))
                                    Text("开启中...", fontSize = 10.sp, color = Color.White)
                                }
                            }
                        } else {
                            // Swapped: Small window shows Remote Video
                            if (isPeerCameraOff) {
                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                    MimAvatar(name = callState.peerName, size = 44.dp)
                                    Spacer(modifier = Modifier.height(4.dp))
                                    Text(
                                        text = "对方已关头",
                                        fontSize = 10.sp,
                                        color = Color(0xFF94A3B8)
                                    )
                                }
                            } else if (rTrack != null && webRtcManager != null) {
                                VideoRenderer(
                                    videoTrack = rTrack,
                                    eglBaseContext = webRtcManager.eglBase.eglBaseContext,
                                    videoScalingType = VideoScalingType.SCALE_ASPECT_FILL,
                                    modifier = Modifier.fillMaxSize(),
                                    rendererEvents = rendererEvents
                                )
                            } else {
                                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                                    MimAvatar(name = callState.peerName, size = 44.dp)
                                    Spacer(modifier = Modifier.height(6.dp))
                                    Text(
                                        text = callState.peerName.take(4),
                                        fontSize = 10.sp,
                                        fontWeight = FontWeight.Medium,
                                        color = Color.White
                                    )
                                }
                            }
                        }

                        // Touch interceptor to swap views on click
                        Box(
                            modifier = Modifier
                                .fillMaxSize()
                                .clickable { isSwappedVideo = !isSwappedVideo }
                        )

                        // Quick mirror flip button in PiP
                        if (!isSwappedVideo && !isCameraOff) {
                            Surface(
                                shape = RoundedCornerShape(6.dp),
                                color = Color.Black.copy(alpha = 0.65f),
                                modifier = Modifier
                                    .align(Alignment.TopStart)
                                    .padding(6.dp)
                                    .clickable { isLocalMirrored = !isLocalMirrored }
                            ) {
                                Icon(
                                    imageVector = Icons.Default.Flip,
                                    contentDescription = if (isLocalMirrored) "切为原像" else "切为镜像",
                                    tint = if (isLocalMirrored) MimTheme.colors.accent else Color.White,
                                    modifier = Modifier
                                        .padding(4.dp)
                                        .size(14.dp)
                                )
                            }
                        }

                        // Floating swap indicator badge
                        Surface(
                            shape = RoundedCornerShape(4.dp),
                            color = Color.Black.copy(alpha = 0.65f),
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .padding(bottom = 6.dp)
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                                verticalAlignment = Alignment.CenterVertically
                            ) {
                                Icon(
                                    imageVector = Icons.Default.SwapHoriz,
                                    contentDescription = "点击互换画面",
                                    tint = MimTheme.colors.accent,
                                    modifier = Modifier.size(12.dp)
                                )
                                Spacer(modifier = Modifier.width(3.dp))
                                Text(
                                    text = if (isSwappedVideo) "对方" else (if (isLocalMirrored) "自己(镜像)" else "自己(原像)"),
                                    fontSize = 9.sp,
                                    fontWeight = FontWeight.Medium,
                                    color = Color.White
                                )
                            }
                        }
                    }
                }
            } else {
                // Audio Call Stage
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Box(
                        modifier = Modifier.size(160.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        if (callState.isConnected) {
                            Box(
                                modifier = Modifier
                                    .size((140 * pulseScale).dp)
                                    .clip(CircleShape)
                                    .background(MimTheme.colors.accent.copy(alpha = 0.15f))
                            )
                        }
                        MimAvatar(name = callState.peerName, size = 110.dp)
                    }

                    Spacer(modifier = Modifier.height(20.dp))

                    Text(
                        text = callState.peerName,
                        fontSize = 24.sp,
                        fontWeight = FontWeight.Bold,
                        color = Color.White
                    )

                    Spacer(modifier = Modifier.height(8.dp))

                    Text(
                        text = if (callState.isConnected) "正在高质量语音通话" else "等待对方接通...",
                        fontSize = 14.sp,
                        color = Color(0xFF94A3B8)
                    )

                    Spacer(modifier = Modifier.height(24.dp))

                    // 8-Band Jumping Audio Equalizer Waveform
                    Row(
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        val heights = listOf(14, 24, 38, 20, 32, 16, 28, 12)
                        heights.forEachIndexed { i, h ->
                            val animH by infiniteTransition.animateFloat(
                                initialValue = (h * 0.4f),
                                targetValue = (h * 1.0f),
                                animationSpec = infiniteRepeatable(
                                    animation = tween(400 + i * 80, easing = LinearEasing),
                                    repeatMode = RepeatMode.Reverse
                                ),
                                label = "wave_$i"
                            )
                            Box(
                                modifier = Modifier
                                    .width(4.dp)
                                    .height(if (callState.isConnected) animH.dp else 6.dp)
                                    .clip(RoundedCornerShape(2.dp))
                                    .background(if (callState.isConnected) MimTheme.colors.accent else Color.White.copy(alpha = 0.3f))
                            )
                        }
                    }
                }
            }
        }

        // --- Bottom Control Dock ---
        Box(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .padding(horizontal = 24.dp, vertical = 28.dp)
        ) {
            Surface(
                shape = RoundedCornerShape(28.dp),
                color = Color(0xFF1E293B).copy(alpha = 0.9f),
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, Color.White.copy(alpha = 0.15f), RoundedCornerShape(28.dp))
            ) {
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 14.dp),
                    horizontalArrangement = Arrangement.SpaceEvenly,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    // 1. Mic Mute Toggle
                    ControlDockButton(
                        icon = if (isMicMuted) Icons.Default.MicOff else Icons.Default.Mic,
                        label = if (isMicMuted) "已静音" else "麦克风",
                        isActive = isMicMuted,
                        activeColor = Color(0xFFEF4444),
                        onClick = {
                            isMicMuted = !isMicMuted
                            webRtcManager?.toggleMic(isMicMuted)
                        }
                    )

                    // 2. Speaker Toggle
                    ControlDockButton(
                        icon = if (isSpeakerOn) Icons.AutoMirrored.Filled.VolumeUp else Icons.AutoMirrored.Filled.VolumeOff,
                        label = if (isSpeakerOn) "扬声器" else "听筒",
                        isActive = isSpeakerOn,
                        activeColor = MimTheme.colors.accent,
                        onClick = {
                            isSpeakerOn = !isSpeakerOn
                            webRtcManager?.toggleSpeaker(isSpeakerOn)
                        }
                    )

                    // 3. Camera Toggle (only in video call)
                    if (callState.isVideo) {
                        ControlDockButton(
                            icon = if (isCameraOff) Icons.Default.VideocamOff else Icons.Default.Videocam,
                            label = if (isCameraOff) "已关头" else "摄像头",
                            isActive = isCameraOff,
                            activeColor = Color(0xFFEF4444),
                            onClick = {
                                isCameraOff = !isCameraOff
                                if (chatRepository != null) {
                                    chatRepository.toggleCamera(isCameraOff)
                                } else {
                                    webRtcManager?.toggleCamera(isCameraOff)
                                }
                            }
                        )

                        // 4. Mirror Toggle (镜像/原像)
                        ControlDockButton(
                            icon = Icons.Default.Flip,
                            label = if (isLocalMirrored) "镜像" else "原像",
                            isActive = isLocalMirrored,
                            activeColor = MimTheme.colors.accent,
                            onClick = { isLocalMirrored = !isLocalMirrored }
                        )

                        // 5. Flip Camera (前后置镜头翻转)
                        ControlDockButton(
                            icon = Icons.Default.FlipCameraAndroid,
                            label = "翻转",
                            isActive = false,
                            activeColor = MimTheme.colors.accent,
                            onClick = { webRtcManager?.switchCamera() }
                        )
                    }

                    // 6. Hang up (End Call) Button
                    IconButton(
                        onClick = onEndCall,
                        modifier = Modifier
                            .size(54.dp)
                            .clip(CircleShape)
                            .background(Color(0xFFDC2626))
                    ) {
                        Icon(
                            imageVector = Icons.Default.CallEnd,
                            contentDescription = "挂断",
                            tint = Color.White,
                            modifier = Modifier.size(28.dp)
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun ControlDockButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    isActive: Boolean,
    activeColor: Color,
    onClick: () -> Unit
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.clickable(onClick = onClick)
    ) {
        Box(
            modifier = Modifier
                .size(46.dp)
                .clip(CircleShape)
                .background(if (isActive) activeColor.copy(alpha = 0.2f) else Color.White.copy(alpha = 0.1f))
                .border(1.dp, if (isActive) activeColor else Color.White.copy(alpha = 0.15f), CircleShape),
            contentAlignment = Alignment.Center
        ) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                tint = if (isActive) activeColor else Color.White,
                modifier = Modifier.size(22.dp)
            )
        }
        Spacer(modifier = Modifier.height(4.dp))
        Text(
            text = label,
            fontSize = 11.sp,
            color = if (isActive) activeColor else Color(0xFF94A3B8)
        )
    }
}
