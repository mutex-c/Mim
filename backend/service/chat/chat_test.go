package chat_test

import (
	"testing"
	"time"

	"mim/models"
	"mim/protocol"
	"mim/service/chat"

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

	// 4. Test Incremental Sync (Simulating offline recovery)
	convID := chat.FormatSingleConvID(u1.ID, u2.ID)
	// Bob pulls all messages since seq 0
	syncResp, err := svc.SyncMessages(convID, 0, 50)
	if err != nil {
		t.Fatalf("sync messages failed: %v", err)
	}
	if len(syncResp.Messages) != 2 {
		t.Fatalf("expected 2 messages synced, got: %d", len(syncResp.Messages))
	}
	if syncResp.Messages[0].SeqId != 1 || syncResp.Messages[1].SeqId != 2 {
		t.Errorf("sync sequence order incorrect")
	}

	// Bob pulls only messages after seq 1
	syncRespDelta, err := svc.SyncMessages(convID, 1, 50)
	if err != nil {
		t.Fatalf("delta sync failed: %v", err)
	}
	if len(syncRespDelta.Messages) != 1 || syncRespDelta.Messages[0].SeqId != 2 {
		t.Errorf("expected 1 message with seq_id 2 in delta sync")
	}

	// 5. Test Revoke Message
	err = svc.RevokeMessage(u1.ID, ack2.MsgId)
	if err != nil {
		t.Fatalf("revoke message failed: %v", err)
	}

	var revoked models.Message
	db.First(&revoked, ack2.MsgId)
	if revoked.IsRevoked != 1 {
		t.Errorf("expected is_revoked=1")
	}

	// 6. Test Mark Read
	err = svc.MarkRead(u2.ID, convID, 2)
	if err != nil {
		t.Fatalf("mark read failed: %v", err)
	}

	var conv models.Conversation
	db.Where("user_id = ? AND conversation_id = ?", u2.ID, convID).First(&conv)
	if conv.UnreadCount != 0 || conv.LastReadSeq != 2 {
		t.Errorf("mark read failed: unread=%d, lastReadSeq=%d", conv.UnreadCount, conv.LastReadSeq)
	}
}
