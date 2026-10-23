package chat_test

import (
	"testing"
	"time"

	"mim/models"
	"mim/protocol"
	"mim/service/chat"

	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupChatTest(t *testing.T) (*chat.Service, *chat.Hub, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	if err := models.AutoMigrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	hub := chat.NewHub()
	svc := chat.NewService(db, hub)

	return svc, hub, db
}

func TestChatReliabilityAndFlow(t *testing.T) {
	svc, hub, db := setupChatTest(t)

	// Create test users
	u1 := models.User{Account: "alice", Nickname: "Alice"}
	u2 := models.User{Account: "bob", Nickname: "Bob"}
	db.Create(&u1)
	db.Create(&u2)

	// Register fake client for Bob
	bobClient := chat.NewClient(u2.ID, "windows", "bob-desktop-1", nil, hub)
	hub.Register(bobClient)

	if !hub.IsOnline(u2.ID) {
		t.Errorf("expected Bob to be online")
	}

	// 1. Send first message
	sendReq1 := &chat.SendMessageReq{
		ClientMsgID:    "msg-uuid-1",
		SenderID:       u1.ID,
		ReceiverID:     u2.ID,
		ConvType:       models.ConvTypeSingle,
		MsgType:        models.MsgTypeText,
		Content:        "Hello Bob",
		SenderDeviceID: "alice-web-1",
	}

	ack1, err := svc.SendMessage(sendReq1)
	if err != nil {
		t.Fatalf("send message 1 failed: %v", err)
	}
	if ack1.SeqId != 1 {
		t.Errorf("expected seq_id=1, got: %d", ack1.SeqId)
	}

	// Verify Bob received push in channel
	select {
	case data := <-bobClient.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil {
			t.Fatalf("failed to decode pushed frame: %v", err)
		}
		if frame.CmdType != protocol.CmdChatMessagePush {
			t.Errorf("expected CmdChatMessagePush, got: %x", frame.CmdType)
		}
		if protocol.HeaderSize != 24 {
			t.Errorf("expected 24 byte aligned frame header")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for pushed message to Bob")
	}

	// 2. Send duplicate message with same client_msg_id (Idempotency test)
	ackDup, err := svc.SendMessage(sendReq1)
	if err != nil {
		t.Fatalf("resend duplicate message failed: %v", err)
	}
	if ackDup.SeqId != ack1.SeqId || ackDup.MsgId != ack1.MsgId {
		t.Errorf("idempotency failed: expected seq %d, got %d", ack1.SeqId, ackDup.SeqId)
	}

	// 3. Send second message (SeqId should strictly increment)
	sendReq2 := &chat.SendMessageReq{
		ClientMsgID:    "msg-uuid-2",
		SenderID:       u1.ID,
		ReceiverID:     u2.ID,
		ConvType:       models.ConvTypeSingle,
		MsgType:        models.MsgTypeText,
		Content:        "Are you there?",
		SenderDeviceID: "alice-web-1",
	}
	ack2, err := svc.SendMessage(sendReq2)
	if err != nil {
		t.Fatalf("send message 2 failed: %v", err)
	}
	if ack2.SeqId != 2 {
		t.Errorf("expected seq_id=2, got: %d", ack2.SeqId)
	}

	// Drain message 2 from Bob's channel
	select {
	case <-bobClient.SendChan:
	case <-time.After(500 * time.Millisecond):
	}

	// Verify CAS Guard in conversation: last_msg_seq should be 2
	convID := chat.FormatSingleConvID(u1.ID, u2.ID)
	var conv1 models.Conversation
	db.Where("user_id = ? AND conversation_id = ?", u1.ID, convID).First(&conv1)
	if conv1.LastMsgSeq != 2 || conv1.LastMsgPreview != "Are you there?" {
		t.Errorf("conversation last_msg_seq mismatch: got %d, preview: %s", conv1.LastMsgSeq, conv1.LastMsgPreview)
	}

	// 4. Test Batch Sync across multiple conversations in single roundtrip
	cursors := []chat.ConvSyncCursor{
		{ConversationID: convID, LastSeqID: 0},
	}
	batchResp, err := svc.BatchSyncMessages(cursors, 50)
	if err != nil {
		t.Fatalf("batch sync failed: %v", err)
	}
	if len(batchResp[convID].Messages) != 2 {
		t.Fatalf("expected 2 messages in batch sync, got: %d", len(batchResp[convID].Messages))
	}

	// 5. Test WebRTC Ephemeral Signaling (In-Memory Routing Only, No DB Write)
	signal := &protocol.WebRTCSignal{
		CallId:     "call-1001",
		FromUserId: u1.ID,
		ToUserId:   u2.ID,
		SignalType: "OFFER",
		Data:       "v=0\r\no=- 123456...",
	}

	var countBefore int64
	db.Model(&models.Message{}).Count(&countBefore)

	err = svc.RouteWebRTCSignal(signal)
	if err != nil {
		t.Fatalf("route webrtc signal failed: %v", err)
	}

	// Check DB: Ephemeral signaling must NOT be written to mim_messages
	var countAfter int64
	db.Model(&models.Message{}).Count(&countAfter)
	if countAfter != countBefore {
		t.Errorf("ephemeral signaling should not be persisted into messages table")
	}

	// Bob should receive signal frame from channel
	select {
	case data := <-bobClient.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil {
			t.Fatalf("failed to decode signal frame: %v", err)
		}
		if frame.CmdType != protocol.CmdWebRTCSignal {
			t.Errorf("expected CmdWebRTCSignal, got %x", frame.CmdType)
		}
		recoveredSignal := &protocol.WebRTCSignal{}
		_ = proto.Unmarshal(frame.Payload, recoveredSignal)
		if recoveredSignal.CallId != "call-1001" {
			t.Errorf("signal callId mismatch")
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timeout waiting for signal to Bob")
	}

	// 6. Test Revoke Message
	err = svc.RevokeMessage(u1.ID, ack2.MsgId)
	if err != nil {
		t.Fatalf("revoke message failed: %v", err)
	}

	var revoked models.Message
	db.First(&revoked, ack2.MsgId)
	if revoked.IsRevoked != 1 {
		t.Errorf("expected is_revoked=1")
	}

	// 7. Test Mark Read
	err = svc.MarkRead(u2.ID, convID, 2)
	if err != nil {
		t.Fatalf("mark read failed: %v", err)
	}

	var conv2 models.Conversation
	db.Where("user_id = ? AND conversation_id = ?", u2.ID, convID).First(&conv2)
	if conv2.UnreadCount != 0 || conv2.LastReadSeq != 2 {
		t.Errorf("mark read failed: unread=%d, lastReadSeq=%d", conv2.UnreadCount, conv2.LastReadSeq)
	}
}
