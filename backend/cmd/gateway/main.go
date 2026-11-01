package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"google.golang.org/grpc"

	"mim/pkg/discovery"
	"mim/pkg/jwtx"
	"mim/pkg/metrics"
	"mim/pkg/redisx"
	"mim/protocol"
	rpcfw "mim/rpc/client"
	rpcgw "mim/rpc/gateway"
	"mim/service/chat"
)

var (
	wsPort    = flag.Int("port", 8081, "WebSocket gateway listen port")
	grpcPort  = flag.Int("grpc-port", 9092, "Gateway internal gRPC push port")
	logicAddr = flag.String("logic-addr", "127.0.0.1:9091", "Logic service address (fallback if etcd empty)")
	redisAddr = flag.String("redis-addr", "127.0.0.1:6379", "redis host:port")
	redisPass = flag.String("redis-pass", "redis", "redis password")
	etcdAddr  = flag.String("etcd-addr", "127.0.0.1:2379", "etcd host:port (optional)")
	nodeID    = flag.String("node-id", "gateway-1", "gateway instance node id")
	jwtSecret = flag.String("jwt-secret", "mim-enterprise-jwt-secret-key-2026", "jwt secret for edge auth")
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024 * 64,
	WriteBufferSize: 1024 * 64,
}

type GatewayApp struct {
	hub       *chat.Hub
	rdb       *redisx.Client
	rpcMgr    *rpcfw.Manager
	logicDisc *discovery.Discovery
}

func main() {
	flag.Parse()

	if envSecret := os.Getenv("MIM_JWT_SECRET"); envSecret != "" {
		*jwtSecret = envSecret
	}

	log.Printf("[Gateway] Starting Mim Long-Connection Gateway [%s]...", *nodeID)

	// 1. Connect Redis (for multi-device routing & distributed kickout)
	rdb, err := redisx.NewClient(redisx.Config{
		Addr:     *redisAddr,
		Password: *redisPass,
	})
	if err != nil {
		log.Printf("[Gateway Warning] Redis not available (%s): %v. Local mode only.", *redisAddr, err)
		rdb = nil
	} else {
		log.Printf("[Gateway] Connected to Redis 7 at %s", *redisAddr)
	}

	hub := chat.NewHub(rdb)
	hub.SetGatewayAddr(fmt.Sprintf("127.0.0.1:%d", *grpcPort))
	rpcMgr := rpcfw.NewManager()

	// 2. Setup Etcd Discovery for Logic Services
	var etcdClient *clientv3.Client
	var logicDisc *discovery.Discovery
	if *etcdAddr != "" {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{*etcdAddr},
			DialTimeout: 2 * time.Second,
		})
		if err == nil {
			etcdClient = cli
			disc, err := discovery.NewDiscovery(etcdClient, "logic")
			if err == nil {
				logicDisc = disc
			}
		}
	}

	app := &GatewayApp{
		hub:       hub,
		rdb:       rdb,
		rpcMgr:    rpcMgr,
		logicDisc: logicDisc,
	}

	// 3. Start Gateway Internal gRPC Push Server
	gwLis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("[Gateway] Failed to listen gRPC port %d: %v", *grpcPort, err)
	}
	grpcServer := grpc.NewServer()
	gwServer := rpcgw.NewServer(hub)
	protocol.RegisterGatewayServiceServer(grpcServer, gwServer)

	go func() {
		log.Printf("[Gateway] Internal gRPC push server listening on :%d", *grpcPort)
		if err := grpcServer.Serve(gwLis); err != nil {
			log.Printf("[Gateway] gRPC server terminated: %v", err)
		}
	}()

	// 4. Register Gateway to Etcd
	var registrar *discovery.Registrar
	if etcdClient != nil {
		gwAddr := fmt.Sprintf("127.0.0.1:%d", *grpcPort)
		reg, err := discovery.NewRegistrar(etcdClient, "gateway", *nodeID, gwAddr, 5)
		if err == nil {
			registrar = reg
		}
	}

	// 5. Start WebSocket Gateway HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", app.handleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP","online":` + fmt.Sprintf("%d", hub.OnlineCount()) + `}`))
	})
	mux.Handle("/metrics", metrics.Handler())

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", *wsPort),
		Handler: mux,
	}

	go func() {
		log.Printf("==================================================")
		log.Printf("  MIM Gateway (Stateless Long-Connection Cluster) ")
		log.Printf("  WebSocket:       ws://127.0.0.1:%d/ws", *wsPort)
		log.Printf("  Internal gRPC:   127.0.0.1:%d", *grpcPort)
		log.Printf("  Logic Target:    %s", *logicAddr)
		log.Printf("==================================================")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[Gateway] Listen error: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[Gateway] Shutting down...")
	if registrar != nil {
		_ = registrar.Close()
	}
	grpcServer.GracefulStop()
	_ = server.Shutdown(context.Background())
	rpcMgr.Close()
	log.Printf("[Gateway] Exited cleanly.")
}

