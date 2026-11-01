package rpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"google.golang.org/grpc"
	"gorm.io/gorm"

	"mim/models"
	"mim/pkg/discovery"
	"mim/pkg/jwtx"
	"mim/protocol"
	rpcfw "mim/rpc/client"
	rpcgw "mim/rpc/gateway"
	rpclogic "mim/rpc/logic"
	"mim/service/auth"
	"mim/service/chat"
	"mim/service/transfer"
)

func TestEndToEndGRPCCommunication(t *testing.T) {
	// 1. Setup in-memory SQLite and Hub
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	_ = models.AutoMigrate(db)

	jwtSecret := "test-grpc-secret-2026"
	hub := chat.NewHub(nil)
	chatSvc := chat.NewService(db, hub, nil)
	authSvc := auth.NewService(db, auth.Config{JwtSecret: jwtSecret, AccessExpire: time.Hour, RefreshExpire: 24 * time.Hour})

	// 2. Start Logic gRPC Server on random port
	logicLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen logic: %v", err)
	}
	logicGrpc := grpc.NewServer()
	logicServer := rpclogic.NewServer(chatSvc, authSvc, jwtSecret, nil)
	protocol.RegisterLogicServiceServer(logicGrpc, logicServer)
	go func() {
		_ = logicGrpc.Serve(logicLis)
	}()
	defer logicGrpc.Stop()

	// 3. Start Gateway gRPC Server on random port
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

	// 4. Dial clients using Manager
	mgr := rpcfw.NewManager()
	defer mgr.Close()

	logicClient, err := mgr.GetLogicClient(logicLis.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect logicClient: %v", err)
	}

	gwClient, err := mgr.GetGatewayClient(gwLis.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect gwClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 5. Test VerifyToken via gRPC
	pair, _ := jwtx.GenerateTokenPair(jwtSecret, time.Hour, 24*time.Hour, 888, "grpc_user", "web", "dev_1")
	authResp, err := logicClient.VerifyToken(ctx, &protocol.AuthHandshake{
		Token:    pair.AccessToken,
		Platform: "web",
		DeviceId: "dev_1",
	})
	if err != nil {
		t.Fatalf("VerifyToken gRPC error: %v", err)
	}
	if !authResp.Valid || authResp.UserId != 888 {
		t.Fatalf("unexpected authResp: %+v", authResp)
	}

	// 6. Test SendMessage via gRPC
	// Create mock recipient user in DB
	db.Create(&models.User{ID: 888, Nickname: "Sender"})
	db.Create(&models.User{ID: 999, Nickname: "Receiver"})

	ack, err := logicClient.SendMessage(ctx, &protocol.ChatMessage{
		ClientMsgId:    "grpc_msg_uuid_1",
		SenderId:       888,
		ReceiverId:     999,
		ConvType:       protocol.ConversationType_SINGLE,
		MsgType:        protocol.MessageType_TEXT,
		Content:        "Hello over gRPC microservice!",
		SendTime:       time.Now().UnixMilli(),
	})
	if err != nil {
		t.Fatalf("SendMessage gRPC error: %v", err)
	}
	if ack.Code != 0 || ack.SeqId == 0 {
		t.Fatalf("unexpected MessageAck: %+v", ack)
	}

	// 7. Test Gateway PushMessage via gRPC
	// Connect a mock client to hub
	mockClient := chat.NewClient(999, "web", "receiver_tab", nil, hub)
	hub.Register(mockClient)

	pushResp, err := gwClient.PushMessage(ctx, &protocol.GatewayPushReq{
		UserId:     999,
		FrameBytes: []byte{0x4D, 0x49, 0x01, 0x00},
	})
	if err != nil {
		t.Fatalf("PushMessage gRPC error: %v", err)
	}
	if !pushResp.Delivered {
		t.Fatalf("expected message to be delivered to online user 999")
	}

	// Drain mockClient SendChan
	select {
	case data := <-mockClient.SendChan:
		if len(data) != 4 || data[0] != 0x4D || data[1] != 0x49 {
			t.Fatalf("unexpected payload received by client: %v", data)
		}
	default:
		t.Fatalf("client did not receive pushed data")
	}

	// 8. Test Gateway KickDevice via gRPC
	kickResp, err := gwClient.KickDevice(ctx, &protocol.GatewayKickReq{
		UserId:      999,
		Platform:    "web",
		NewDeviceId: "new_receiver_tab",
		Reason:      "Kicked via gRPC",
	})
	if err != nil {
		t.Fatalf("KickDevice gRPC error: %v", err)
	}
	if !kickResp.Kicked {
		t.Fatalf("expected device to be kicked")
	}

	// 9. Test SyncMessages via gRPC
	syncResp, err := logicClient.SyncMessages(ctx, &protocol.SyncRequest{
		ConversationId: "888_999",
		LastSeqId:      0,
		Limit:          10,
	})
	if err != nil {
		t.Fatalf("SyncMessages gRPC error: %v", err)
	}
	if len(syncResp.Messages) == 0 {
		t.Fatalf("expected synced messages, got 0")
	}

	// 10. Test DeliveryAck via gRPC
	delivAck, err := logicClient.ProcessDeliveryAck(ctx, &protocol.DeliveryAck{
		ClientMsgId:    "grpc_msg_uuid_1",
		ConversationId: "888_999",
		SeqId:          ack.SeqId,
		DeliveryTime:   time.Now().UnixMilli(),
	})
	if err != nil || delivAck.Code != 0 {
		t.Fatalf("ProcessDeliveryAck gRPC failed: %v, %+v", err, delivAck)
	}
}

