package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"mim/models"
	"mim/pkg/jwtx"
	"mim/pkg/metrics"
	"mim/protocol"
	"mim/service/chat"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin for local dev / desktop clients
	},
	ReadBufferSize:  1024 * 64,
	WriteBufferSize: 1024 * 64,
}

// WebSocket Gateway Connection Handler
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Parse and verify token from query parameter: /ws?token=<token>
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	claims, err := jwtx.ParseToken(s.cfg.JwtSecret, token)
	if err != nil {
		http.Error(w, "Unauthorized: invalid or expired token", http.StatusUnauthorized)
		return
	}
	userID := claims.UserID

	platform := r.URL.Query().Get("platform")
	if platform == "" {
		platform = "web"
	}
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		deviceID = fmt.Sprintf("%s_%d", platform, time.Now().UnixNano())
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := chat.NewClient(userID, platform, deviceID, conn, s.hub)
	s.hub.Register(client)
	metrics.OnlineConnections.WithLabelValues(platform).Inc()
	log.Printf("[WS] User %d connected on %s (Device: %s). Online users: %d", userID, platform, deviceID, s.hub.OnlineCount())

	// Start pump routines
	var wg sync.WaitGroup
	wg.Add(2)

	// Write pump
	go func() {
		defer wg.Done()
		for data := range client.SendChan {
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				break
			}
		}
	}()

	// Read pump
	go func() {
		defer func() {
			metrics.OnlineConnections.WithLabelValues(platform).Dec()
			s.hub.Unregister(client)
			client.Close()
			log.Printf("[WS] User %d disconnected. Online users: %d", userID, s.hub.OnlineCount())
			wg.Done()
		}()

		conn.SetReadLimit(10 * 1024 * 1024)
		_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
			return nil
		})

		for {
			msgType, rawData, err := conn.ReadMessage()
			if err != nil {
				log.Printf("[WS] User %d ReadMessage error: %v", userID, err)
				break
			}

			_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))

			// Process incoming frame
			if msgType == websocket.BinaryMessage {
				s.processBinaryFrame(client, rawData)
			} else if msgType == websocket.TextMessage {
				s.processTextFrame(client, rawData)
			}
		}
	}()

	wg.Wait()
}

// processBinaryFrame parses the 24-byte protocol frame and dispatches actions
func (s *Server) processBinaryFrame(client *chat.Client, data []byte) {
	frame, err := protocol.Decode(data)
	if err != nil {
		log.Printf("[WS] Invalid binary frame from user %d: %v", client.UserID, err)
		return
	}

	switch frame.CmdType {
	case protocol.CmdHeartbeatPing:
		pong := protocol.NewFrame(protocol.CmdHeartbeatPong, frame.SeqID, 0, nil)
		if bytes, err := pong.Encode(); err == nil {
			client.Send(bytes)
		}

	case protocol.CmdChatMessageSend:
		s.handleFrameChatMessage(client, frame)

	case protocol.CmdSyncReq:
		s.handleFrameSync(client, frame)

	case protocol.CmdDeliveryAck, protocol.CmdMessageAck:
		s.handleFrameDeliveryAck(client, frame)

	case protocol.CmdReadReceipt:
		s.handleFrameReadReceipt(client, frame)

	case protocol.CmdWebRTCSignal:
		s.handleFrameWebRTC(client, frame)

	default:
		log.Printf("[WS] Unhandled CmdType: 0x%04X", frame.CmdType)
	}
}

