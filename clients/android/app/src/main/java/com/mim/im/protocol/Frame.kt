package com.mim.im.protocol

import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.util.zip.CRC32

object CmdType {
    const val HeartbeatPing: Short = 0x0001
    const val HeartbeatPong: Short = 0x0002
    const val AuthReq: Short = 0x0010
    const val AuthResp: Short = 0x0011

    const val ChatMessageSend: Short = 0x0100
    const val ChatMessageAck: Short = 0x0101
    const val ChatMessagePush: Short = 0x0102

    const val SyncReq: Short = 0x0200
    const val SyncResp: Short = 0x0201

    const val RevokeReq: Short = 0x0300
    const val RevokePush: Short = 0x0302

    const val WebRTCSignal: Short = 0x0400
}

object FrameFlag {
    const val JSON: Byte = 0x01
    const val Encrypted: Byte = 0x02
}

/**
 * 24-byte custom binary protocol frame implementation for Mim
 */
data class Frame(
    val cmdType: Short,
    val seqId: Long,
    val flag: Byte = FrameFlag.JSON,
    val payload: ByteArray = ByteArray(0)
) {
    companion object {
        const val MAGIC: Short = 0x4D49 // "MI"
        const val VERSION: Byte = 0x01
        const val HEADER_SIZE = 24

        fun decode(bytes: ByteArray): Frame? {
            if (bytes.size < HEADER_SIZE) return null
            val buffer = ByteBuffer.wrap(bytes).order(ByteOrder.BIG_ENDIAN)

            val magic = buffer.short
            if (magic != MAGIC) return null

            val version = buffer.get()
            val flag = buffer.get()
            val cmdType = buffer.short
            val seqId = buffer.long
            val reserved = buffer.short
            val payloadLen = buffer.int

            if (payloadLen < 0 || bytes.size < HEADER_SIZE + payloadLen) return null

            val checksum = buffer.int
            val payload = ByteArray(payloadLen)
            if (payloadLen > 0) {
                buffer.get(payload)
            }

            // CRC32 check
            val crc = CRC32()
            crc.update(payload)
            if (checksum != 0 && crc.value.toInt() != checksum) {
                // Checksum mismatch, ignore corrupted packet
                return null
            }

            return Frame(
                cmdType = cmdType,
                seqId = seqId,
                flag = flag,
                payload = payload
            )
        }
    }

    fun encode(): ByteArray {
        val payloadLen = payload.size
        val crc = CRC32()
        crc.update(payload)
        val checksum = crc.value.toInt()

        val totalSize = HEADER_SIZE + payloadLen
        val buffer = ByteBuffer.allocate(totalSize).order(ByteOrder.BIG_ENDIAN)

        buffer.putShort(MAGIC)
        buffer.put(VERSION)
        buffer.put(flag)
        buffer.putShort(cmdType)
        buffer.putLong(seqId)
        buffer.putShort(0.toShort()) // Reserved
        buffer.putInt(payloadLen)
        buffer.putInt(checksum)

        if (payloadLen > 0) {
            buffer.put(payload)
        }

        return buffer.array()
    }

    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as Frame

        if (cmdType != other.cmdType) return false
        if (seqId != other.seqId) return false
        if (flag != other.flag) return false
        if (!payload.contentEquals(other.payload)) return false

        return true
    }

    override fun hashCode(): Int {
        var result = cmdType.toInt()
        result = 31 * result + seqId.hashCode()
        result = 31 * result + flag.toInt()
        result = 31 * result + payload.contentHashCode()
        return result
    }
}
