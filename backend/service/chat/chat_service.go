package chat

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"mim/models"
	"mim/protocol"

	"google.golang.org/protobuf/proto"
	"gorm.io/gorm"
)

var (
	ErrRevokeTimeExceeded = errors.New("cannot revoke message older than 2 minutes")
	ErrMessageNotFound    = errors.New("message not found")
	ErrRevokeForbidden    = errors.New("cannot revoke message from another user")
)

type Service struct {
	db     *gorm.DB
	hub    *Hub
	seqMap map[string]uint64
	seqMu  sync.Mutex
}

func NewService(db *gorm.DB, hub *Hub) *Service {
	return &Service{
		db:     db,
		hub:    hub,
		seqMap: make(map[string]uint64),
	}
}

// FormatSingleConvID generates consistent conversation id for 1v1 chat: min_max
func FormatSingleConvID(u1, u2 uint64) string {
	if u1 < u2 {
		return fmt.Sprintf("%d_%d", u1, u2)
	}
	return fmt.Sprintf("%d_%d", u2, u1)
}

// NextSeqID allocates a strictly monotonically increasing sequence number for a conversation
// Monotonically increasing, gap-tolerant: Redis INCR / DB transaction failure creates gaps but preserves monotonic order
func (s *Service) NextSeqID(convID string) uint64 {
	s.seqMu.Lock()
	defer s.seqMu.Unlock()

	curr, ok := s.seqMap[convID]
	if !ok {
		// Initialize from DB if first time
		var maxSeq uint64
		row := s.db.Model(&models.Message{}).Where("conversation_id = ?", convID).Select("COALESCE(MAX(seq_id), 0)").Row()
		_ = row.Scan(&maxSeq)
		curr = maxSeq
	}

	curr++
	s.seqMap[convID] = curr
	return curr
}

type SendMessageReq struct {
	ClientMsgID    string
	SenderID       uint64
	ReceiverID     uint64 // 0 for group
	GroupID        uint64
	ConvType       int8
	MsgType        int8
	Content        string
	QuotedMsgID    uint64
	QuotedContent  string
	SenderDeviceID string
}