// handleFrameChatMessage processes a ChatMessageSend packet
func (s *Server) handleFrameChatMessage(client *chat.Client, frame *protocol.Frame) {
	var chatMsg protocol.ChatMessage
	err := proto.Unmarshal(frame.Payload, &chatMsg)
	if err != nil {
		// Fallback to JSON payload
		err = json.Unmarshal(frame.Payload, &chatMsg)
	}
	if err != nil {
		log.Printf("[WS] Failed to decode chat message payload: %v", err)
		return
	}

	// 1. DFA Sensitive Word Filter
	filteredContent := s.dfaFilter.Filter(chatMsg.Content, '*')
	chatMsg.Content = filteredContent

	// 2. Verified Sender ID
	chatMsg.SenderId = client.UserID

	groupID := uint64(0)
	receiverID := chatMsg.ReceiverId
	if chatMsg.ConvType == 1 { // Group
		groupID = chatMsg.ReceiverId
		receiverID = 0
	}

	sendReq := &chat.SendMessageReq{
		ClientMsgID:    chatMsg.ClientMsgId,
		SenderID:       client.UserID,
		ReceiverID:     receiverID,
		GroupID:        groupID,
		ConvType:       int8(chatMsg.ConvType),
		MsgType:        int8(chatMsg.MsgType),
		Content:        filteredContent,
		SenderDeviceID: client.DeviceID,
	}
	if chatMsg.Quote != nil {
		sendReq.QuotedMsgID = chatMsg.Quote.QuotedMsgId
		sendReq.QuotedContent = chatMsg.Quote.PreviewContent
	}

	metrics.MessagesReceived.WithLabelValues(fmt.Sprintf("%d", chatMsg.ConvType), fmt.Sprintf("%d", chatMsg.MsgType)).Inc()
	start := time.Now()
	ack, err := s.chatService.SendMessage(sendReq)
	metrics.MessageLatency.WithLabelValues("send_message").Observe(time.Since(start).Seconds())

	if err != nil {
		log.Printf("[WS] SendMessage error: %v", err)
		ack = &protocol.MessageAck{
			ClientMsgId: chatMsg.ClientMsgId,
			Code:        500,
			ErrorMsg:    err.Error(),
			ServerTime:  time.Now().UnixMilli(),
		}
	}

	// Reply ACK to sender
	ackPayload, _ := json.Marshal(ack)
	ackFrame := protocol.NewFrame(protocol.CmdMessageAck, ack.SeqId, protocol.FlagJSON, ackPayload)
	if ackBytes, err := ackFrame.Encode(); err == nil {
		client.Send(ackBytes)
	}
}

// handleFrameSync processes incremental sync request
func (s *Server) handleFrameSync(client *chat.Client, frame *protocol.Frame) {
	var syncReq protocol.SyncRequest
	if err := proto.Unmarshal(frame.Payload, &syncReq); err != nil {
		_ = json.Unmarshal(frame.Payload, &syncReq)
	}

	resp, err := s.chatService.SyncMessages(syncReq.ConversationId, syncReq.LastSeqId, int(syncReq.Limit))
	if err != nil {
		return
	}

	respPayload, _ := json.Marshal(resp)
	respFrame := protocol.NewFrame(protocol.CmdSyncResp, resp.LatestSeqId, protocol.FlagJSON, respPayload)
	if respBytes, err := respFrame.Encode(); err == nil {
		client.Send(respBytes)
	}
}

// handleFrameDeliveryAck processes receiver delivery ACK (SENT -> DELIVERED)
func (s *Server) handleFrameDeliveryAck(client *chat.Client, frame *protocol.Frame) {
	var ack protocol.DeliveryAck
	if err := proto.Unmarshal(frame.Payload, &ack); err != nil {
		_ = json.Unmarshal(frame.Payload, &ack)
	}
	metrics.DeliveryAckTotal.WithLabelValues("delivered").Inc()
	if ack.SeqId > 0 && ack.ConversationId != "" {
		_ = s.chatService.ProcessDeliveryAck(&ack, client.UserID)
	}
}

