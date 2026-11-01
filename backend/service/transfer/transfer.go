package transfer

import (
	"context"
	"encoding/json"
	"log"

	"mim/pkg/discovery"
	"mim/pkg/kafkax"
	"mim/pkg/redisx"
	"mim/protocol"
	rpcfw "mim/rpc/client"
)

const (
	TopicInbound = "mim_msg_inbound"
	TopicPush    = "mim_msg_push"
)

// PushTask represents an asynchronous message push job dispatched through Kafka
type PushTask struct {
	TargetUserID    uint64 `json:"target_user_id"`
	ConversationID  string `json:"conversation_id"`
	SeqID           uint64 `json:"seq_id"`
	FrameBytes      []byte `json:"frame_bytes"`
	ExcludeDeviceID string `json:"exclude_device_id,omitempty"`
}

// TransferService consumes Kafka push events and routes to target gateway instances
type TransferService struct {
	consumer    *kafkax.Consumer
	producer    *kafkax.Producer
	rdb         *redisx.Client
	rpcMgr      *rpcfw.Manager
	gatewayDisc *discovery.Discovery
}

func NewTransferService(
	producer *kafkax.Producer,
	consumer *kafkax.Consumer,
	rdb *redisx.Client,
	rpcMgr *rpcfw.Manager,
	gwDisc *discovery.Discovery,
) *TransferService {
	return &TransferService{
		producer:    producer,
		consumer:    consumer,
		rdb:         rdb,
		rpcMgr:      rpcMgr,
		gatewayDisc: gwDisc,
	}
}

// Start launches the Kafka push task consumer loop
func (s *TransferService) Start(ctx context.Context) {
	if s.consumer == nil {
		return
	}

	s.consumer.Start(ctx, func(msg kafkax.Message) error {
		var task PushTask
		if err := json.Unmarshal(msg.Value, &task); err != nil {
			log.Printf("[Transfer] Invalid push task payload: %v", err)
			return err
		}

		return s.DispatchPush(ctx, &task)
	})
}

// DispatchPush delivers the push task to gateway nodes hosting the target user
func (s *TransferService) DispatchPush(ctx context.Context, task *PushTask) error {
	if task == nil || task.TargetUserID == 0 {
		return nil
	}

	// 1. Intelligent Redis Routing: determine target gateway nodes hosting user's active devices
	var targetAddrs []string
	if s.rdb != nil {
		sessions, err := s.rdb.GetUserDeviceSessions(ctx, task.TargetUserID)
		if err == nil {
			if len(sessions) == 0 {
				// User is offline across the cluster, skip push (synced later via SyncMessages)
				return nil
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

	// 2. Fallback to all discovered active gateway nodes if specific addresses not recorded
	if len(targetAddrs) == 0 {
		if s.gatewayDisc != nil {
			targetAddrs = s.gatewayDisc.GetAllEndpoints()
		}
	}

	if len(targetAddrs) == 0 {
		log.Printf("[Transfer Warning] No active gateway found for user %d, skipping push", task.TargetUserID)
		return nil
	}

	req := &protocol.GatewayPushReq{
		UserId:          task.TargetUserID,
		FrameBytes:      task.FrameBytes,
		ExcludeDeviceId: task.ExcludeDeviceID,
	}

	// 3. Deliver to target gateway node(s) via gRPC
	for _, addr := range targetAddrs {
		gwCli, err := s.rpcMgr.GetGatewayClient(addr)
		if err != nil {
			log.Printf("[Transfer Warning] Failed to get gateway client for %s: %v", addr, err)
			continue
		}

		resp, err := gwCli.PushMessage(ctx, req)
		if err != nil {
			log.Printf("[Transfer Warning] PushMessage to gateway %s failed: %v", addr, err)
			continue
		}
		if resp != nil && resp.Delivered {
			// Successfully pushed to user on this gateway
			break
		}
	}

	return nil
}

// PublishPushTask sends a push task to the Kafka push topic
func (s *TransferService) PublishPushTask(ctx context.Context, task *PushTask) error {
	if s.producer == nil || task == nil {
		return nil
	}

	bytes, err := json.Marshal(task)
	if err != nil {
		return err
	}

	return s.producer.Publish(ctx, TopicPush, []byte(task.ConversationID), bytes)
}