// TestDistributedClusterEndToEndPushing verifies the complete multi-gateway microservice flow:
// - Gateway 1 hosts User 101 (web, dev_A1)
// - Gateway 2 hosts User 202 (android, dev_B1) and User 101 (desktop, dev_A2)
// - Logic Service runs independently with ClusterPusher (NO shared Hub!)
// - Proves that SendMessage, Multi-Device Sync, WebRTC Signaling, DeliveryAck, ReadReceipt,
//   and Group Chat dispatch cross-node to the correct Gateway instances without loss.
func TestDistributedClusterEndToEndPushing(t *testing.T) {
	// 1. Setup in-memory SQLite DB
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	_ = models.AutoMigrate(db)

	db.Create(&models.User{ID: 101, Account: "user101", Nickname: "User 101"})
	db.Create(&models.User{ID: 202, Account: "user202", Nickname: "User 202"})
	db.Create(&models.Group{ID: 303, Title: "Test Dev Group", OwnerID: 101})
	db.Create(&models.GroupMember{GroupID: 303, UserID: 101, Role: 1})
	db.Create(&models.GroupMember{GroupID: 303, UserID: 202, Role: 3})

	mgr := rpcfw.NewManager()
	defer mgr.Close()

	// 2. Start Gateway 1
	gwLis1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen gw1: %v", err)
	}
	hub1 := chat.NewHub(nil)
	clientA1 := chat.NewClient(101, "web", "dev_A1", nil, hub1)
	hub1.Register(clientA1)

	gwGrpc1 := grpc.NewServer()
	protocol.RegisterGatewayServiceServer(gwGrpc1, rpcgw.NewServer(hub1))
	go func() { _ = gwGrpc1.Serve(gwLis1) }()
	defer gwGrpc1.Stop()

	reg1, _ := discovery.NewRegistrar(nil, "gateway", "gw-node-1", gwLis1.Addr().String(), 10)
	defer reg1.Close()

	// 3. Start Gateway 2
	gwLis2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen gw2: %v", err)
	}
	hub2 := chat.NewHub(nil)
	clientB1 := chat.NewClient(202, "android", "dev_B1", nil, hub2)
	hub2.Register(clientB1)
	// User 101 also connected on Desktop on Gateway 2!
	clientA2 := chat.NewClient(101, "desktop", "dev_A2", nil, hub2)
	hub2.Register(clientA2)

	gwGrpc2 := grpc.NewServer()
	protocol.RegisterGatewayServiceServer(gwGrpc2, rpcgw.NewServer(hub2))
	go func() { _ = gwGrpc2.Serve(gwLis2) }()
	defer gwGrpc2.Stop()

	reg2, _ := discovery.NewRegistrar(nil, "gateway", "gw-node-2", gwLis2.Addr().String(), 10)
	defer reg2.Close()

	// 4. Setup Discovery & ClusterPusher for Logic Service
	gwDisc, err := discovery.NewDiscovery(nil, "gateway")
	if err != nil {
		t.Fatalf("failed to create gateway discovery: %v", err)
	}

	clusterPusher := transfer.NewClusterPusher(nil, mgr, gwDisc, "", nil)
	chatSvc := chat.NewService(db, clusterPusher, nil)
	authSvc := auth.NewService(db, auth.Config{JwtSecret: "test-secret", AccessExpire: time.Hour, RefreshExpire: 24 * time.Hour})

	logicLis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen logic: %v", err)
	}
	logicGrpc := grpc.NewServer()
	protocol.RegisterLogicServiceServer(logicGrpc, rpclogic.NewServer(chatSvc, authSvc, "test-secret", nil))
	go func() { _ = logicGrpc.Serve(logicLis) }()
	defer logicGrpc.Stop()

	logicClient, err := mgr.GetLogicClient(logicLis.Addr().String())
	if err != nil {
		t.Fatalf("failed to connect logicClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 5. Test 1v1 Message Push & Multi-Device Sync
	ack, err := logicClient.SendMessage(ctx, &protocol.ChatMessage{
		ClientMsgId:    "cluster_test_msg_1",
		SenderId:       101,
		ReceiverId:     202,
		ConvType:       protocol.ConversationType_SINGLE,
		MsgType:        protocol.MessageType_TEXT,
		Content:        "Hello cross-gateway microservice!",
		SendTime:       time.Now().UnixMilli(),
		ConversationId: "101_202",
		SenderDeviceId: "dev_A1",
	})
	if err != nil || ack.Code != 0 {
		t.Fatalf("SendMessage failed: %v, %+v", err, ack)
	}

	// Verify Client B1 on Gateway 2 received the push
	select {
	case data := <-clientB1.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil || frame.CmdType != protocol.CmdChatMessagePush {
			t.Fatalf("expected CmdChatMessagePush on Client B1, got: %v, cmd: 0x%04X", err, frame.CmdType)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout: Client B1 on Gateway 2 did not receive pushed message")
	}

	// Verify Client A2 (User 101 on desktop, Gateway 2) received multi-device sync
	select {
	case data := <-clientA2.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil || frame.CmdType != protocol.CmdChatMessagePush {
			t.Fatalf("expected CmdChatMessagePush sync on Client A2, got: %v, cmd: 0x%04X", err, frame.CmdType)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout: Client A2 on Gateway 2 did not receive multi-device sync")
	}

	// Verify Client A1 (the sending device dev_A1) did NOT receive duplicate echo
	select {
	case data := <-clientA1.SendChan:
		t.Fatalf("Client A1 unexpectedly received echo data: %v", data)
	default:
		// Passed: properly excluded sending device
	}

	// 6. Test WebRTC Signaling cross-node routing
	signalResp, err := logicClient.RouteWebRTCSignal(ctx, &protocol.WebRTCSignal{
		CallId:     "call_xyz_1",
		FromUserId: 101,
		ToUserId:   202,
		SignalType: "OFFER",
		Data:       "v=0\r\no=alex...",
	})
	if err != nil || !signalResp.Success {
		t.Fatalf("RouteWebRTCSignal failed: %v, %+v", err, signalResp)
	}

	select {
	case data := <-clientB1.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil || frame.CmdType != protocol.CmdWebRTCSignal {
			t.Fatalf("expected CmdWebRTCSignal on Client B1, got: %v, cmd: 0x%04X", err, frame.CmdType)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout: Client B1 on Gateway 2 did not receive WebRTC signal")
	}

	// 7. Test DeliveryAck cross-node forward to sender on Gateway 1
	delivAckResp, err := logicClient.ProcessDeliveryAck(ctx, &protocol.DeliveryAck{
		ClientMsgId:    "cluster_test_msg_1",
		ConversationId: "101_202",
		SeqId:          ack.SeqId,
		DeliveryTime:   time.Now().UnixMilli(),
	})
	if err != nil || delivAckResp.Code != 0 {
		t.Fatalf("ProcessDeliveryAck failed: %v", err)
	}

	select {
	case data := <-clientA1.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil || frame.CmdType != protocol.CmdDeliveryAck {
			t.Fatalf("expected CmdDeliveryAck on Client A1, got: %v, cmd: 0x%04X", err, frame.CmdType)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout: Client A1 on Gateway 1 did not receive DeliveryAck")
	}

	// 8. Test ReadReceipt cross-node forward to sender on Gateway 1
	readReceiptResp, err := logicClient.ProcessReadReceipt(ctx, &protocol.ReadReceipt{
		ConversationId: "101_202",
		ReadSeqId:      ack.SeqId,
		UserId:         202,
		ReadTime:       time.Now().UnixMilli(),
	})
	if err != nil || readReceiptResp.Code != 0 {
		t.Fatalf("ProcessReadReceipt failed: %v", err)
	}

	select {
	case data := <-clientA1.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil || frame.CmdType != protocol.CmdReadReceipt {
			t.Fatalf("expected CmdReadReceipt on Client A1, got: %v, cmd: 0x%04X", err, frame.CmdType)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout: Client A1 on Gateway 1 did not receive ReadReceipt")
	}

	// 9. Test Group Chat Message Delivery to members across gateways
	grpAck, err := logicClient.SendMessage(ctx, &protocol.ChatMessage{
		ClientMsgId:    "cluster_group_msg_1",
		SenderId:       101,
		ReceiverId:     303, // Group ID
		ConvType:       protocol.ConversationType_GROUP,
		MsgType:        protocol.MessageType_TEXT,
		Content:        "Hello dev team group across clusters!",
		SendTime:       time.Now().UnixMilli(),
		ConversationId: "group_303",
		SenderDeviceId: "dev_A1",
	})
	if err != nil || grpAck.Code != 0 {
		t.Fatalf("Group SendMessage failed: %v, %+v", err, grpAck)
	}

	// Client B1 (member of group 303 on Gateway 2) must receive the group message
	select {
	case data := <-clientB1.SendChan:
		frame, err := protocol.Decode(data)
		if err != nil || frame.CmdType != protocol.CmdChatMessagePush {
			t.Fatalf("expected CmdChatMessagePush on group member B1, got: %v, cmd: 0x%04X", err, frame.CmdType)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout: Client B1 on Gateway 2 did not receive group message")
	}
}