// handleFrameReadReceipt processes receiver read cursor receipt (DELIVERED -> READ)
func (s *Server) handleFrameReadReceipt(client *chat.Client, frame *protocol.Frame) {
	var receipt protocol.ReadReceipt
	if err := proto.Unmarshal(frame.Payload, &receipt); err != nil {
		_ = json.Unmarshal(frame.Payload, &receipt)
	}
	metrics.ReadReceiptTotal.WithLabelValues("read").Inc()
	receipt.UserId = client.UserID // enforce authenticated user ID
	if receipt.ConversationId != "" {
		_ = s.chatService.ProcessReadReceipt(&receipt)
	}
}

// handleFrameWebRTC processes ephemeral WebRTC signaling (strictly memory-only)
func (s *Server) handleFrameWebRTC(client *chat.Client, frame *protocol.Frame) {
	var signal protocol.WebRTCSignal
	if err := proto.Unmarshal(frame.Payload, &signal); err != nil {
		_ = json.Unmarshal(frame.Payload, &signal)
	}

	signal.FromUserId = client.UserID

	// For call initiation (OFFER), verify that caller and callee are friends
	if signal.SignalType == "OFFER" {
		var count int64
		s.db.Model(&models.Friend{}).Where("user_id = ? AND friend_id = ?", client.UserID, signal.ToUserId).Count(&count)
		if count == 0 {
			respSignal := protocol.WebRTCSignal{
				CallId:     signal.CallId,
				FromUserId: signal.ToUserId,
				ToUserId:   client.UserID,
				SignalType: "REJECT",
				Data:       `{"reason":"对方不是您的好友，无法发起通话"}`,
			}
			replyPayload, _ := json.Marshal(&respSignal)
			replyFrame := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, replyPayload)
			if encoded, err := replyFrame.Encode(); err == nil {
				client.Send(encoded)
			}
			return
		}
	}

	// Relay signal to target peer
	payload, _ := json.Marshal(&signal)
	f := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, payload)
	if encoded, err := f.Encode(); err == nil {
		s.hub.SendToUser(signal.ToUserId, encoded)
	}

	// If HANGUP or REJECT, record call history into database for caller & callee
	if signal.SignalType == "HANGUP" || signal.SignalType == "REJECT" {
		var fromUser, toUser models.User
		_ = s.db.First(&fromUser, signal.FromUserId)
		_ = s.db.First(&toUser, signal.ToUserId)

		fromName := fromUser.Nickname
		if fromName == "" {
			fromName = fmt.Sprintf("User %d", signal.FromUserId)
		}
		toName := toUser.Nickname
		if toName == "" {
			toName = fmt.Sprintf("User %d", signal.ToUserId)
		}

		callID := signal.CallId
		if callID == "" {
			callID = fmt.Sprintf("call_%d_%d", time.Now().Unix(), signal.FromUserId)
		}

		status := "Completed"
		if signal.SignalType == "REJECT" {
			status = "Rejected"
		}

		callType := "video"
		duration := 0
		if signal.Data != "" {
			var d struct {
				CallType string `json:"callType"`
				Duration int    `json:"duration"`
			}
			if err := json.Unmarshal([]byte(signal.Data), &d); err == nil {
				if d.CallType != "" {
					callType = d.CallType
				}
				if d.Duration > 0 {
					duration = d.Duration
				}
			}
		}

		rec := models.CallRecord{
			CallID:     callID,
			CallerID:   signal.FromUserId,
			CallerName: fromName,
			CalleeID:   signal.ToUserId,
			CalleeName: toName,
			CallType:   callType,
			Status:     status,
			Duration:   duration,
			CreatedAt:  time.Now(),
		}
		s.db.Create(&rec)
	}
}

// processTextFrame supports developer debugging
func (s *Server) processTextFrame(client *chat.Client, data []byte) {
	var jsonReq struct {
		Action  string          `json:"action"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &jsonReq); err != nil {
		return
	}

	if jsonReq.Action == "ping" {
		_ = client.Conn.WriteJSON(map[string]interface{}{
			"action": "pong",
			"time":   time.Now().UnixMilli(),
		})
	}
}
