package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"mim/pkg/discovery"
	"mim/pkg/kafkax"
	"mim/pkg/redisx"
	"mim/protocol"
	rpcfw "mim/rpc/client"
)

// ClusterPusher implements chat.MessagePusher in a distributed microservice cluster.
// For WebRTC signaling or when Kafka is not configured, it routes directly to gateways via gRPC.
// For persistent chat messages, group pushes, revokes, and ACKs, it publishes PushTasks to Kafka
// for TransferService consumption and async peak-clipping delivery (with direct gRPC fallback).
type ClusterPusher struct {
	producer    *kafkax.Producer
	rpcMgr      *rpcfw.Manager
	gatewayDisc *discovery.Discovery
	defaultGw   string
	rdb         *redisx.Client
}

func NewClusterPusher(
	producer *kafkax.Producer,
	rpcMgr *rpcfw.Manager,
	gwDisc *discovery.Discovery,
	defaultGw string,
	rdb *redisx.Client,
) *ClusterPusher {
	return &ClusterPusher{
		producer:    producer,
		rpcMgr:      rpcMgr,
		gatewayDisc: gwDisc,
		defaultGw:   defaultGw,
		rdb:         rdb,
	}
}

// SendToUser pushes data to all active devices of userID across all gateway nodes
func (p *ClusterPusher) SendToUser(userID uint64, data []byte) bool {
	return p.dispatch(userID, "", data)
}

// SendToUserExclude pushes data to all devices of userID except excludeDeviceID
func (p *ClusterPusher) SendToUserExclude(userID uint64, excludeDeviceID string, data []byte) {
	p.dispatch(userID, excludeDeviceID, data)
}

func (p *ClusterPusher) dispatch(userID uint64, excludeDeviceID string, data []byte) bool {
	if userID == 0 || len(data) == 0 {
		return false
	}

	// 1. Fast inspection of frame: WebRTC signaling must be ultra-low latency, bypass Kafka directly to Gateway
	isWebRTC := false
	if len(data) >= 6 {
		cmdType := uint16(data[4])<<8 | uint16(data[5])
		if cmdType == protocol.CmdWebRTCSignal {
			isWebRTC = true
		}
	}

	// 2. If Kafka producer is available and not WebRTC, publish to Kafka pipeline
	if p.producer != nil && !isWebRTC {
		task := PushTask{
			TargetUserID:    userID,
			FrameBytes:      data,
			ExcludeDeviceID: excludeDeviceID,
		}
		taskBytes, err := json.Marshal(task)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			err = p.producer.Publish(ctx, TopicPush, []byte(fmt.Sprintf("%d", userID)), taskBytes)
			cancel()
			if err == nil {
				return true
			}
		}
	}

	// 3. Direct gRPC push to gateway (fallback or for WebRTC or when Kafka is not configured)
	return p.directGRPCPush(userID, excludeDeviceID, data)
}

func (p *ClusterPusher) directGRPCPush(userID uint64, excludeDeviceID string, data []byte) bool {
	if p.rpcMgr == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 1. Intelligent Redis Routing: determine target gateway nodes hosting user's active devices
	var targetAddrs []string
	if p.rdb != nil {
		sessions, err := p.rdb.GetUserDeviceSessions(ctx, userID)
		if err == nil {
			if len(sessions) == 0 {
				// User is offline across the entire cluster, skip redundant gateway RPC broadcasts
				return false
			}
			seen := make(map[string]bool)
			for _, sess := range sessions {
				if sess.GatewayAddr != "" && !seen[sess.GatewayAddr] {
					seen[sess.GatewayAddr] = true
					targetAddrs = append(targetAddrs, sess.GatewayAddr)
				}
			}
		}
	}

	// 2. If no specific gateway address was found, fallback to all discovered active gateways
	if len(targetAddrs) == 0 {
		if p.gatewayDisc != nil {
			targetAddrs = p.gatewayDisc.GetAllEndpoints()
		}
		if len(targetAddrs) == 0 && p.defaultGw != "" {
			targetAddrs = []string{p.defaultGw}
		}
	}

	if len(targetAddrs) == 0 {
		return false
	}

	req := &protocol.GatewayPushReq{
		UserId:          userID,
		FrameBytes:      data,
		ExcludeDeviceId: excludeDeviceID,
	}

	deliveredAny := false
	for _, addr := range targetAddrs {
		gwCli, err := p.rpcMgr.GetGatewayClient(addr)
		if err != nil {
			log.Printf("[ClusterPusher Warning] Failed to get gateway client for %s: %v", addr, err)
			continue
		}
		resp, err := gwCli.PushMessage(ctx, req)
		if err != nil {
			log.Printf("[ClusterPusher Warning] PushMessage to gateway %s failed: %v", addr, err)
			continue
		}
		if resp != nil && resp.Delivered {
			deliveredAny = true
		}
	}

	return deliveredAny
}
