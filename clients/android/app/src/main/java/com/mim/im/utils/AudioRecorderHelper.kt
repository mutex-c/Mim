package com.mim.im.utils

import android.content.Context
import android.media.MediaRecorder
import android.os.Build
import android.util.Log
import java.io.File

class AudioRecorderHelper(private val context: Context) {
    companion object {
        private const val TAG = "AudioRecorderHelper"
        const val MIN_RECORDING_DURATION_MS = 1000L
    }

    private var recorder: MediaRecorder? = null
    private var currentOutputFile: File? = null
    private var recordStartTime: Long = 0L

    val isRecording: Boolean
        get() = recorder != null

    fun startRecording(): Boolean {
        return try {
            val audioDir = File(context.cacheDir, "audio_records")
            if (!audioDir.exists()) {
                audioDir.mkdirs()
            }
            val outputFile = File(audioDir, "rec_${System.currentTimeMillis()}.m4a")
            currentOutputFile = outputFile

            recorder = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                MediaRecorder(context)
            } else {
                @Suppress("DEPRECATION")
                MediaRecorder()
            }.apply {
                setAudioSource(MediaRecorder.AudioSource.MIC)
                setOutputFormat(MediaRecorder.OutputFormat.MPEG_4)
                setAudioEncoder(MediaRecorder.AudioEncoder.AAC)
                setAudioSamplingRate(44100)
                setAudioEncodingBitRate(96000)
                setOutputFile(outputFile.absolutePath)
                prepare()
                start()
            }

            recordStartTime = System.currentTimeMillis()
            true
        } catch (e: Exception) {
            Log.e(TAG, "Failed to start audio recording", e)
            cleanup()
            false
        }
    }

    fun stopRecording(): Pair<File, Long>? {
        val duration = System.currentTimeMillis() - recordStartTime
        val file = currentOutputFile

        try {
            recorder?.apply {
                stop()
                reset()
                release()
            }
        } catch (e: Exception) {
            Log.e(TAG, "Error stopping recorder", e)
        } finally {
            recorder = null
        }

        if (duration < MIN_RECORDING_DURATION_MS || file == null || !file.exists() || file.length() == 0L) {
            file?.delete()
            currentOutputFile = null
            return null
        }

        currentOutputFile = null
        return Pair(file, duration)
    }

    fun cancelRecording() {
        try {
            recorder?.apply {
                stop()
                reset()
                release()
            }
        } catch (_: Exception) {} finally {
            recorder = null
            currentOutputFile?.delete()
            currentOutputFile = null
        }
    }

    fun getMaxAmplitude(): Int {
        return try {
            recorder?.maxAmplitude ?: 0
        } catch (_: Exception) {
            0
        }
    }

    private fun cleanup() {
        try {
            recorder?.reset()
            recorder?.release()
        } catch (_: Exception) {}
        recorder = null
        currentOutputFile?.delete()
        currentOutputFile = null
    }
}
