package com.mim.im.utils

import android.content.Context
import android.media.AudioAttributes
import android.media.MediaPlayer
import android.util.Log
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.io.File

class AudioPlayerHelper(private val context: Context) {
    companion object {
        private const val TAG = "AudioPlayerHelper"
    }

    private var mediaPlayer: MediaPlayer? = null
    private val scope = CoroutineScope(Dispatchers.Main + SupervisorJob())
    private var progressJob: Job? = null

    private val _currentPlayingMsgId = MutableStateFlow<String?>(null)
    val currentPlayingMsgId: StateFlow<String?> = _currentPlayingMsgId.asStateFlow()

    private val _isPlaying = MutableStateFlow(false)
    val isPlaying: StateFlow<Boolean> = _isPlaying.asStateFlow()

    private val _playbackProgress = MutableStateFlow(0f)
    val playbackProgress: StateFlow<Float> = _playbackProgress.asStateFlow()

    fun playOrToggle(msgId: String, dataSource: String) {
        if (_currentPlayingMsgId.value == msgId) {
            mediaPlayer?.let { player ->
                if (player.isPlaying) {
                    player.pause()
                    _isPlaying.value = false
                    stopProgressTracking()
                } else {
                    player.start()
                    _isPlaying.value = true
                    startProgressTracking()
                }
                return
            }
        }

        // Different track or first time
        stop()

        try {
            mediaPlayer = MediaPlayer().apply {
                setAudioAttributes(
                    AudioAttributes.Builder()
                        .setContentType(AudioAttributes.CONTENT_TYPE_SPEECH)
                        .setUsage(AudioAttributes.USAGE_MEDIA)
                        .build()
                )

                if (dataSource.startsWith("http://") || dataSource.startsWith("https://")) {
                    setDataSource(dataSource)
                } else {
                    val file = File(dataSource)
                    setDataSource(file.absolutePath)
                }

                setOnPreparedListener { mp ->
                    mp.start()
                    _currentPlayingMsgId.value = msgId
                    _isPlaying.value = true
                    startProgressTracking()
                }

                setOnCompletionListener {
                    stop()
                }

                setOnErrorListener { _, what, extra ->
                    Log.e(TAG, "MediaPlayer error what=$what, extra=$extra")
                    stop()
                    true
                }

                prepareAsync()
            }
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start media player", e)
            stop()
        }
    }

    private fun startProgressTracking() {
        stopProgressTracking()
        progressJob = scope.launch {
            while (isActive) {
                mediaPlayer?.let { player ->
                    try {
                        if (player.isPlaying && player.duration > 0) {
                            _playbackProgress.value = player.currentPosition.toFloat() / player.duration
                        }
                    } catch (_: Exception) {}
                }
                delay(100)
            }
        }
    }

    private fun stopProgressTracking() {
        progressJob?.cancel()
        progressJob = null
    }

    fun stop() {
        stopProgressTracking()
        try {
            mediaPlayer?.apply {
                if (isPlaying) {
                    stop()
                }
                reset()
                release()
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error stopping media player", e)
        } finally {
            mediaPlayer = null
            _currentPlayingMsgId.value = null
            _isPlaying.value = false
            _playbackProgress.value = 0f
        }
    }

    fun release() {
        stop()
        scope.cancel()
    }
}
