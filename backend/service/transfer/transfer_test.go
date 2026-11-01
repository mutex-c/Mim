package transfer_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"

	"mim/pkg/discovery"
	"mim/pkg/kafkax"
	"mim/protocol"
	rpcfw "mim/rpc/client"
	rpcgw "mim/rpc/gateway"
	"mim/service/chat"
	"mim/service/transfer"
)

func TestTransferServicePushPipeline(t *testing.T) {
	// 1. Setup mock gateway on random port
	hub := chat.NewHub(nil)
	gwLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen gw: %v", err)
	}
	gwGrpc := grpc.NewServer()
	gwServer := rpcgw.NewServer(hub)
	protocol.RegisterGatewayServiceServer(gwGrpc, gwServer)
	go func() {
		_ = gwGrpc.Serve(gwLis)
	}()
	defer gwGrpc.Stop()

	gwAddr := gwLis.Addr().String()

	// 2. Connect mock client to hub
	targetUID := uint64(5555)
	mockClient := chat.NewClient(targetUID, "web", "browser_dev", nil, hub)
	hub.Register(mockClient)

	// 3. Setup Discovery pointing to mock gateway
	reg, _ := discovery.NewRegistrar(nil, "gateway", "gw-node-1", gwAddr, 5)
	defer reg.Close()

	disc, err := discovery.NewDiscovery(nil, "gateway")
	if err != nil {
		t.Fatalf("NewDiscovery failed: %v", err)
	}
	defer disc.Close()

	// 4. Setup Kafka Producer & Consumer
	producer := kafkax.NewProducer(nil)
	defer producer.Close()

	consumer := kafkax.NewConsumer(nil, "transfer_group", transfer.TopicPush)
	defer consumer.Close()

	rpcMgr := rpcfw.NewManager()
	defer rpcMgr.Close()

	transferSvc := transfer.NewTransferService(producer, consumer, nil, rpcMgr, disc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// Listen for client received data
	go func() {
		select {
		case data := <-mockClient.SendChan:
			if len(data) >= 2 && data[0] == 0x4D && data[1] == 0x49 {
				wg.Done()
			}
		case <-time.After(3 * time.Second):
			t.Errorf("timeout waiting for client frame")
		}
	}()

	// Start transfer service consumer
	transferSvc.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	// 5. Publish push task to Kafka
	pushTask := &transfer.PushTask{
		TargetUserID:   targetUID,
		ConversationID: "111_5555",
		SeqID:          99,
		FrameBytes:     []byte{0x4D, 0x49, 0x01, 0x00, 0x01, 0x01},
	}

	err = transferSvc.PublishPushTask(ctx, pushTask)
	if err != nil {
		t.Fatalf("PublishPushTask failed: %v", err)
	}

	wg.Wait()
}

func TestClusterPusherDirectRouting(t *testing.T) {
	// 1. Setup mock gateway
	hub := chat.NewHub(nil)
	gwLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen gw: %v", err)
	}
	gwGrpc := grpc.NewServer()
	gwServer := rpcgw.NewServer(hub)
	protocol.RegisterGatewayServiceServer(gwGrpc, gwServer)
	go func() {
		_ = gwGrpc.Serve(gwLis)
	}()
	defer gwGrpc.Stop()

	gwAddr := gwLis.Addr().String()

	// 2. Register mock client
	targetUID := uint64(7777)
	mockClient := chat.NewClient(targetUID, "ios", "iphone_dev", nil, hub)
	hub.Register(mockClient)

	// 3. Setup Discovery
	reg, _ := discovery.NewRegistrar(nil, "gateway", "gw-node-direct", gwAddr, 5)
	defer reg.Close()

	disc, err := discovery.NewDiscovery(nil, "gateway")
	if err != nil {
		t.Fatalf("NewDiscovery failed: %v", err)
	}
	defer disc.Close()

	rpcMgr := rpcfw.NewManager()
	defer rpcMgr.Close()

	// ClusterPusher with no producer (forces direct gRPC push) and empty defaultGw
	pusher := transfer.NewClusterPusher(nil, rpcMgr, disc, "", nil)

	// WebRTC signal frame
	webrtcFrame := []byte{0x4D, 0x49, 0x01, 0x00, 0x05, 0x01, 0xAA, 0xBB}
	delivered := pusher.SendToUser(targetUID, webrtcFrame)
	if !delivered {
		t.Errorf("expected delivered to be true")
	}

	select {
	case data := <-mockClient.SendChan:
		if len(data) != len(webrtcFrame) {
			t.Errorf("received frame size mismatch: got %d, want %d", len(data), len(webrtcFrame))
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timeout waiting for direct gRPC pushed frame")
	}

	// Test unroutable/offline user
	offlineDelivered := pusher.SendToUser(9999999, webrtcFrame)
	if offlineDelivered {
		t.Errorf("expected delivered to be false for offline user")
	}
}

// TestClusterPusherMemoryProducerFallsBackToDirect guards a silent message-loss regression:
// when no Kafka cluster is configured, kafkax returns a process-local memory producer.
// Publishing chat frames to it from the logic service would succeed but never be consumed
// by mim-transfer (a different process), dropping every message without any error.
// The pusher must detect memory mode and take the direct gRPC path instead.
func TestClusterPusherMemoryProducerFallsBackToDirect(t *testing.T) {
	hub := chat.NewHub(nil)
	gwLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen gw: %v", err)
	}
	gwGrpc := grpc.NewServer()
	protocol.RegisterGatewayServiceServer(gwGrpc, rpcgw.NewServer(hub))
	go func() {
		_ = gwGrpc.Serve(gwLis)
	}()
	defer gwGrpc.Stop()

	gwAddr := gwLis.Addr().String()

	targetUID := uint64(8888)
	mockClient := chat.NewClient(targetUID, "web", "browser_dev", nil, hub)
	hub.Register(mockClient)

	reg, _ := discovery.NewRegistrar(nil, "gateway", "gw-node-mem", gwAddr, 5)
	defer reg.Close()

	disc, err := discovery.NewDiscovery(nil, "gateway")
	if err != nil {
		t.Fatalf("NewDiscovery failed: %v", err)
	}
	defer disc.Close()

	rpcMgr := rpcfw.NewManager()
	defer rpcMgr.Close()

	producer := kafkax.NewProducer(nil)
	defer producer.Close()
	if producer.IsClusterMode() {
		t.Fatalf("memory producer must not report cluster mode")
	}

	pusher := transfer.NewClusterPusher(producer, rpcMgr, disc, "", nil)

	// CmdChatMessagePush (0x0101) — the exact frame type that used to be swallowed.
	chatFrame := []byte{0x4D, 0x49, 0x01, 0x00, 0x01, 0x01, 0xAA, 0xBB}
	if delivered := pusher.SendToUser(targetUID, chatFrame); !delivered {
		t.Errorf("expected memory-mode producer to fall back to direct gRPC delivery")
	}

	select {
	case data := <-mockClient.SendChan:
		if len(data) != len(chatFrame) {
			t.Errorf("received frame size mismatch: got %d, want %d", len(data), len(chatFrame))
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timeout: chat frame was swallowed instead of being pushed directly")
	}
}
