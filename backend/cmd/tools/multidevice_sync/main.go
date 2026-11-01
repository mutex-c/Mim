package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"mim/pkg/jwtx"
	"mim/protocol"

	"github.com/gorilla/websocket"
)

const (
	serverWsURL   = "ws://127.0.0.1:8080/ws"
	serverHttpURL = "http://127.0.0.1:8080"
	jwtSecret     = "mim-enterprise-jwt-secret-key-2026"
)

func main() {
	log.Println("================================================================")
	log.Println("   Multi-Device Online & Offline Incremental Sync Test Suite    ")
	log.Println("================================================================")

	// Step 1: Generate Tokens
	// User 1 on Desktop
	p1Desktop, err := jwtx.GenerateTokenPair(jwtSecret, 24*time.Hour, 24*time.Hour, 1, "alex", "desktop", "dev_u1_desktop")
	if err != nil {
		log.Fatalf("Token gen error: %v", err)
	}

	// User 1 on Android
	p1Android, err := jwtx.GenerateTokenPair(jwtSecret, 24*time.Hour, 24*time.Hour, 1, "alex", "android", "dev_u1_android")
	if err != nil {
		log.Fatalf("Token gen error: %v", err)
	}

	// User 2 on Android
	p2Phone, err := jwtx.GenerateTokenPair(jwtSecret, 24*time.Hour, 24*time.Hour, 2, "sarah", "android", "dev_u2_phone")
	if err != nil {
		log.Fatalf("Token gen error: %v", err)
	}

	// Connect User 1 Desktop
	u1DesktopURL := fmt.Sprintf("%s?user_id=1&token=%s&platform=desktop&device_id=dev_u1_desktop", serverWsURL, p1Desktop.AccessToken)
	ws1Desktop, _, err := websocket.DefaultDialer.Dial(u1DesktopURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect ws1Desktop: %v", err)
	}
	defer ws1Desktop.Close()
	log.Println("[PASS] User 1 (Desktop) connected")

	// Connect User 1 Android
	u1AndroidURL := fmt.Sprintf("%s?user_id=1&token=%s&platform=android&device_id=dev_u1_android", serverWsURL, p1Android.AccessToken)
	ws1Android, _, err := websocket.DefaultDialer.Dial(u1AndroidURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect ws1Android: %v", err)
	}
	defer ws1Android.Close()
	log.Println("[PASS] User 1 (Android) connected")

	// Connect User 2 Phone
	u2PhoneURL := fmt.Sprintf("%s?user_id=2&token=%s&platform=android&device_id=dev_u2_phone", serverWsURL, p2Phone.AccessToken)
	ws2Phone, _, err := websocket.DefaultDialer.Dial(u2PhoneURL, nil)
	if err != nil {
		log.Fatalf("Failed to connect ws2Phone: %v", err)
	}
	defer ws2Phone.Close()
	log.Println("[PASS] User 2 (Phone) connected")

	time.Sleep(200 * time.Millisecond)

	// =========================================================================
	// TEST CASE 1: Multi-Device Real-Time Push & Echo Suppression
	// =========================================================================
	log.Println("\n--- Test Case 1: Real-time Multi-Device Sync (Desktop -> Android & Receiver) ---")

	u1DesktopAcks := make(chan *protocol.MessageAck, 10)
	u1DesktopPushes := make(chan *protocol.ChatMessage, 10)
	u1AndroidPushes := make(chan *protocol.ChatMessage, 10)
	u2PhonePushes := make(chan *protocol.ChatMessage, 10)

	var wsMutex sync.Mutex
	stopReaders := make(chan struct{})

	// Reader for ws1Desktop
	go func() {
		for {
			select {
			case <-stopReaders:
				return
			default:
				_, data, err := ws1Desktop.ReadMessage()
				if err != nil {
					return
				}
				frame, err := protocol.Decode(data)
				if err != nil {
					continue
				}
				if frame.CmdType == protocol.CmdMessageAck {
					var ack protocol.MessageAck
					_ = json.Unmarshal(frame.Payload, &ack)
					u1DesktopAcks <- &ack
				} else if frame.CmdType == protocol.CmdChatMessagePush {
					var chatMsg protocol.ChatMessage
					_ = json.Unmarshal(frame.Payload, &chatMsg)
					u1DesktopPushes <- &chatMsg
				}
			}
		}
	}()

	// Reader for ws1Android
	go func() {
		for {
			select {
			case <-stopReaders:
				return
			default:
				_, data, err := ws1Android.ReadMessage()
				if err != nil {
					return
				}
				frame, err := protocol.Decode(data)
				if err != nil {
					continue
				}
				if frame.CmdType == protocol.CmdChatMessagePush {
					var chatMsg protocol.ChatMessage
					_ = json.Unmarshal(frame.Payload, &chatMsg)
					u1AndroidPushes <- &chatMsg
				}
			}
		}
	}()

	// Reader for ws2Phone
	go func() {
		for {
			select {
			case <-stopReaders:
				return
			default:
				_, data, err := ws2Phone.ReadMessage()
				if err != nil {
					return
				}
				frame, err := protocol.Decode(data)
				if err != nil {
					continue
				}
				if frame.CmdType == protocol.CmdChatMessagePush {
					var chatMsg protocol.ChatMessage
					_ = json.Unmarshal(frame.Payload, &chatMsg)
					u2PhonePushes <- &chatMsg
				}
			}
		}
	}()

	// User 1 sends message on Desktop to User 2
	clientMsgID1 := fmt.Sprintf("u1_d_msg_%d", time.Now().UnixNano())
	sendMsg1 := protocol.ChatMessage{
		ClientMsgId: clientMsgID1,
		ConvType:    protocol.ConversationType_SINGLE,
		ReceiverId:  2,
		MsgType:     protocol.MessageType_TEXT,
		Content:     "Hello Sarah from my Desktop!",
	}
	sendBytes1, _ := json.Marshal(&sendMsg1)
	sendFrame1 := protocol.NewFrame(protocol.CmdChatMessageSend, 1, protocol.FlagJSON, sendBytes1)
	encoded1, _ := sendFrame1.Encode()

	wsMutex.Lock()
	_ = ws1Desktop.WriteMessage(websocket.BinaryMessage, encoded1)
	wsMutex.Unlock()

	// 1. Verify Desktop gets ACK
	var firstAckSeq uint64
	select {
	case ack := <-u1DesktopAcks:
		if ack.ClientMsgId != clientMsgID1 || ack.Code != 0 {
			log.Fatalf("[FAIL] Desktop received invalid ACK: %+v", ack)
		}
		firstAckSeq = ack.SeqId
		log.Printf("[PASS] User 1 Desktop received ACK (Seq: %d, MsgID: %d)", ack.SeqId, ack.MsgId)
	case <-time.After(2 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for ACK on Desktop")
	}

	// 2. Verify Desktop receives NO push echo
	select {
	case echoMsg := <-u1DesktopPushes:
		log.Fatalf("[FAIL] Desktop received unexpected echo push: %+v", echoMsg)
	case <-time.After(300 * time.Millisecond):
		log.Println("[PASS] Echo suppression verified: Desktop received 0 duplicate echo")
	}

	// 3. Verify Android receives multi-device sync push
	select {
	case syncPush := <-u1AndroidPushes:
		if syncPush.SenderId != 1 || syncPush.Content != "Hello Sarah from my Desktop!" {
			log.Fatalf("[FAIL] Android received invalid multi-device sync: %+v", syncPush)
		}
		log.Printf("[PASS] User 1 Android received real-time sync push (Sender: %d, Content: %q, Seq: %d)",
			syncPush.SenderId, syncPush.Content, syncPush.SeqId)
	case <-time.After(2 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for multi-device sync push on Android")
	}

	// 4. Verify Receiver (User 2 Phone) receives the message
	select {
	case recvPush := <-u2PhonePushes:
		if recvPush.SenderId != 1 || recvPush.Content != "Hello Sarah from my Desktop!" {
			log.Fatalf("[FAIL] User 2 Phone received invalid message: %+v", recvPush)
		}
		log.Printf("[PASS] User 2 Phone received chat push (Sender: %d, Content: %q, Seq: %d)",
			recvPush.SenderId, recvPush.Content, recvPush.SeqId)
	case <-time.After(2 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for push on User 2 Phone")
	}

	// =========================================================================
	// TEST CASE 2: Offline Disconnection & Incremental Sync
	// =========================================================================
	log.Println("\n--- Test Case 2: Offline Disconnect & Incremental Synchronization ---")

	// Disconnect User 1 Android to simulate going offline
	log.Println("Simulating User 1 Android disconnecting (entering elevator / tunnel)...")
	_ = ws1Android.Close()
	time.Sleep(300 * time.Millisecond)

	// While Android is offline:
	// Desktop sends 2 messages
	for i := 1; i <= 2; i++ {
		cid := fmt.Sprintf("u1_offline_test_%d_%d", i, time.Now().UnixNano())
		msg := protocol.ChatMessage{
			ClientMsgId: cid,
			ConvType:    protocol.ConversationType_SINGLE,
			ReceiverId:  2,
			MsgType:     protocol.MessageType_TEXT,
			Content:     fmt.Sprintf("Offline batch msg %d from Desktop", i),
		}
		b, _ := json.Marshal(&msg)
		f := protocol.NewFrame(protocol.CmdChatMessageSend, uint64(i+1), protocol.FlagJSON, b)
		enc, _ := f.Encode()
		_ = ws1Desktop.WriteMessage(websocket.BinaryMessage, enc)
		<-u1DesktopAcks
	}

	// User 2 sends 2 replies back to User 1
	for i := 1; i <= 2; i++ {
		cid := fmt.Sprintf("u2_reply_%d_%d", i, time.Now().UnixNano())
		msg := protocol.ChatMessage{
			ClientMsgId: cid,
			ConvType:    protocol.ConversationType_SINGLE,
			ReceiverId:  1,
			MsgType:     protocol.MessageType_TEXT,
			Content:     fmt.Sprintf("Reply %d from Sarah while you were away", i),
		}
		b, _ := json.Marshal(&msg)
		f := protocol.NewFrame(protocol.CmdChatMessageSend, uint64(i+1), protocol.FlagJSON, b)
		enc, _ := f.Encode()
		_ = ws2Phone.WriteMessage(websocket.BinaryMessage, enc)
	}

	time.Sleep(500 * time.Millisecond)
	log.Println("[INFO] 4 new messages generated while User 1 Android was offline.")

	// Now User 1 Android comes back online!
	// It knows its localMaxSeq was 1 (the first message it received before disconnect).
	// It queries HTTP /api/chat/messages?conv_id=1_2&last_seq_id=1
	convID := "1_2"
	lastSeqID := firstAckSeq
	reqURL := fmt.Sprintf("%s/api/chat/messages?conv_id=%s&last_seq_id=%d&limit=50", serverHttpURL, convID, lastSeqID)

	req, _ := http.NewRequest("GET", reqURL, nil)
	req.Header.Set("Authorization", "Bearer "+p1Android.AccessToken)
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("[FAIL] HTTP Sync request failed: %v", err)
	}
	defer httpResp.Body.Close()

	bodyBytes, _ := io.ReadAll(httpResp.Body)
	var apiRes struct {
		Code    int               `json:"code"`
		Message string            `json:"message"`
		Data    []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(bodyBytes, &apiRes); err != nil {
		log.Fatalf("[FAIL] Failed to unmarshal API response: %v\nBody: %s", err, string(bodyBytes))
	}

	if apiRes.Code != 0 {
		log.Fatalf("[FAIL] HTTP sync returned non-zero code: %d, msg: %s", apiRes.Code, apiRes.Message)
	}

	log.Printf("[PASS] HTTP incremental sync returned %d messages (Expected: 4)", len(apiRes.Data))
	if len(apiRes.Data) != 4 {
		log.Fatalf("[FAIL] Expected 4 missed messages, got %d", len(apiRes.Data))
	}

	// Parse messages to verify strictly ascending seq_id and gap-freedom
	type SyncedMsg struct {
		ID      uint64 `json:"id"`
		SeqID   uint64 `json:"seq_id"`
		Content string `json:"content"`
		Sender  uint64 `json:"sender_id"`
	}

	var synced []SyncedMsg
	for _, raw := range apiRes.Data {
		var m SyncedMsg
		_ = json.Unmarshal(raw, &m)
		synced = append(synced, m)
	}

	expectedSeq := firstAckSeq + 1
	for idx, m := range synced {
		log.Printf("  [%d] SeqID: %d, Sender: %d, Content: %q", idx+1, m.SeqID, m.Sender, m.Content)
		if m.SeqID != expectedSeq {
			log.Fatalf("[FAIL] Sequence gap or reordering! Expected seq %d, got %d", expectedSeq, m.SeqID)
		}
		expectedSeq++
	}
	log.Println("[PASS] Incremental sync sequence validation: 100% gap-free and monotonically ordered!")

	// =========================================================================
	// TEST CASE 3: Initial Load (last_seq_id = 0 returns latest messages)
	// =========================================================================
	log.Println("\n--- Test Case 3: Initial Load / Latest Messages (last_seq_id = 0) ---")

	initURL := fmt.Sprintf("%s/api/chat/messages?conv_id=%s&limit=3", serverHttpURL, convID)
	initReq, _ := http.NewRequest("GET", initURL, nil)
	initReq.Header.Set("Authorization", "Bearer "+p1Android.AccessToken)
	initResp, err := http.DefaultClient.Do(initReq)
	if err != nil {
		log.Fatalf("[FAIL] Initial load request failed: %v", err)
	}
	defer initResp.Body.Close()

	initBytes, _ := io.ReadAll(initResp.Body)
	var initApiRes struct {
		Code int         `json:"code"`
		Data []SyncedMsg `json:"data"`
	}
	_ = json.Unmarshal(initBytes, &initApiRes)

	log.Printf("[PASS] Initial load with limit=3 returned %d messages", len(initApiRes.Data))
	if len(initApiRes.Data) != 3 {
		log.Fatalf("[FAIL] Expected 3 latest messages, got %d", len(initApiRes.Data))
	}

	latestSeq := firstAckSeq + 4
	if initApiRes.Data[0].SeqID != latestSeq-2 || initApiRes.Data[1].SeqID != latestSeq-1 || initApiRes.Data[2].SeqID != latestSeq {
		log.Fatalf("[FAIL] Initial load did not return latest messages chronologically! Expected seqs %d, %d, %d, Got seqs: %d, %d, %d",
			latestSeq-2, latestSeq-1, latestSeq,
			initApiRes.Data[0].SeqID, initApiRes.Data[1].SeqID, initApiRes.Data[2].SeqID)
	}
	log.Printf("[PASS] Initial load returned latest messages: seq %d -> %d -> %d in perfect order!",
		initApiRes.Data[0].SeqID, initApiRes.Data[1].SeqID, initApiRes.Data[2].SeqID)

	// =========================================================================
	// TEST CASE 4: WebSocket Binary Protocol Sync (CmdSyncReq -> CmdSyncResp)
	// =========================================================================
	log.Println("\n--- Test Case 4: WebSocket Binary Protocol Incremental Sync (CmdSyncReq) ---")

	// Reconnect User 1 Android on WebSocket
	ws1AndroidSync, _, err := websocket.DefaultDialer.Dial(u1AndroidURL, nil)
	if err != nil {
		log.Fatalf("Failed to reconnect ws1Android: %v", err)
	}
	defer ws1AndroidSync.Close()

	// Send CmdSyncReq for conv 1_2 from seq 9
	syncReq := protocol.SyncRequest{
		ConversationId: "1_2",
		LastSeqId:      firstAckSeq,
		Limit:          50,
	}
	syncReqBytes, _ := json.Marshal(&syncReq)
	syncFrame := protocol.NewFrame(protocol.CmdSyncReq, firstAckSeq, protocol.FlagJSON, syncReqBytes)
	encSyncFrame, _ := syncFrame.Encode()

	err = ws1AndroidSync.WriteMessage(websocket.BinaryMessage, encSyncFrame)
	if err != nil {
		log.Fatalf("Failed to send CmdSyncReq: %v", err)
	}

	// Read CmdSyncResp
	_ = ws1AndroidSync.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, syncRespData, err := ws1AndroidSync.ReadMessage()
	if err != nil {
		log.Fatalf("Failed to read CmdSyncResp: %v", err)
	}

	syncRespFrame, err := protocol.Decode(syncRespData)
	if err != nil {
		log.Fatalf("Failed to decode sync response frame: %v", err)
	}

	if syncRespFrame.CmdType != protocol.CmdSyncResp {
		log.Fatalf("Expected CmdSyncResp (0x%04x), got 0x%04x", protocol.CmdSyncResp, syncRespFrame.CmdType)
	}

	var wsSyncResp protocol.SyncResponse
	if err := json.Unmarshal(syncRespFrame.Payload, &wsSyncResp); err != nil {
		log.Fatalf("Failed to unmarshal CmdSyncResp payload: %v", err)
	}

	log.Printf("[PASS] WebSocket CmdSyncResp returned %d messages (LatestSeqId: %d)",
		len(wsSyncResp.Messages), wsSyncResp.LatestSeqId)

	if len(wsSyncResp.Messages) != 4 {
		log.Fatalf("[FAIL] Expected 4 messages from WebSocket sync, got %d", len(wsSyncResp.Messages))
	}

	for idx, m := range wsSyncResp.Messages {
		log.Printf("  [WS-%d] SeqID: %d, Sender: %d, Content: %q", idx+1, m.SeqId, m.SenderId, m.Content)
	}
	log.Println("[PASS] WebSocket binary protocol incremental sync verified 100% operational!")

	// Close readers
	close(stopReaders)

	log.Println("\n================================================================")
	log.Println("   ALL MULTI-DEVICE & INCREMENTAL SYNC TESTS PASSED (100%)      ")
	log.Println("================================================================")
}
