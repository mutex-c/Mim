export const MAGIC_NUMBER = 0x4D49;
export const CURRENT_VERSION = 0x01;
// 24 Bytes: 2B Magic + 1B Version + 1B Flags + 2B CmdType + 2B Reserved + 8B SeqID + 4B Length + 4B Checksum
export const HEADER_SIZE = 24;

export const FLAG_COMPRESSED = 1 << 0;
export const FLAG_ENCRYPTED = 1 << 1;
export const FLAG_JSON = 1 << 2;

export enum CmdType {
  HeartbeatPing = 0x0001,
  HeartbeatPong = 0x0002,
  AuthReq = 0x0010,
  AuthResp = 0x0011,

  ChatMessageSend = 0x0100,
  ChatMessagePush = 0x0101,
  MessageAck = 0x0102,

  GroupMessageSend = 0x0200,
  GroupMessagePush = 0x0201,

  SyncReq = 0x0300,
  SyncResp = 0x0301,
  BatchSyncReq = 0x0302,
  BatchSyncResp = 0x0303,

  RevokeNotify = 0x0400,
  ReadReceipt = 0x0401,

  WebRTCSignal = 0x0500,
}

// CRC32 IEEE 802.3 Table
const CRC32_TABLE = new Uint32Array(256);
for (let i = 0; i < 256; i++) {
  let c = i;
  for (let j = 0; j < 8; j++) {
    c = (c & 1) ? (0xEDB88320 ^ (c >>> 1)) : (c >>> 1);
  }
  CRC32_TABLE[i] = c >>> 0;
}

export function computeCRC32(bytes: Uint8Array): number {
  let crc = 0xFFFFFFFF;
  for (let i = 0; i < bytes.length; i++) {
    crc = (crc >>> 8) ^ CRC32_TABLE[(crc ^ bytes[i]) & 0xFF];
  }
  return (crc ^ 0xFFFFFFFF) >>> 0;
}

export interface Frame {
  magic: number;
  version: number;
  flags: number;
  cmdType: CmdType;
  reserved: number;
  seqId: bigint;
  length: number;
  checksum: number;
  payload: Uint8Array;
}

// Encode binary Frame into ArrayBuffer (Strict Network Byte Order: Big-Endian)
export function encodeFrame(cmdType: CmdType, seqId: bigint, flags: number, payload: Uint8Array): ArrayBuffer {
  const length = payload.byteLength;
  const checksum = computeCRC32(payload);
  const buffer = new ArrayBuffer(HEADER_SIZE + length);
  const view = new DataView(buffer);

  // 1. Magic (2B) - Big Endian
  view.setUint16(0, MAGIC_NUMBER, false);
  // 2. Version (1B)
  view.setUint8(2, CURRENT_VERSION);
  // 3. Flags (1B)
  view.setUint8(3, flags);
  // 4. CmdType (2B) - Big Endian
  view.setUint16(4, cmdType, false);
  // 5. Reserved (2B) - 8-byte alignment padding
  view.setUint16(6, 0, false);
  // 6. SeqID (8B) - starts at offset 8 (Memory Aligned) - Big Endian
  view.setBigUint64(8, seqId, false);
  // 7. Length (4B) - Big Endian
  view.setUint32(16, length, false);
  // 8. Checksum (4B) - Big Endian
  view.setUint32(20, checksum, false);

  // 9. Payload
  if (length > 0) {
    new Uint8Array(buffer, HEADER_SIZE).set(payload);
  }

  return buffer;
}

// Decode ArrayBuffer into binary Frame (Strict Network Byte Order: Big-Endian)
export function decodeFrame(buffer: ArrayBuffer): Frame {
  if (buffer.byteLength < HEADER_SIZE) {
    throw new Error('Packet shorter than 24-byte header size');
  }

  const view = new DataView(buffer);
  const magic = view.getUint16(0, false);
  if (magic !== MAGIC_NUMBER) {
    throw new Error(`Invalid protocol magic: 0x${magic.toString(16)}`);
  }

  const version = view.getUint8(2);
  const flags = view.getUint8(3);
  const cmdType = view.getUint16(4, false) as CmdType;
  const reserved = view.getUint16(6, false);
  const seqId = view.getBigUint64(8, false); // Offset 8
  const length = view.getUint32(16, false);
  const checksum = view.getUint32(20, false);

  const payload = new Uint8Array(buffer, HEADER_SIZE, length);
  if (computeCRC32(payload) !== checksum) {
    throw new Error('CRC32 checksum mismatch in binary frame');
  }

  return {
    magic,
    version,
    flags,
    cmdType,
    reserved,
    seqId,
    length,
    checksum,
    payload,
  };
}
