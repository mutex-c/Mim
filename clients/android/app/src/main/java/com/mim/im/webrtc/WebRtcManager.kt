package com.mim.im.webrtc

import android.content.Context
import android.media.AudioManager
import android.util.Log
import com.mim.im.data.model.WebRTCSignalDto
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONObject
import org.webrtc.*

open class SimpleSdpObserver : SdpObserver {
    override fun onCreateSuccess(desc: SessionDescription) {}
    override fun onSetSuccess() {}
    override fun onCreateFailure(error: String?) {}
    override fun onSetFailure(error: String?) {}
}

class WebRtcManager(private val context: Context) {
    companion object {
        private const val TAG = "WebRtcManager"
        private val ICE_SERVERS = listOf(
            PeerConnection.IceServer.builder("stun:stun.qq.com:3478").createIceServer(),
            PeerConnection.IceServer.builder("stun:stun.miwifi.com:3478").createIceServer(),
            PeerConnection.IceServer.builder("stun:stun.cloudflare.com:3478").createIceServer(),
            PeerConnection.IceServer.builder("stun:stun.l.google.com:19302").createIceServer(),
            PeerConnection.IceServer.builder("stun:stun1.l.google.com:19302").createIceServer()
        )
    }

    val eglBase: EglBase by lazy { EglBase.create() }

    private var factory: PeerConnectionFactory? = null
    private var peerConnection: PeerConnection? = null

    private var audioSource: AudioSource? = null
    private var localAudioTrack: AudioTrack? = null

    private var videoCapturer: CameraVideoCapturer? = null
    private var surfaceTextureHelper: SurfaceTextureHelper? = null
    private var videoSource: VideoSource? = null
    private var localVideoTrackInstance: VideoTrack? = null

    private val _localVideoTrack = MutableStateFlow<VideoTrack?>(null)
    val localVideoTrack: StateFlow<VideoTrack?> = _localVideoTrack.asStateFlow()

    private val _remoteVideoTrack = MutableStateFlow<VideoTrack?>(null)
    val remoteVideoTrack: StateFlow<VideoTrack?> = _remoteVideoTrack.asStateFlow()

    private val pendingIceCandidates = mutableListOf<IceCandidate>()
    private val audioManager = context.getSystemService(Context.AUDIO_SERVICE) as AudioManager
    private val scope = CoroutineScope(Dispatchers.Main)

    init {
        initPeerConnectionFactory()
    }

    private fun initPeerConnectionFactory() {
        try {
            val options = PeerConnectionFactory.InitializationOptions.builder(context)
                .setEnableInternalTracer(false)
                .createInitializationOptions()
            PeerConnectionFactory.initialize(options)

            val encoderFactory = DefaultVideoEncoderFactory(
                eglBase.eglBaseContext,
                true,
                true
            )
            val decoderFactory = DefaultVideoDecoderFactory(eglBase.eglBaseContext)

            factory = PeerConnectionFactory.builder()
                .setVideoEncoderFactory(encoderFactory)
                .setVideoDecoderFactory(decoderFactory)
                .setOptions(PeerConnectionFactory.Options())
                .createPeerConnectionFactory()
        } catch (e: Exception) {
            Log.e(TAG, "initPeerConnectionFactory failed", e)
        }
    }

    private fun createCameraCapturer(): CameraVideoCapturer? {
        val enumerator: CameraEnumerator = if (Camera2Enumerator.isSupported(context)) {
            Camera2Enumerator(context)
        } else {
            Camera1Enumerator(true)
        }

        val deviceNames = enumerator.deviceNames
        // Prioritize front camera
        for (deviceName in deviceNames) {
            if (enumerator.isFrontFacing(deviceName)) {
                val capturer = enumerator.createCapturer(deviceName, null)
                if (capturer != null) return capturer
            }
        }
        // Fallback to back camera
        for (deviceName in deviceNames) {
            if (!enumerator.isFrontFacing(deviceName)) {
                val capturer = enumerator.createCapturer(deviceName, null)
                if (capturer != null) return capturer
            }
        }
        return null
    }

