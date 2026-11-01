package gateway

import (
	"context"

	"mim/pkg/metrics"
	"mim/protocol"
	"mim/service/chat"
)

type Server struct {
	protocol.UnimplementedGatewayServiceServer
	hub *chat.Hub
}

func NewServer(hub *chat.Hub) *Server {
	return &Server{
		hub: hub,
	}
}

// PushMessage delivers a binary protocol frame to a user's active connections on this gateway
func (s *Server) PushMessage(ctx context.Context, req *protocol.GatewayPushReq) (*protocol.GatewayPushResp, error) {
	if s.hub == nil || req == nil {
		return &protocol.GatewayPushResp{Delivered: false}, nil
	}

	metrics.MessagesPushed.WithLabelValues("push").Inc()
	var delivered bool
	if req.ExcludeDeviceId != "" {
		s.hub.SendToUserExclude(req.UserId, req.ExcludeDeviceId, req.FrameBytes)
		delivered = true
	} else {
		delivered = s.hub.SendToUser(req.UserId, req.FrameBytes)
	}
	platforms := s.hub.GetUserPlatforms(req.UserId)

	return &protocol.GatewayPushResp{
		Delivered:          delivered,
		TargetDeviceCount: int32(len(platforms)),
	}, nil
}

// KickDevice kicks any local connection matching (user_id, platform)
func (s *Server) KickDevice(ctx context.Context, req *protocol.GatewayKickReq) (*protocol.GatewayKickResp, error) {
	if s.hub == nil || req == nil {
		return &protocol.GatewayKickResp{Kicked: false}, nil
	}

	// Kick local device via hub
	platforms := s.hub.GetUserPlatforms(req.UserId)
	hadMatch := false
	for _, p := range platforms {
		if p == req.Platform {
			hadMatch = true
			break
		}
	}

	if hadMatch {
		metrics.KickedUsersTotal.WithLabelValues(req.Platform).Inc()
		// Use KickLocal method on Hub
		s.hub.KickLocalDevice(req.UserId, req.Platform, req.NewDeviceId, req.Reason)
	}

	return &protocol.GatewayKickResp{
		Kicked: hadMatch,
	}, nil
}

// Broadcast sends frame to multiple users connected to this gateway
func (s *Server) Broadcast(ctx context.Context, req *protocol.GatewayBroadcastReq) (*protocol.GatewayBroadcastResp, error) {
	if s.hub == nil || req == nil {
		return &protocol.GatewayBroadcastResp{SentCount: 0}, nil
	}

	sentCount := int32(0)
	for _, uid := range req.UserIds {
		if s.hub.SendToUser(uid, req.FrameBytes) {
			sentCount++
		}
	}

	return &protocol.GatewayBroadcastResp{
		SentCount: sentCount,
	}, nil
}
