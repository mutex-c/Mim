package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
)

const (
	// MagicNumber is the 2-byte protocol identifier ('M', 'I')
	MagicNumber uint16 = 0x4D49
	// CurrentVersion represents the protocol version
	CurrentVersion uint8 = 0x01
	// HeaderSize is the fixed size of the binary packet header
	HeaderSize = 22

	// Command Types
	CmdHeartbeatPing uint16 = 0x0001
	CmdHeartbeatPong uint16 = 0x0002
	CmdAuthReq       uint16 = 0x0010
	CmdAuthResp      uint16 = 0x0011

	CmdChatMessageSend uint16 = 0x0100
	CmdChatMessagePush uint16 = 0x0101
	CmdMessageAck      uint16 = 0x0102

	CmdGroupMessageSend uint16 = 0x0200
	CmdGroupMessagePush uint16 = 0x0201

	CmdSyncReq  uint16 = 0x0300
	CmdSyncResp uint16 = 0x0301

	CmdRevokeNotify uint16 = 0x0400
	CmdReadReceipt  uint16 = 0x0401

	CmdWebRTCSignal uint16 = 0x0500

	// Flags
	FlagCompressed uint8 = 1 << 0
	FlagEncrypted  uint8 = 1 << 1
)

var (
	ErrInvalidMagic    = errors.New("invalid protocol magic number")
	ErrInvalidChecksum = errors.New("crc32 checksum verification failed")
	ErrPacketTooShort  = errors.New("packet is shorter than header length")
	ErrPayloadTooLarge = errors.New("payload length exceeds maximum allowed size")
)

const MaxPayloadSize = 10 * 1024 * 1024 // 10MB

// Frame represents a complete binary transport packet
type Frame struct {
	Magic    uint16
	Version  uint8
	Flags    uint8
	CmdType  uint16
	SeqID    uint64
	Length   uint32
	Checksum uint32
	Payload  []byte
}

// NewFrame constructs a new Frame and automatically computes Length & Checksum
func NewFrame(cmdType uint16, seqID uint64, flags uint8, payload []byte) *Frame {
	checksum := crc32.ChecksumIEEE(payload)
	return &Frame{
		Magic:    MagicNumber,
		Version:  CurrentVersion,
		Flags:    flags,
		CmdType:  cmdType,
		SeqID:    seqID,
		Length:   uint32(len(payload)),
		Checksum: checksum,
		Payload:  payload,
	}
}

// Encode serializes the Frame into wire format bytes
func (f *Frame) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)

	// 1. Magic (2B)
	if err := binary.Write(buf, binary.BigEndian, f.Magic); err != nil {
		return nil, err
	}
	// 2. Version (1B)
	if err := binary.Write(buf, binary.BigEndian, f.Version); err != nil {
		return nil, err
	}
	// 3. Flags (1B)
	if err := binary.Write(buf, binary.BigEndian, f.Flags); err != nil {
		return nil, err
	}
	// 4. CmdType (2B)
	if err := binary.Write(buf, binary.BigEndian, f.CmdType); err != nil {
		return nil, err
	}
	// 5. Sequence (8B)
	if err := binary.Write(buf, binary.BigEndian, f.SeqID); err != nil {
		return nil, err
	}
	// 6. Length (4B)
	if err := binary.Write(buf, binary.BigEndian, f.Length); err != nil {
		return nil, err
	}
	// 7. Checksum (4B)
	if err := binary.Write(buf, binary.BigEndian, f.Checksum); err != nil {
		return nil, err
	}
	// 8. Payload
	if len(f.Payload) > 0 {
		buf.Write(f.Payload)
	}

	return buf.Bytes(), nil
}

// Decode parses a binary packet from a byte slice
func Decode(data []byte) (*Frame, error) {
	if len(data) < HeaderSize {
		return nil, ErrPacketTooShort
	}

	reader := bytes.NewReader(data)
	f := &Frame{}

	if err := binary.Read(reader, binary.BigEndian, &f.Magic); err != nil {
		return nil, err
	}
	if f.Magic != MagicNumber {
		return nil, ErrInvalidMagic
	}

	if err := binary.Read(reader, binary.BigEndian, &f.Version); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &f.Flags); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &f.CmdType); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &f.SeqID); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.BigEndian, &f.Length); err != nil {
		return nil, err
	}
	if f.Length > MaxPayloadSize {
		return nil, ErrPayloadTooLarge
	}

	if err := binary.Read(reader, binary.BigEndian, &f.Checksum); err != nil {
		return nil, err
	}

	if f.Length > 0 {
		f.Payload = make([]byte, f.Length)
		if _, err := io.ReadFull(reader, f.Payload); err != nil {
			return nil, err
		}
		// Validate checksum
		if crc32.ChecksumIEEE(f.Payload) != f.Checksum {
			return nil, ErrInvalidChecksum
		}
	}

	return f, nil
}