// SendMessage processes, persists and pushes a chat message with full reliability guarantees
func (s *Service) SendMessage(req *SendMessageReq) (*protocol.MessageAck, error) {
	// 1. Idempotency Check (Check if client_msg_id already exists)
	var existing models.Message
	err := s.db.Where("client_msg_id = ?", req.ClientMsgID).First(&existing).Error
	if err == nil {
		// Already processed, return existing ACK directly
		return &protocol.MessageAck{
			ClientMsgId: existing.ClientMsgID,
			MsgId:       existing.ID,
			SeqId:       existing.SeqID,
			Code:        0,
			ServerTime:  existing.CreatedAt.UnixMilli(),
		}, nil
	}

	// 2. Determine conversation ID
	convID := ""
	if req.ConvType == models.ConvTypeSingle {
		convID = FormatSingleConvID(req.SenderID, req.ReceiverID)
	} else {
		convID = fmt.Sprintf("group_%d", req.GroupID)
	}

	// 3. Allocate Monotonic SeqId
	seqID := s.NextSeqID(convID)
	now := time.Now()

	// 4. Persistence to Database
	msg := models.Message{
		ClientMsgID:    req.ClientMsgID,
		ConversationID: convID,
		ConvType:       req.ConvType,
		SenderID:       req.SenderID,
		ReceiverID:     req.ReceiverID,
		SeqID:          seqID,
		MsgType:        req.MsgType,
		Content:        req.Content,
		QuotedMsgID:    req.QuotedMsgID,
		QuotedContent:  req.QuotedContent,
		CreatedAt:      now,
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&msg).Error; err != nil {
			return err
		}

		// Update or Create Sender Conversation with CAS / Seq Guard to prevent reordered overwrites
		var senderConv models.Conversation
		tx.Where("user_id = ? AND conversation_id = ?", req.SenderID, convID).FirstOrCreate(&senderConv, models.Conversation{
			ConversationID: convID,
			UserID:         req.SenderID,
			TargetID:       req.ReceiverID,
			ConvType:       req.ConvType,
		})
		tx.Model(&models.Conversation{}).
			Where("id = ? AND last_msg_seq < ?", senderConv.ID, seqID).
			Updates(map[string]interface{}{
				"last_msg_id":      msg.ID,
				"last_msg_seq":     seqID,
				"last_msg_preview": req.Content,
				"last_msg_time":    now,
				"last_read_seq":    seqID, // sender has read their own message
			})

		// Update or Create Receiver Conversation (if single chat) with CAS / Seq Guard
		if req.ConvType == models.ConvTypeSingle {
			var receiverConv models.Conversation
			tx.Where("user_id = ? AND conversation_id = ?", req.ReceiverID, convID).FirstOrCreate(&receiverConv, models.Conversation{
				ConversationID: convID,
				UserID:         req.ReceiverID,
				TargetID:       req.SenderID,
				ConvType:       req.ConvType,
			})
			tx.Model(&models.Conversation{}).
				Where("id = ? AND last_msg_seq < ?", receiverConv.ID, seqID).
				Updates(map[string]interface{}{
					"last_msg_id":      msg.ID,
					"last_msg_seq":     seqID,
					"last_msg_preview": req.Content,
					"last_msg_time":    now,
					"unread_count":     gorm.Expr("unread_count + 1"),
				})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. Construct Protobuf ChatMessage & Wrap in Aligned Binary Frame
	protoMsg := &protocol.ChatMessage{
		MsgId:          msg.ID,
		ClientMsgId:    msg.ClientMsgID,
		ConversationId: convID,
		ConvType:       protocol.ConversationType(req.ConvType),
		SenderId:       req.SenderID,
		ReceiverId:     req.ReceiverID,
		MsgType:        protocol.MessageType(req.MsgType),
		Content:        req.Content,
		SeqId:          seqID,
		SendTime:       now.UnixMilli(),
	}
	if req.QuotedMsgID > 0 {
		protoMsg.Quote = &protocol.QuoteInfo{
			QuotedMsgId:    req.QuotedMsgID,
			PreviewContent: req.QuotedContent,
		}
	}

	payloadBytes, _ := proto.Marshal(protoMsg)
	pushFrame := protocol.NewFrame(protocol.CmdChatMessagePush, seqID, 0, payloadBytes)
	encodedFrame, _ := pushFrame.Encode()

	// 6. Push to Receiver (if online)
	if req.ConvType == models.ConvTypeSingle {
		s.hub.SendToUser(req.ReceiverID, encodedFrame)
		// Multi-device sync for Sender (exclude current sending device)
		s.hub.SendToUserExclude(req.SenderID, req.SenderDeviceID, encodedFrame)
	}

	// 7. Return ACK
	return &protocol.MessageAck{
		ClientMsgId: req.ClientMsgID,
		MsgId:       msg.ID,
		SeqId:       seqID,
		Code:        0,
		ServerTime:  now.UnixMilli(),
	}, nil
}

// RouteWebRTCSignal routes ephemeral WebRTC signaling (Offer/Answer/Candidate) strictly in memory
// Ephemeral signaling is NEVER stored in mim_messages database table
func (s *Service) RouteWebRTCSignal(signal *protocol.WebRTCSignal) error {
	payload, err := proto.Marshal(signal)
	if err != nil {
		return err
	}

	frame := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, 0, payload)
	encoded, err := frame.Encode()
	if err != nil {
		return err
	}

	s.hub.SendToUser(signal.ToUserId, encoded)
	return nil
}