    fun startLocalMedia(isVideo: Boolean) {
        val fac = factory ?: return

        // 1. Audio Track
        if (localAudioTrack == null) {
            val audioConstraints = MediaConstraints()
            audioSource = fac.createAudioSource(audioConstraints)
            localAudioTrack = fac.createAudioTrack("mim_audio_track", audioSource)
            localAudioTrack?.setEnabled(true)
        }

        // 2. Video Track
        if (isVideo && localVideoTrackInstance == null) {
            try {
                val capturer = createCameraCapturer()
                if (capturer != null) {
                    videoCapturer = capturer
                    surfaceTextureHelper = SurfaceTextureHelper.create("MimCaptureThread", eglBase.eglBaseContext)
                    val source = fac.createVideoSource(capturer.isScreencast)
                    videoSource = source
                    capturer.initialize(surfaceTextureHelper, context, source.capturerObserver)
                    val vTrack = fac.createVideoTrack("mim_video_track", source)
                    vTrack.setEnabled(true)
                    localVideoTrackInstance = vTrack
                    _localVideoTrack.value = vTrack

                    try {
                        capturer.startCapture(1280, 720, 30)
                    } catch (e: Exception) {
                        Log.w(TAG, "720p capture failed, falling back to 640x480", e)
                        capturer.startCapture(640, 480, 30)
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "startLocalMedia video capture failed", e)
            }
        }

        setAudioMode(inCall = true, speakerOn = isVideo)
    }

    private fun createPeerConnection(onSignal: (signalType: String, data: String) -> Unit): PeerConnection? {
        val fac = factory ?: return null
        val rtcConfig = PeerConnection.RTCConfiguration(ICE_SERVERS).apply {
            sdpSemantics = PeerConnection.SdpSemantics.UNIFIED_PLAN
            continualGatheringPolicy = PeerConnection.ContinualGatheringPolicy.GATHER_CONTINUALLY
        }

        val pc = fac.createPeerConnection(rtcConfig, object : PeerConnection.Observer {
            override fun onSignalingChange(state: PeerConnection.SignalingState) {
                Log.d(TAG, "onSignalingChange: $state")
            }

            override fun onIceConnectionChange(state: PeerConnection.IceConnectionState) {
                Log.d(TAG, "onIceConnectionChange: $state")
                if (state == PeerConnection.IceConnectionState.CONNECTED || state == PeerConnection.IceConnectionState.COMPLETED) {
                    configureHighQualityBitrate()
                }
            }

            override fun onIceConnectionReceivingChange(receiving: Boolean) {}

            override fun onIceGatheringChange(state: PeerConnection.IceGatheringState) {
                Log.d(TAG, "onIceGatheringChange: $state")
            }

            override fun onIceCandidate(candidate: IceCandidate) {
                Log.d(TAG, "onIceCandidate: ${candidate.sdpMid}")
                val candObj = JSONObject().apply {
                    put("candidate", candidate.sdp)
                    put("sdpMid", candidate.sdpMid)
                    put("sdpMLineIndex", candidate.sdpMLineIndex)
                }
                onSignal("CANDIDATE", candObj.toString())
            }

            override fun onIceCandidatesRemoved(candidates: Array<out IceCandidate>?) {}

            override fun onAddStream(stream: MediaStream) {}

            override fun onRemoveStream(stream: MediaStream) {}

            override fun onDataChannel(dataChannel: DataChannel) {}

            override fun onRenegotiationNeeded() {}

            override fun onAddTrack(receiver: RtpReceiver, mediaStreams: Array<out MediaStream>) {
                val track = receiver.track()
                if (track is VideoTrack) {
                    scope.launch {
                        _remoteVideoTrack.value = track
                    }
                }
            }

            override fun onTrack(transceiver: RtpTransceiver) {
                val track = transceiver.receiver.track()
                if (track is VideoTrack) {
                    scope.launch {
                        _remoteVideoTrack.value = track
                    }
                }
            }
        })

        localAudioTrack?.let {
            pc?.addTrack(it, listOf("mim_stream"))
        }
        localVideoTrackInstance?.let {
            pc?.addTrack(it, listOf("mim_stream"))
        }

        return pc
    }

    fun initiateCall(
        callId: String,
        targetUserId: Long,
        isVideo: Boolean,
        onSendSignal: (signalType: String, toUserId: Long, callId: String, data: String) -> Unit
    ) {
        startLocalMedia(isVideo)

        peerConnection = createPeerConnection { type, data ->
            onSendSignal(type, targetUserId, callId, data)
        }

        val constraints = MediaConstraints().apply {
            mandatory.add(MediaConstraints.KeyValuePair("OfferToReceiveAudio", "true"))
            if (isVideo) {
                mandatory.add(MediaConstraints.KeyValuePair("OfferToReceiveVideo", "true"))
                mandatory.add(MediaConstraints.KeyValuePair("maxWidth", "1920"))
                mandatory.add(MediaConstraints.KeyValuePair("maxHeight", "1080"))
                mandatory.add(MediaConstraints.KeyValuePair("maxFrameRate", "30"))
            }
        }

        peerConnection?.createOffer(object : SimpleSdpObserver() {
            override fun onCreateSuccess(desc: SessionDescription) {
                val hdSdp = enhanceSdpBitrate(desc.description)
                val newDesc = SessionDescription(desc.type, hdSdp)
                peerConnection?.setLocalDescription(SimpleSdpObserver(), newDesc)
                configureHighQualityBitrate()

                val offerJson = JSONObject().apply {
                    val sdpObj = JSONObject().apply {
                        put("type", "offer")
                        put("sdp", hdSdp)
                    }
                    put("sdp", sdpObj)
                    put("callType", if (isVideo) "video" else "audio")
                }
                onSendSignal("OFFER", targetUserId, callId, offerJson.toString())
            }

            override fun onCreateFailure(error: String?) {
                Log.e(TAG, "createOffer failed: $error")
            }
        }, constraints)
    }

    fun handleOffer(
        signal: WebRTCSignalDto,
        onSendSignal: (signalType: String, toUserId: Long, callId: String, data: String) -> Unit
    ) {
        val isVideo = signal.data.contains("\"callType\":\"video\"") || signal.data.contains("video")
        startLocalMedia(isVideo)

        peerConnection = createPeerConnection { type, data ->
            onSendSignal(type, signal.fromUserId, signal.callId, data)
        }

        val sdpString = parseSdpString(signal.data)
        if (sdpString != null) {
            val remoteDesc = SessionDescription(SessionDescription.Type.OFFER, sdpString)
            peerConnection?.setRemoteDescription(object : SimpleSdpObserver() {
                override fun onSetSuccess() {
                    flushPendingCandidates()
                    val constraints = MediaConstraints().apply {
                        mandatory.add(MediaConstraints.KeyValuePair("OfferToReceiveAudio", "true"))
                        if (isVideo) {
                            mandatory.add(MediaConstraints.KeyValuePair("OfferToReceiveVideo", "true"))
                            mandatory.add(MediaConstraints.KeyValuePair("maxWidth", "1920"))
                            mandatory.add(MediaConstraints.KeyValuePair("maxHeight", "1080"))
                            mandatory.add(MediaConstraints.KeyValuePair("maxFrameRate", "30"))
                        }
                    }
                    peerConnection?.createAnswer(object : SimpleSdpObserver() {
                        override fun onCreateSuccess(desc: SessionDescription) {
                            val hdSdp = enhanceSdpBitrate(desc.description)
                            val newDesc = SessionDescription(desc.type, hdSdp)
                            peerConnection?.setLocalDescription(SimpleSdpObserver(), newDesc)
                            configureHighQualityBitrate()

                            val answerJson = JSONObject().apply {
                                val sdpObj = JSONObject().apply {
                                    put("type", "answer")
                                    put("sdp", hdSdp)
                                }
                                put("sdp", sdpObj)
                            }
                            onSendSignal("ANSWER", signal.fromUserId, signal.callId, answerJson.toString())
                        }

                        override fun onCreateFailure(error: String?) {
                            Log.e(TAG, "createAnswer failed: $error")
                        }
                    }, constraints)
                }

                override fun onSetFailure(error: String?) {
                    Log.e(TAG, "setRemoteDescription for OFFER failed: $error")
                }
            }, remoteDesc)
        }
    }

    fun handleAnswer(signal: WebRTCSignalDto) {
        val sdpString = parseSdpString(signal.data) ?: return
        val remoteDesc = SessionDescription(SessionDescription.Type.ANSWER, sdpString)
        peerConnection?.setRemoteDescription(object : SimpleSdpObserver() {
            override fun onSetSuccess() {
                flushPendingCandidates()
                configureHighQualityBitrate()
            }

            override fun onSetFailure(error: String?) {
                Log.e(TAG, "setRemoteDescription for ANSWER failed: $error")
            }
        }, remoteDesc)
    }

    fun handleCandidate(signal: WebRTCSignalDto) {
        try {
            val json = JSONObject(signal.data)
            val sdpMid = json.optString("sdpMid", "")
            val sdpMLineIndex = json.optInt("sdpMLineIndex", 0)
            val candidateStr = json.optString("candidate", "")
            if (candidateStr.isNotBlank()) {
                val candidate = IceCandidate(sdpMid, sdpMLineIndex, candidateStr)
                val pc = peerConnection
                if (pc != null && pc.remoteDescription != null) {
                    pc.addIceCandidate(candidate)
                } else {
                    pendingIceCandidates.add(candidate)
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "handleCandidate parse error", e)
        }
    }

    private fun flushPendingCandidates() {
        val pc = peerConnection ?: return
        if (pc.remoteDescription == null) return
        val iterator = pendingIceCandidates.iterator()
        while (iterator.hasNext()) {
            val cand = iterator.next()
            pc.addIceCandidate(cand)
            iterator.remove()
        }
    }

    private fun parseSdpString(data: String): String? {
        if (data.isBlank()) return null
        if (data.startsWith("v=0")) return data
        return try {
            val obj = JSONObject(data)
            if (obj.has("sdp")) {
                val sdpVal = obj.get("sdp")
                if (sdpVal is JSONObject) {
                    sdpVal.optString("sdp")
                } else {
                    sdpVal.toString()
                }
            } else {
                obj.optString("sdp", "")
            }
        } catch (_: Exception) {
            null
        }
    }

    fun toggleMic(muted: Boolean) {
        localAudioTrack?.setEnabled(!muted)
    }

    fun toggleCamera(cameraOff: Boolean) {
        localVideoTrackInstance?.setEnabled(!cameraOff)
    }

    fun switchCamera() {
        try {
            videoCapturer?.switchCamera(null)
        } catch (e: Exception) {
            Log.e(TAG, "switchCamera failed", e)
        }
    }

    private fun configureHighQualityBitrate() {
        try {
            val senders = peerConnection?.senders ?: return
            for (sender in senders) {
                if (sender.track()?.kind() == "video") {
                    val params = sender.parameters
                    if (params.encodings != null && params.encodings.isNotEmpty()) {
                        for (encoding in params.encodings) {
                            encoding.minBitrateBps = 1500 * 1000   // 1.5 Mbps minimum
                            encoding.maxBitrateBps = 4000 * 1000   // 4.0 Mbps maximum for crisp HD
                            encoding.maxFramerate = 30
                            encoding.scaleResolutionDownBy = 1.0
                        }
                        sender.parameters = params
                    }
                }
            }
        } catch (e: Exception) {
            Log.e(TAG, "configureHighQualityBitrate failed", e)
        }
    }

    private fun enhanceSdpBitrate(sdp: String): String {
        val lines = sdp.split("\r\n")
        val result = mutableListOf<String>()
        var inVideo = false
        var addedBandwidth = false

        for (line in lines) {
            if (line.startsWith("m=video")) {
                inVideo = true
            } else if (line.startsWith("m=audio") || line.startsWith("m=application")) {
                inVideo = false
            }

            result.add(line)

            if (inVideo && !addedBandwidth && (line.startsWith("c=IN") || line.startsWith("a=rtcp-mux"))) {
                result.add("b=AS:4000")
                result.add("b=TIAS:4000000")
                addedBandwidth = true
            }
        }
        return result.joinToString("\r\n")
    }

    fun toggleSpeaker(speakerOn: Boolean) {
        setAudioMode(inCall = true, speakerOn = speakerOn)
    }

    fun endCall() {
        try {
            videoCapturer?.stopCapture()
            videoCapturer?.dispose()
        } catch (_: Exception) {}
        videoCapturer = null

        try {
            surfaceTextureHelper?.dispose()
        } catch (_: Exception) {}
        surfaceTextureHelper = null

        localVideoTrackInstance?.dispose()
        localVideoTrackInstance = null
        _localVideoTrack.value = null

        videoSource?.dispose()
        videoSource = null

        localAudioTrack?.dispose()
        localAudioTrack = null

        audioSource?.dispose()
        audioSource = null

        try {
            peerConnection?.close()
            peerConnection?.dispose()
        } catch (_: Exception) {}
        peerConnection = null

        _remoteVideoTrack.value = null
        pendingIceCandidates.clear()

        setAudioMode(inCall = false, speakerOn = false)
    }

    private fun setAudioMode(inCall: Boolean, speakerOn: Boolean) {
        try {
            if (inCall) {
                audioManager.mode = AudioManager.MODE_IN_COMMUNICATION
                audioManager.isSpeakerphoneOn = speakerOn
            } else {
                audioManager.mode = AudioManager.MODE_NORMAL
                audioManager.isSpeakerphoneOn = false
            }
        } catch (e: Exception) {
            Log.e(TAG, "setAudioMode error", e)
        }
    }
}