func (app *GatewayApp) getLogicClient() (protocol.LogicServiceClient, error) {
	target := *logicAddr
	if app.logicDisc != nil {
		if ep, err := app.logicDisc.GetEndpoint(); err == nil && ep != "" {
			target = ep
		}
	}
	return app.rpcMgr.GetLogicClient(target)
}

func (app *GatewayApp) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	// Fast Edge JWT Check
	claims, err := jwtx.ParseToken(*jwtSecret, token)
	if err != nil {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
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
		log.Printf("[Gateway] WS upgrade failed: %v", err)
		return
	}

	client := chat.NewClient(userID, platform, deviceID, conn, app.hub)
	app.hub.Register(client)
	metrics.OnlineConnections.WithLabelValues(platform).Inc()
	log.Printf("[Gateway] User %d connected on %s (Device: %s). Online: %d", userID, platform, deviceID, app.hub.OnlineCount())

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
			app.hub.Unregister(client)
			client.Close()
			log.Printf("[Gateway] User %d disconnected. Online: %d", userID, app.hub.OnlineCount())
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
				break
			}
			_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))

			if msgType == websocket.BinaryMessage {
				app.processBinaryFrame(client, rawData)
			}
		}
	}()

	wg.Wait()
}

func (app *GatewayApp) processBinaryFrame(client *chat.Client, data []byte) {
	frame, err := protocol.Decode(data)
	if err != nil {
		return
	}

	switch frame.CmdType {
	case protocol.CmdHeartbeatPing:
		pong := protocol.NewFrame(protocol.CmdHeartbeatPong, frame.SeqID, 0, nil)
		if bytes, err := pong.Encode(); err == nil {
			client.Send(bytes)
		}

	case protocol.CmdChatMessageSend:
		var chatMsg protocol.ChatMessage
		if err := proto.Unmarshal(frame.Payload, &chatMsg); err != nil {
			_ = json.Unmarshal(frame.Payload, &chatMsg)
		}
		chatMsg.SenderId = client.UserID
		chatMsg.SenderDeviceId = client.DeviceID
		metrics.MessagesReceived.WithLabelValues(fmt.Sprintf("%d", chatMsg.ConvType), fmt.Sprintf("%d", chatMsg.MsgType)).Inc()

		logicCli, err := app.getLogicClient()
		if err != nil {
			log.Printf("[Gateway] Failed to get Logic client: %v", err)
			return
		}

		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		ack, err := logicCli.SendMessage(ctx, &chatMsg)
		cancel()
		metrics.MessageLatency.WithLabelValues("gateway_send_msg").Observe(time.Since(start).Seconds())

		if err != nil {
			ack = &protocol.MessageAck{
				ClientMsgId: chatMsg.ClientMsgId,
				Code:        500,
				ErrorMsg:    err.Error(),
				ServerTime:  time.Now().UnixMilli(),
			}
		}
		ackPayload, _ := json.Marshal(ack)
		ackFrame := protocol.NewFrame(protocol.CmdMessageAck, ack.SeqId, protocol.FlagJSON, ackPayload)
		if ackBytes, err := ackFrame.Encode(); err == nil {
			client.Send(ackBytes)
		}

	case protocol.CmdSyncReq:
		var syncReq protocol.SyncRequest
		if err := proto.Unmarshal(frame.Payload, &syncReq); err != nil {
			_ = json.Unmarshal(frame.Payload, &syncReq)
		}
		logicCli, err := app.getLogicClient()
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			resp, err := logicCli.SyncMessages(ctx, &syncReq)
			cancel()
			if err == nil {
				respPayload, _ := json.Marshal(resp)
				respFrame := protocol.NewFrame(protocol.CmdSyncResp, resp.LatestSeqId, protocol.FlagJSON, respPayload)
				if respBytes, err := respFrame.Encode(); err == nil {
					client.Send(respBytes)
				}
			}
		}

	case protocol.CmdDeliveryAck:
		var delivAck protocol.DeliveryAck
		if err := proto.Unmarshal(frame.Payload, &delivAck); err != nil {
			_ = json.Unmarshal(frame.Payload, &delivAck)
		}
		metrics.DeliveryAckTotal.WithLabelValues("delivered").Inc()
		if logicCli, err := app.getLogicClient(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = logicCli.ProcessDeliveryAck(ctx, &delivAck)
			cancel()
		}

	case protocol.CmdReadReceipt:
		var receipt protocol.ReadReceipt
		if err := proto.Unmarshal(frame.Payload, &receipt); err != nil {
			_ = json.Unmarshal(frame.Payload, &receipt)
		}
		metrics.ReadReceiptTotal.WithLabelValues("read").Inc()
		receipt.UserId = client.UserID
		if logicCli, err := app.getLogicClient(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = logicCli.ProcessReadReceipt(ctx, &receipt)
			cancel()
		}

	case protocol.CmdWebRTCSignal:
		var signal protocol.WebRTCSignal
		if err := proto.Unmarshal(frame.Payload, &signal); err != nil {
			_ = json.Unmarshal(frame.Payload, &signal)
		}
		signal.FromUserId = client.UserID
		if logicCli, err := app.getLogicClient(); err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, _ = logicCli.RouteWebRTCSignal(ctx, &signal)
			cancel()
		}
	}
}