// SyncMessages incrementally pulls messages after lastSeqID (Monotonic gap-tolerant query)
func (s *Service) SyncMessages(convID string, lastSeqID uint64, limit int) (*protocol.SyncResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var msgs []models.Message
	err := s.db.Where("conversation_id = ? AND seq_id > ?", convID, lastSeqID).
		Order("seq_id ASC").
		Limit(limit).
		Find(&msgs).Error

	if err != nil {
		return nil, err
	}

	resp := &protocol.SyncResponse{
		ConversationId: convID,
		Messages:       make([]*protocol.ChatMessage, 0, len(msgs)),
		HasMore:        len(msgs) == limit,
	}

	for _, m := range msgs {
		resp.Messages = append(resp.Messages, &protocol.ChatMessage{
			MsgId:          m.ID,
			ClientMsgId:    m.ClientMsgID,
			ConversationId: m.ConversationID,
			ConvType:       protocol.ConversationType(m.ConvType),
			SenderId:       m.SenderID,
			ReceiverId:     m.ReceiverID,
			MsgType:        protocol.MessageType(m.MsgType),
			Content:        m.Content,
			SeqId:          m.SeqID,
			SendTime:       m.CreatedAt.UnixMilli(),
		})
		if m.SeqID > resp.LatestSeqId {
			resp.LatestSeqId = m.SeqID
		}
	}

	return resp, nil
}

type ConvSyncCursor struct {
	ConversationID string
	LastSeqID      uint64
}

// BatchSyncMessages fetches incremental messages across multiple conversations in a single roundtrip
func (s *Service) BatchSyncMessages(cursors []ConvSyncCursor, perConvLimit int) (map[string]*protocol.SyncResponse, error) {
	results := make(map[string]*protocol.SyncResponse)
	for _, c := range cursors {
		resp, err := s.SyncMessages(c.ConversationID, c.LastSeqID, perConvLimit)
		if err != nil {
			return nil, err
		}
		results[c.ConversationID] = resp
	}
	return results, nil
}

// RevokeMessage handles 2-minute time window message revocation
func (s *Service) RevokeMessage(userID, msgID uint64) error {
	var msg models.Message
	if err := s.db.First(&msg, msgID).Error; err != nil {
		return ErrMessageNotFound
	}

	if msg.SenderID != userID {
		return ErrRevokeForbidden
	}

	// Check 2 minute limit
	if time.Since(msg.CreatedAt) > 2*time.Minute {
		return ErrRevokeTimeExceeded
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		msg.IsRevoked = 1
		msg.Content = "[消息已撤回]"
		if err := tx.Save(&msg).Error; err != nil {
			return err
		}

		// Update conversation preview with Seq Guard
		tx.Model(&models.Conversation{}).
			Where("conversation_id = ? AND last_msg_seq <= ?", msg.ConversationID, msg.SeqID).
			Update("last_msg_preview", "你撤回了一条消息")

		// Push revoke notify via hub
		protoRevoke := &protocol.ChatMessage{
			MsgId:          msg.ID,
			ConversationId: msg.ConversationID,
			SenderId:       userID,
			MsgType:        protocol.MessageType_REVOKE,
			Content:        "消息已撤回",
			SeqId:          msg.SeqID,
			SendTime:       time.Now().UnixMilli(),
		}
		payload, _ := proto.Marshal(protoRevoke)
		frame := protocol.NewFrame(protocol.CmdRevokeNotify, msg.SeqID, 0, payload)
		encoded, _ := frame.Encode()

		if msg.ConvType == models.ConvTypeSingle {
			s.hub.SendToUser(msg.ReceiverID, encoded)
			s.hub.SendToUser(msg.SenderID, encoded)
		}
		return nil
	})
}

// MarkRead resets unread count and updates read sequence cursor
func (s *Service) MarkRead(userID uint64, convID string, lastReadSeq uint64) error {
	return s.db.Model(&models.Conversation{}).
		Where("user_id = ? AND conversation_id = ?", userID, convID).
		Updates(map[string]interface{}{
			"unread_count":  0,
			"last_read_seq": lastReadSeq,
		}).Error
}
