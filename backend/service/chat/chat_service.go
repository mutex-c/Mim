package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"mim/models"
	"mim/pkg/redisx"
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
	pusher MessagePusher
	rdb    *redisx.Client
	seqMap map[string]uint64
	seqMu  sync.Mutex
}

func NewService(db *gorm.DB, pusher MessagePusher, rdb *redisx.Client) *Service {
	return &Service{
		db:     db,
		pusher: pusher,
		rdb:    rdb,
		seqMap: make(map[string]uint64),
	}
}

func (s *Service) sendToUser(userID uint64, data []byte) bool {
	if s.pusher != nil {
		return s.pusher.SendToUser(userID, data)
	}
	return false
}

func (s *Service) sendToUserExclude(userID uint64, excludeDeviceID string, data []byte) {
	if s.pusher != nil {
		s.pusher.SendToUserExclude(userID, excludeDeviceID, data)
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
// Monotonically increasing, gap-tolerant: Redis INCR with DB fallback preserves monotonic order
func (s *Service) NextSeqID(convID string) uint64 {
	if s.rdb != nil {
		seq, err := s.rdb.NextSeqID(context.Background(), convID, func() uint64 {
			var maxSeq uint64
			row := s.db.Model(&models.Message{}).Where("conversation_id = ?", convID).Select("COALESCE(MAX(seq_id), 0)").Row()
			_ = row.Scan(&maxSeq)
			return maxSeq
		})
		if err == nil {
			return seq
		}
	}

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
			// Idempotency: handle concurrent insert race condition (TOCTOU) gracefully
			var dup models.Message
			if s.db.Where("client_msg_id = ?", req.ClientMsgID).First(&dup).Error == nil {
				msg = dup
				seqID = dup.SeqID
				return nil
			}
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
		now := time.Now()
		msgPreview := req.Content
		if req.MsgType == models.MsgTypeImage {
			msgPreview = "[图片]"
		} else if req.MsgType == models.MsgTypeVoice {
			msgPreview = "[语音]"
		} else if req.MsgType == models.MsgTypeFile || strings.HasPrefix(req.Content, "/api/file/download") {
			msgPreview = "[文件]"
		} else if req.MsgType == models.MsgTypeRevoke {
			msgPreview = "一条消息已被撤回"
		}

		tx.Model(&models.Conversation{}).
			Where("id = ? AND last_msg_seq < ?", senderConv.ID, seqID).
			Updates(map[string]interface{}{
				"last_msg_id":      msg.ID,
				"last_msg_seq":     seqID,
				"last_msg_preview": msgPreview,
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
					"last_msg_preview": msgPreview,
					"last_msg_time":    now,
					"unread_count":     gorm.Expr("unread_count + 1"),
				})
		} else {
			// Group Chat: Update or Create Conversation for ALL group members
			var groupMembers []models.GroupMember
			tx.Where("group_id = ?", req.GroupID).Find(&groupMembers)
			for _, m := range groupMembers {
				var memberConv models.Conversation
				tx.Where("user_id = ? AND conversation_id = ?", m.UserID, convID).FirstOrCreate(&memberConv, models.Conversation{
					ConversationID: convID,
					UserID:         m.UserID,
					TargetID:       req.GroupID,
					ConvType:       req.ConvType,
				})

				updates := map[string]interface{}{
					"last_msg_id":      msg.ID,
					"last_msg_seq":     seqID,
					"last_msg_preview": msgPreview,
					"last_msg_time":    now,
				}
				if m.UserID == req.SenderID {
					updates["last_read_seq"] = seqID
				} else {
					updates["unread_count"] = gorm.Expr("unread_count + 1")
				}

				tx.Model(&models.Conversation{}).
					Where("id = ? AND last_msg_seq < ?", memberConv.ID, seqID).
					Updates(updates)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 5. Look up sender info for push metadata
	var sender models.User
	s.db.Select("id, nickname, avatar").First(&sender, req.SenderID)

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

	type ChatPushPayload struct {
		*protocol.ChatMessage
		SenderName   string `json:"sender_name"`
		SenderAvatar string `json:"sender_avatar"`
	}

	pushObj := ChatPushPayload{
		ChatMessage:  protoMsg,
		SenderName:   sender.Nickname,
		SenderAvatar: sender.Avatar,
	}

	payloadBytes, _ := json.Marshal(pushObj)
	pushFrame := protocol.NewFrame(protocol.CmdChatMessagePush, seqID, protocol.FlagJSON, payloadBytes)
	encodedFrame, _ := pushFrame.Encode()

	// 6. Push to Receiver (if single) or All Group Members (if group)
	// Note on group fan-out: Currently fans out 1 push task per member with partition key = userID,
	// which guarantees per-user chronological ordering across Kafka partitions and Transfer consumers.
	// For super-groups (10,000+ members), this can be further optimized by gateway-level aggregation
	// (dispatching 1 frame with a target user ID list per active gateway).
	if req.ConvType == models.ConvTypeSingle {
		s.sendToUser(req.ReceiverID, encodedFrame)
		// Multi-device sync for Sender (exclude current sending device)
		s.sendToUserExclude(req.SenderID, req.SenderDeviceID, encodedFrame)
	} else {
		// Group push to all members
		var members []models.GroupMember
		s.db.Where("group_id = ?", req.GroupID).Find(&members)
		for _, m := range members {
			if m.UserID == req.SenderID {
				s.sendToUserExclude(req.SenderID, req.SenderDeviceID, encodedFrame)
			} else {
				s.sendToUser(m.UserID, encodedFrame)
			}
		}
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

	s.sendToUser(signal.ToUserId, encoded)
	return nil
}

// SyncMessages incrementally pulls messages after lastSeqID (Monotonic gap-tolerant query)
func (s *Service) SyncMessages(convID string, lastSeqID uint64, limit int) (*protocol.SyncResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var msgs []models.Message
	var err error
	if lastSeqID > 0 {
		err = s.db.Where("conversation_id = ? AND seq_id > ?", convID, lastSeqID).
			Order("seq_id ASC").
			Limit(limit).
			Find(&msgs).Error
	} else {
		// When lastSeqID is 0, fetch the latest messages in chronological order
		var revMsgs []models.Message
		err = s.db.Where("conversation_id = ?", convID).
			Order("seq_id DESC").
			Limit(limit).
			Find(&revMsgs).Error
		if err == nil {
			msgs = make([]models.Message, len(revMsgs))
			for i, m := range revMsgs {
				msgs[len(revMsgs)-1-i] = m
			}
		}
	}

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
func (s *Service) RevokeMessage(userID, msgID uint64, clientMsgID string) error {
	var msg models.Message
	var err error
	if msgID > 0 {
		err = s.db.First(&msg, msgID).Error
	} else if clientMsgID != "" {
		err = s.db.Where("client_msg_id = ? AND sender_id = ?", clientMsgID, userID).First(&msg).Error
	} else {
		return ErrMessageNotFound
	}

	if err != nil {
		return ErrMessageNotFound
	}

	if msg.SenderID != userID {
		return ErrRevokeForbidden
	}

	// Check recall time limit (strictly 2 minutes by default per spec, or custom user config)
	maxRecallWindow := 2 * time.Minute
	var userCfg models.UserConfig
	if err := s.db.Where("user_id = ?", userID).First(&userCfg).Error; err == nil && userCfg.RecallTimeLimit > 0 {
		configuredLimit := time.Duration(userCfg.RecallTimeLimit) * time.Second
		maxRecallWindow = configuredLimit
	}

	if time.Since(msg.CreatedAt) > maxRecallWindow {
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
			Update("last_msg_preview", "一条消息已被撤回")

		// Push revoke notify via hub
		protoRevoke := &protocol.ChatMessage{
			MsgId:          msg.ID,
			ClientMsgId:    msg.ClientMsgID,
			ConversationId: msg.ConversationID,
			SenderId:       userID,
			MsgType:        protocol.MessageType_REVOKE,
			Content:        "消息已撤回",
			SeqId:          msg.SeqID,
			SendTime:       time.Now().UnixMilli(),
		}
		payload, _ := json.Marshal(protoRevoke)
		frame := protocol.NewFrame(protocol.CmdRevokeNotify, msg.SeqID, protocol.FlagJSON, payload)
		encoded, _ := frame.Encode()

		if msg.ConvType == models.ConvTypeSingle {
			s.sendToUser(msg.ReceiverID, encoded)
			s.sendToUser(msg.SenderID, encoded)
		} else {
			// Group push to all members
			var groupID uint64
			if msg.ReceiverID > 0 {
				groupID = msg.ReceiverID
			} else {
				fmt.Sscanf(msg.ConversationID, "group_%d", &groupID)
			}
			var members []models.GroupMember
			s.db.Where("group_id = ?", groupID).Find(&members)
			for _, m := range members {
				s.sendToUser(m.UserID, encoded)
			}
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

// ProcessDeliveryAck consumes client delivery confirmation (CmdDeliveryAck / 0x0103)
// Transitions message status: SENT -> DELIVERED, and notifies sender
func (s *Service) ProcessDeliveryAck(ack *protocol.DeliveryAck, receiverID uint64) error {
	if ack == nil || ack.ConversationId == "" || ack.SeqId == 0 {
		return errors.New("invalid delivery ack")
	}

	var msg models.Message
	if err := s.db.Where("conversation_id = ? AND seq_id = ?", ack.ConversationId, ack.SeqId).First(&msg).Error; err != nil {
		return err
	}

	// Update message status to Delivered
	if msg.Status < models.MsgStatusDelivered {
		s.db.Model(&models.Message{}).Where("id = ?", msg.ID).Update("status", models.MsgStatusDelivered)
	}

	// Notify sender that message has arrived at receiver's device
	ackPayload, _ := json.Marshal(ack)
	ackFrame := protocol.NewFrame(protocol.CmdDeliveryAck, ack.SeqId, protocol.FlagJSON, ackPayload)
	if bytes, err := ackFrame.Encode(); err == nil {
		s.sendToUser(msg.SenderID, bytes)
	}
	return nil
}

// ProcessReadReceipt consumes client read receipt (CmdReadReceipt / 0x0401)
// Transitions message status: DELIVERED -> READ, updates unread counter, and notifies sender
func (s *Service) ProcessReadReceipt(receipt *protocol.ReadReceipt) error {
	if receipt == nil || receipt.ConversationId == "" {
		return errors.New("invalid read receipt")
	}

	// 1. Update reader's conversation record (reset unread count)
	if err := s.MarkRead(receipt.UserId, receipt.ConversationId, receipt.ReadSeqId); err != nil {
		return err
	}

	// 2. Mark messages up to read_seq_id as Read
	s.db.Model(&models.Message{}).
		Where("conversation_id = ? AND seq_id <= ? AND sender_id != ? AND status < ?",
			receipt.ConversationId, receipt.ReadSeqId, receipt.UserId, models.MsgStatusRead).
		Update("status", models.MsgStatusRead)

	// 3. Notify other conversation participant(s) with CmdReadReceipt frame
	receiptPayload, _ := json.Marshal(receipt)
	receiptFrame := protocol.NewFrame(protocol.CmdReadReceipt, receipt.ReadSeqId, protocol.FlagJSON, receiptPayload)
	if bytes, err := receiptFrame.Encode(); err == nil {
		if strings.HasPrefix(receipt.ConversationId, "group_") {
			var groupID uint64
			fmt.Sscanf(receipt.ConversationId, "group_%d", &groupID)
			var members []models.GroupMember
			s.db.Where("group_id = ?", groupID).Find(&members)
			for _, m := range members {
				if m.UserID != receipt.UserId {
					s.sendToUser(m.UserID, bytes)
				}
			}
		} else {
			// Single chat: parse min_max
			var u1, u2 uint64
			if _, err := fmt.Sscanf(receipt.ConversationId, "%d_%d", &u1, &u2); err == nil {
				targetID := u1
				if targetID == receipt.UserId {
					targetID = u2
				}
				s.sendToUser(targetID, bytes)
			}
		}
	}
	return nil
}

