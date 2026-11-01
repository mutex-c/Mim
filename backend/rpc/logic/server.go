package logic

import (
	"context"
	"errors"

	"mim/pkg/jwtx"
	"mim/pkg/kafkax"
	"mim/protocol"
	"mim/service/auth"
	"mim/service/chat"
)

type Server struct {
	protocol.UnimplementedLogicServiceServer
	chatService *chat.Service
	authService *auth.Service
	jwtSecret   string
	producer    *kafkax.Producer
}

func NewServer(chatSvc *chat.Service, authSvc *auth.Service, jwtSecret string, producer *kafkax.Producer) *Server {
	return &Server{
		chatService: chatSvc,
		authService: authSvc,
		jwtSecret:   jwtSecret,
		producer:    producer,
	}
}

// SendMessage receives chat message from gateway, persists, generates SeqID, and pushes
func (s *Server) SendMessage(ctx context.Context, req *protocol.ChatMessage) (*protocol.MessageAck, error) {
	if req == nil {
		return nil, errors.New("chat message is nil")
	}

	groupID := uint64(0)
	receiverID := req.ReceiverId
	if req.ConvType == protocol.ConversationType_GROUP {
		groupID = req.ReceiverId
		receiverID = 0
	}

	sendReq := &chat.SendMessageReq{
		ClientMsgID: req.ClientMsgId,
		SenderID:    req.SenderId,
		ReceiverID:  receiverID,
		GroupID:     groupID,
		ConvType:       int8(req.ConvType),
		MsgType:        int8(req.MsgType),
		Content:        req.Content,
		SenderDeviceID: req.SenderDeviceId,
	}

	if req.Quote != nil {
		sendReq.QuotedMsgID = req.Quote.QuotedMsgId
		sendReq.QuotedContent = req.Quote.PreviewContent
	}

	ack, err := s.chatService.SendMessage(sendReq)
	if err != nil {
		return ack, err
	}

	return ack, nil
}

// RouteWebRTCSignal routes ephemeral WebRTC signaling across cluster nodes
func (s *Server) RouteWebRTCSignal(ctx context.Context, req *protocol.WebRTCSignal) (*protocol.WebRTCSignalResp, error) {
	if req == nil {
		return &protocol.WebRTCSignalResp{Success: false}, errors.New("signal is nil")
	}
	err := s.chatService.RouteWebRTCSignal(req)
	return &protocol.WebRTCSignalResp{Success: err == nil}, err
}

// SyncMessages handles incremental message synchronization
func (s *Server) SyncMessages(ctx context.Context, req *protocol.SyncRequest) (*protocol.SyncResponse, error) {
	if req == nil {
		return nil, errors.New("sync request is nil")
	}
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}
	return s.chatService.SyncMessages(req.ConversationId, req.LastSeqId, limit)
}

// ProcessDeliveryAck consumes delivery acknowledgement
func (s *Server) ProcessDeliveryAck(ctx context.Context, req *protocol.DeliveryAck) (*protocol.MessageAck, error) {
	if req == nil {
		return nil, errors.New("delivery ack is nil")
	}
	err := s.chatService.ProcessDeliveryAck(req, 0)
	if err != nil {
		return &protocol.MessageAck{Code: 500, ErrorMsg: err.Error()}, nil
	}
	return &protocol.MessageAck{Code: 0, ClientMsgId: req.ClientMsgId, SeqId: req.SeqId}, nil
}

// ProcessReadReceipt consumes read cursor receipt
func (s *Server) ProcessReadReceipt(ctx context.Context, req *protocol.ReadReceipt) (*protocol.MessageAck, error) {
	if req == nil {
		return nil, errors.New("read receipt is nil")
	}
	err := s.chatService.ProcessReadReceipt(req)
	if err != nil {
		return &protocol.MessageAck{Code: 500, ErrorMsg: err.Error()}, nil
	}
	return &protocol.MessageAck{Code: 0, SeqId: req.ReadSeqId}, nil
}

// VerifyToken validates client authentication token
func (s *Server) VerifyToken(ctx context.Context, req *protocol.AuthHandshake) (*protocol.AuthHandshakeResp, error) {
	if req == nil || req.Token == "" {
		return &protocol.AuthHandshakeResp{Valid: false, Error: "missing token"}, nil
	}

	claims, err := jwtx.ParseToken(s.jwtSecret, req.Token)
	if err != nil {
		return &protocol.AuthHandshakeResp{Valid: false, Error: err.Error()}, nil
	}

	return &protocol.AuthHandshakeResp{
		Valid:  true,
		UserId: claims.UserID,
	}, nil
}
