package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"mim/pkg/jwtx"
	"mim/protocol"

	"github.com/gorilla/websocket"
)

type BenchConfig struct {
	TotalClients int
	TotalMsgs    int
	TargetURL    string
	Secret       string
	RampupBatch  int
}

type ClientNode struct {
	ID        uint64
	DeviceID  string
	Token     string
	Conn      *websocket.Conn
	writeMu   sync.Mutex
	CloseChan chan struct{}
	closeOnce sync.Once
}

func (c *ClientNode) Write(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.Conn == nil {
		return fmt.Errorf("connection closed")
	}
	_ = c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return c.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *ClientNode) Close() {
	c.closeOnce.Do(func() {
		close(c.CloseChan)
		c.writeMu.Lock()
		defer c.writeMu.Unlock()
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	})
}

func main() {
	totalClients := flag.Int("c", 1000, "Number of concurrent WebSocket connections")
	totalMsgs := flag.Int("m", 3000, "Number of chat messages to send")
	targetURL := flag.String("url", "ws://127.0.0.1:8080/ws", "WebSocket server gateway URL")
	flag.Parse()

	cfg := BenchConfig{
		TotalClients: *totalClients,
		TotalMsgs:    *totalMsgs,
		TargetURL:    *targetURL,
		Secret:       "mim-enterprise-jwt-secret-key-2026",
		RampupBatch:  50,
	}

	fmt.Println("================================================================================")
	fmt.Println("       MIM High-Concurrency Benchmark & Performance Stress Harness             ")
	fmt.Println("================================================================================")
	fmt.Printf(" Target Gateway:        %s\n", cfg.TargetURL)
	fmt.Printf(" Target Connections:    %d Concurrent Clients\n", cfg.TotalClients)
	fmt.Printf(" Target Message Burst:  %d End-to-End Encoded Messages\n", cfg.TotalMsgs)
	fmt.Println("================================================================================")

	// Step 1: Pre-generate JWT tokens for all clients
	log.Printf("[Phase 1/4] Pre-generating %d Client JWT Tokens...", cfg.TotalClients)
	clients := make([]*ClientNode, cfg.TotalClients)
	for i := 0; i < cfg.TotalClients; i++ {
		uid := uint64(i + 1)
		devID := fmt.Sprintf("bench_dev_%d", uid)
		pair, err := jwtx.GenerateTokenPair(cfg.Secret, 24*time.Hour, 24*time.Hour, uid, fmt.Sprintf("bench_u%d", uid), "desktop", devID)
		if err != nil {
			log.Fatalf("Token generation failed: %v", err)
		}
		clients[i] = &ClientNode{
			ID:        uid,
			DeviceID:  devID,
			Token:     pair.AccessToken,
			CloseChan: make(chan struct{}),
		}
	}
	log.Println("[Phase 1/4] Tokens generated successfully.")

	// Step 2: Concurrent Connection Ramp-Up
	log.Printf("[Phase 2/4] Ramp-up: Establishing %d concurrent WebSocket connections...", cfg.TotalClients)
	startRampup := time.Now()
	var connectedCount int64
	var failedCount int64

	// Concurrency-controlled dialer
	dialerSem := make(chan struct{}, cfg.RampupBatch)
	var dialWg sync.WaitGroup

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
	}

	for _, client := range clients {
		dialWg.Add(1)
		go func(c *ClientNode) {
			defer dialWg.Done()
			dialerSem <- struct{}{}
			defer func() { <-dialerSem }()

			wsURL := fmt.Sprintf("%s?user_id=%d&token=%s&platform=desktop&device_id=%s",
				cfg.TargetURL, c.ID, c.Token, c.DeviceID)

			conn, resp, err := dialer.Dial(wsURL, http.Header{})
			if err != nil {
				atomic.AddInt64(&failedCount, 1)
				return
			}
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}
			c.Conn = conn
			atomic.AddInt64(&connectedCount, 1)
		}(client)
	}

	dialWg.Wait()
	rampupDuration := time.Since(startRampup)
	connRate := float64(connectedCount) / rampupDuration.Seconds()

	fmt.Println("\n--------------------------------------------------------------------------------")
	fmt.Printf(" Connection Ramp-Up Results:\n")
	fmt.Printf("   Connected:           %d / %d (%.2f%%)\n", connectedCount, cfg.TotalClients, float64(connectedCount)/float64(cfg.TotalClients)*100)
	fmt.Printf("   Failed:              %d\n", failedCount)
	fmt.Printf("   Elapsed Time:        %.2f s\n", rampupDuration.Seconds())
	fmt.Printf("   Connect Rate:        %.1f conn/sec\n", connRate)
	fmt.Println("--------------------------------------------------------------------------------")

	if connectedCount == 0 {
		log.Fatalf("No connections succeeded. Is the MIM server running?")
	}

	// Step 3: Keepalive Ping-Pong RTT Benchmark
	log.Printf("[Phase 3/4] Measuring Ping-Pong Heartbeat RTT across %d connections...", connectedCount)

	activeClients := make([]*ClientNode, 0, connectedCount)
	for _, c := range clients {
		if c.Conn != nil {
			activeClients = append(activeClients, c)
		}
	}

	pingLatencies := make([]float64, 0, len(activeClients))
	var pingLatMu sync.Mutex
	var pingWg sync.WaitGroup

	hbPayload := []byte{}
	hbFrame := protocol.NewFrame(protocol.CmdHeartbeatPing, 0, 0, hbPayload)
	hbBytes, _ := hbFrame.Encode()

	for _, c := range activeClients {
		pingWg.Add(1)
		go func(node *ClientNode) {
			defer pingWg.Done()
			t0 := time.Now()

			if err := node.Write(hbBytes); err != nil {
				return
			}

			_ = node.Conn.SetReadDeadline(time.Now().Add(5 * time.Second))
			_, data, err := node.Conn.ReadMessage()
			if err != nil {
				return
			}
			f, err := protocol.Decode(data)
			if err != nil || f.CmdType != protocol.CmdHeartbeatPong {
				return
			}
			rtt := time.Since(t0).Seconds() * 1000.0 // ms
			pingLatMu.Lock()
			pingLatencies = append(pingLatencies, rtt)
			pingLatMu.Unlock()
		}(c)
	}

	pingWg.Wait()
	printLatencyStats("Heartbeat Ping-Pong RTT (Round Trip Time)", pingLatencies)

	// Step 4: High-Frequency Chat Message Throughput & Delivery Latency
	log.Printf("[Phase 4/4] End-to-End Chat Throughput & Latency Test (%d messages)...", cfg.TotalMsgs)

	// Pair up clients: client[i] -> client[(i+1) % N]
	numActive := len(activeClients)
	if numActive < 2 {
		log.Fatalf("Need at least 2 active clients for chat test")
	}

	// Track delivery latency using ClientMsgID map
	type PendingMsg struct {
		SendTime time.Time
	}
	pendingTracker := &sync.Map{}

	var msgsSent int64
	var acksReceived int64
	var pushesReceived int64
	var sendFailures int64

	ackLatencies := make([]float64, 0, cfg.TotalMsgs)
	pushLatencies := make([]float64, 0, cfg.TotalMsgs)
	var ackMu sync.Mutex
	var pushMu sync.Mutex

	stopChatReaders := make(chan struct{})

	// Start readers on all active clients to capture ACKs and PUSHes
	for _, c := range activeClients {
		go func(node *ClientNode) {
			for {
				select {
				case <-stopChatReaders:
					return
				default:
					_ = node.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
					_, data, err := node.Conn.ReadMessage()
					if err != nil {
						return
					}
					recvTime := time.Now()
					f, err := protocol.Decode(data)
					if err != nil {
						continue
					}

					switch f.CmdType {
					case protocol.CmdMessageAck:
						atomic.AddInt64(&acksReceived, 1)
						var ack protocol.MessageAck
						if err := json.Unmarshal(f.Payload, &ack); err == nil {
							if val, ok := pendingTracker.Load(ack.ClientMsgId); ok {
								p := val.(PendingMsg)
								rtt := recvTime.Sub(p.SendTime).Seconds() * 1000.0
								ackMu.Lock()
								ackLatencies = append(ackLatencies, rtt)
								ackMu.Unlock()
							}
						}

					case protocol.CmdChatMessagePush:
						atomic.AddInt64(&pushesReceived, 1)
						var push protocol.ChatMessage
						if err := json.Unmarshal(f.Payload, &push); err == nil {
							if val, ok := pendingTracker.Load(push.ClientMsgId); ok {
								p := val.(PendingMsg)
								rtt := recvTime.Sub(p.SendTime).Seconds() * 1000.0
								pushMu.Lock()
								pushLatencies = append(pushLatencies, rtt)
								pushMu.Unlock()
							}
						}
					}
				}
			}
		}(c)
	}

	time.Sleep(200 * time.Millisecond)

	// Dispatch burst messages across clients
	t0Chat := time.Now()
	var sendWg sync.WaitGroup
	sendThrottle := make(chan struct{}, 100) // 100 concurrent message dispatches

	for mIdx := 0; mIdx < cfg.TotalMsgs; mIdx++ {
		senderIdx := mIdx % numActive
		receiverIdx := (senderIdx + 1) % numActive

		sender := activeClients[senderIdx]
		receiver := activeClients[receiverIdx]

		clientMsgID := fmt.Sprintf("bench_msg_%d_%d", mIdx, time.Now().UnixNano())
		pendingTracker.Store(clientMsgID, PendingMsg{SendTime: time.Now()})

		sendWg.Add(1)
		go func(s *ClientNode, r *ClientNode, cMsgID string, seq uint64) {
			defer sendWg.Done()
			sendThrottle <- struct{}{}
			defer func() { <-sendThrottle }()

			chatReq := protocol.ChatMessage{
				ClientMsgId: cMsgID,
				ConvType:    protocol.ConversationType_SINGLE,
				ReceiverId:  r.ID,
				MsgType:     protocol.MessageType_TEXT,
				Content:     "Benchmarking high concurrency throughput with binary aligned frame!",
			}
			b, _ := json.Marshal(&chatReq)
			frame := protocol.NewFrame(protocol.CmdChatMessageSend, seq, protocol.FlagJSON, b)
			enc, err := frame.Encode()
			if err != nil {
				atomic.AddInt64(&sendFailures, 1)
				return
			}

			if err := s.Write(enc); err != nil {
				atomic.AddInt64(&sendFailures, 1)
			} else {
				atomic.AddInt64(&msgsSent, 1)
			}
		}(sender, receiver, clientMsgID, uint64(mIdx+1))
	}

	sendWg.Wait()
	log.Println("[INFO] All messages dispatched, collecting ACKs and PUSHes...")

	// Wait up to 5 seconds for in-flight ACKs/PUSHes to arrive
	timeout := time.After(5 * time.Second)
	for {
		if atomic.LoadInt64(&acksReceived) >= int64(cfg.TotalMsgs) &&
			atomic.LoadInt64(&pushesReceived) >= int64(cfg.TotalMsgs) {
			break
		}
		select {
		case <-timeout:
			goto BenchmarkFinished
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

BenchmarkFinished:
	chatDuration := time.Since(t0Chat)
	close(stopChatReaders)

	finalSent := atomic.LoadInt64(&msgsSent)
	finalAcks := atomic.LoadInt64(&acksReceived)
	finalPushes := atomic.LoadInt64(&pushesReceived)
	finalQPS := float64(finalAcks+finalPushes) / chatDuration.Seconds()

	fmt.Println("\n================================================================================")
	fmt.Printf(" High-Concurrency Chat Throughput Results:\n")
	fmt.Printf("   Total Dispatched:    %d messages\n", finalSent)
	fmt.Printf("   ACKs Confirmed:      %d / %d (%.2f%%)\n", finalAcks, cfg.TotalMsgs, float64(finalAcks)/float64(cfg.TotalMsgs)*100)
	fmt.Printf("   PUSHes Delivered:    %d / %d (%.2f%%)\n", finalPushes, cfg.TotalMsgs, float64(finalPushes)/float64(cfg.TotalMsgs)*100)
	fmt.Printf("   Total Time:          %.2f s\n", chatDuration.Seconds())
	fmt.Printf("   Composite QPS:       %.1f ops/sec (ACK + PUSH)\n", finalQPS)
	fmt.Printf("   Message TPS:         %.1f msgs/sec (End-to-End Completed)\n", float64(finalAcks)/chatDuration.Seconds())
	fmt.Println("================================================================================")

	printLatencyStats("Server ACK Latency (Send -> DB Persistence -> ACK)", ackLatencies)
	printLatencyStats("End-to-End Delivery Latency (Sender Dispatch -> Receiver PUSH)", pushLatencies)

	// Clean up connections
	log.Println("Tearing down benchmark connections...")
	for _, c := range activeClients {
		c.Close()
	}
	log.Println("Benchmark Completed Successfully.")
}

func printLatencyStats(title string, lats []float64) {
	fmt.Printf("\n--- %s ---\n", title)
	if len(lats) == 0 {
		fmt.Println("   No data points collected.")
		return
	}

	sort.Float64s(lats)
	n := len(lats)

	var sum float64
	min := lats[0]
	max := lats[n-1]
	for _, v := range lats {
		sum += v
	}
	avg := sum / float64(n)

	p50 := percentile(lats, 50)
	p90 := percentile(lats, 90)
	p95 := percentile(lats, 95)
	p99 := percentile(lats, 99)

	fmt.Printf("   Sample Count: %d\n", n)
	fmt.Printf("   Min Latency:  %.2f ms\n", min)
	fmt.Printf("   P50 (Median): %.2f ms\n", p50)
	fmt.Printf("   P90:          %.2f ms\n", p90)
	fmt.Printf("   P95:          %.2f ms\n", p95)
	fmt.Printf("   P99:          %.2f ms\n", p99)
	fmt.Printf("   Max Latency:  %.2f ms\n", max)
	fmt.Printf("   Mean Latency: %.2f ms\n", avg)
	fmt.Println("--------------------------------------------------------------------------------")
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Ceil(p/100.0*float64(len(sorted)))) - 1
	if idx < 0 {
		idx = 0
	}
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
