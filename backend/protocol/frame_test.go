package protocol_test

import (
	"bytes"
	"testing"

	"mim/protocol"

	"google.golang.org/protobuf/proto"
)

func TestFrameEncodeDecode(t *testing.T) {
	// 1. Create a ChatMessage protobuf message
	chatMsg := &protocol.ChatMessage{
		MsgId:          123456789,
		ClientMsgId:    "client-uuid-test-1",
		ConversationId: "1_2",
		ConvType:       protocol.ConversationType_SINGLE,
		SenderId:       1,
		ReceiverId:     2,
		MsgType:        protocol.MessageType_TEXT,
		Content:        "Hello Mim Binary Protocol!",
		SeqId:          1001,
		SendTime:       1771650000000,
	}

	payloadBytes, err := proto.Marshal(chatMsg)
	if err != nil {
		t.Fatalf("failed to marshal protobuf: %v", err)
	}

	// 2. Wrap into a Frame
	frame := protocol.NewFrame(protocol.CmdChatMessageSend, 1001, 0, payloadBytes)
	encoded, err := frame.Encode()
	if err != nil {
		t.Fatalf("failed to encode frame: %v", err)
	}

	if len(encoded) != protocol.HeaderSize+len(payloadBytes) {
		t.Fatalf("unexpected encoded length: got %d, expected %d", len(encoded), protocol.HeaderSize+len(payloadBytes))
	}

	// 3. Decode Frame
	decodedFrame, err := protocol.Decode(encoded)
	if err != nil {
		t.Fatalf("failed to decode frame: %v", err)
	}

	if decodedFrame.Magic != protocol.MagicNumber {
		t.Errorf("magic mismatch: got %x", decodedFrame.Magic)
	}
	if decodedFrame.CmdType != protocol.CmdChatMessageSend {
		t.Errorf("cmd mismatch: got %x", decodedFrame.CmdType)
	}
	if decodedFrame.SeqID != 1001 {
		t.Errorf("seq id mismatch: got %d", decodedFrame.SeqID)
	}

	// 4. Unmarshal payload back to ChatMessage
	recoveredMsg := &protocol.ChatMessage{}
	if err := proto.Unmarshal(decodedFrame.Payload, recoveredMsg); err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if recoveredMsg.Content != chatMsg.Content {
		t.Errorf("content mismatch: got %s, expected %s", recoveredMsg.Content, chatMsg.Content)
	}
	if recoveredMsg.MsgId != chatMsg.MsgId {
		t.Errorf("msg_id mismatch: got %d, expected %d", recoveredMsg.MsgId, chatMsg.MsgId)
	}
}

func TestFrameChecksumValidation(t *testing.T) {
	payload := []byte("Sensitive message payload")
	frame := protocol.NewFrame(protocol.CmdChatMessageSend, 1, 0, payload)
	encoded, err := frame.Encode()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Tamper payload byte
	encoded[len(encoded)-1] ^= 0xFF

	_, err = protocol.Decode(encoded)
	if err != protocol.ErrInvalidChecksum {
		t.Fatalf("expected ErrInvalidChecksum, got: %v", err)
	}
}

func TestFrameInvalidMagic(t *testing.T) {
	corrupted := bytes.Repeat([]byte{0x00}, 30)
	_, err := protocol.Decode(corrupted)
	if err != protocol.ErrInvalidMagic {
		t.Fatalf("expected ErrInvalidMagic, got: %v", err)
	}
}
