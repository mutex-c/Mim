package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"mim/pkg/jwtx"
	"mim/protocol"

	"github.com/gorilla/websocket"
)

const (
	serverWsURL = "ws://127.0.0.1:8080/ws"
	jwtSecret   = "mim-enterprise-jwt-secret-key-2026"
)

func main() {
	log.Println("=== Starting Automated WebRTC Signaling End-to-End Test ===")

	// 1. Generate Auth Tokens for User 1 (Alex) and User 2 (Sarah)
	p1, err := jwtx.GenerateTokenPair(jwtSecret, 24*time.Hour, 24*time.Hour, 1, "alex", "desktop", "dev_desktop_1")
	if err != nil {
		log.Fatalf("Failed to generate token 1: %v", err)
	}
	token1 := p1.AccessToken

	p2, err := jwtx.GenerateTokenPair(jwtSecret, 24*time.Hour, 24*time.Hour, 2, "sarah", "android", "dev_android_2")
	if err != nil {
		log.Fatalf("Failed to generate token 2: %v", err)
	}
	token2 := p2.AccessToken

	// 2. Connect User 1 (Desktop) via WebSocket
	url1 := fmt.Sprintf("%s?user_id=1&token=%s&platform=desktop&device_id=dev_desktop_1", serverWsURL, token1)
	ws1, _, err := websocket.DefaultDialer.Dial(url1, nil)
	if err != nil {
		log.Fatalf("User 1 failed to connect WebSocket: %v", err)
	}
	defer ws1.Close()
	log.Println("[PASS] User 1 (Desktop) connected to WebSocket")

	// 3. Connect User 2 (Android) via WebSocket
	url2 := fmt.Sprintf("%s?user_id=2&token=%s&platform=android&device_id=dev_android_2", serverWsURL, token2)
	ws2, _, err := websocket.DefaultDialer.Dial(url2, nil)
	if err != nil {
		log.Fatalf("User 2 failed to connect WebSocket: %v", err)
	}
	defer ws2.Close()
	log.Println("[PASS] User 2 (Android) connected to WebSocket")

	callID := fmt.Sprintf("test_call_%d", time.Now().Unix())
	offerReceivedChan := make(chan bool, 1)
	answerReceivedChan := make(chan bool, 1)
	candidateReceivedChan := make(chan bool, 1)
	hangupReceivedChan := make(chan bool, 1)

	// User 2 Message Reader (Simulating Android receiver)
	go func() {
		for {
			_, data, err := ws2.ReadMessage()
			if err != nil {
				return
			}
			frame, err := protocol.Decode(data)
			if err != nil {
				continue
			}
			if frame.CmdType == protocol.CmdWebRTCSignal {
				var signal protocol.WebRTCSignal
				_ = json.Unmarshal(frame.Payload, &signal)
				log.Printf("[Android / User 2] Received WebRTC Signal: %s from User %d (CallID: %s)", signal.SignalType, signal.FromUserId, signal.CallId)

				switch signal.SignalType {
				case "OFFER":
					offerReceivedChan <- true
					// Simulate Android answering the call with an ANSWER SDP
					answerData := `{"sdp":{"type":"answer","sdp":"v=0\r\no=- 2 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\n"}}`
					ansSignal := protocol.WebRTCSignal{
						CallId:     signal.CallId,
						FromUserId: 2,
						ToUserId:   1,
						SignalType: "ANSWER",
						Data:       answerData,
					}
					payload, _ := json.Marshal(&ansSignal)
					f := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, payload)
					encoded, _ := f.Encode()
					_ = ws2.WriteMessage(websocket.BinaryMessage, encoded)
					log.Printf("[Android / User 2] Sent WebRTC ANSWER to User 1")

				case "CANDIDATE":
					candidateReceivedChan <- true

				case "HANGUP":
					hangupReceivedChan <- true
					return
				}
			}
		}
	}()

	// User 1 Message Reader (Simulating Desktop caller)
	go func() {
		for {
			_, data, err := ws1.ReadMessage()
			if err != nil {
				return
			}
			frame, err := protocol.Decode(data)
			if err != nil {
				continue
			}
			if frame.CmdType == protocol.CmdWebRTCSignal {
				var signal protocol.WebRTCSignal
				_ = json.Unmarshal(frame.Payload, &signal)
				log.Printf("[Desktop / User 1] Received WebRTC Signal: %s from User %d (CallID: %s)", signal.SignalType, signal.FromUserId, signal.CallId)

				if signal.SignalType == "ANSWER" {
					answerReceivedChan <- true
				}
			}
		}
	}()

	time.Sleep(200 * time.Millisecond)

	// Step 4: User 1 sends OFFER (HD Video Call) to User 2
	offerData := `{"sdp":{"type":"offer","sdp":"v=0\r\no=- 1 2 IN IP4 127.0.0.1\r\ns=-\r\nt=0 0\r\nm=video 9 UDP/TLS/RTP/SAVPF 96\r\n"},"callType":"video"}`
	offerSignal := protocol.WebRTCSignal{
		CallId:     callID,
		FromUserId: 1,
		ToUserId:   2,
		SignalType: "OFFER",
		Data:       offerData,
	}
	payload, _ := json.Marshal(&offerSignal)
	f := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, payload)
	encoded, _ := f.Encode()
	_ = ws1.WriteMessage(websocket.BinaryMessage, encoded)
	log.Printf("[Desktop / User 1] Sent WebRTC OFFER to User 2 (CallID: %s)", callID)

	// Wait for User 2 to receive OFFER
	select {
	case <-offerReceivedChan:
		log.Println("[PASS] Step 1: User 2 successfully received OFFER")
	case <-time.After(3 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for User 2 to receive OFFER")
	}

	// Wait for User 1 to receive ANSWER
	select {
	case <-answerReceivedChan:
		log.Println("[PASS] Step 2: User 1 successfully received ANSWER")
	case <-time.After(3 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for User 1 to receive ANSWER")
	}

	// Step 5: User 1 sends ICE CANDIDATE to User 2
	candData := `{"candidate":"candidate:1 1 UDP 2122260223 127.0.0.1 54321 typ host","sdpMid":"0","sdpMLineIndex":0}`
	candSignal := protocol.WebRTCSignal{
		CallId:     callID,
		FromUserId: 1,
		ToUserId:   2,
		SignalType: "CANDIDATE",
		Data:       candData,
	}
	candPayload, _ := json.Marshal(&candSignal)
	candFrame := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, candPayload)
	candEncoded, _ := candFrame.Encode()
	_ = ws1.WriteMessage(websocket.BinaryMessage, candEncoded)
	log.Printf("[Desktop / User 1] Sent ICE CANDIDATE to User 2")

	// Wait for User 2 to receive CANDIDATE
	select {
	case <-candidateReceivedChan:
		log.Println("[PASS] Step 3: User 2 successfully received ICE CANDIDATE")
	case <-time.After(3 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for User 2 to receive CANDIDATE")
	}

	// Step 6: User 1 hangs up call after 3 seconds of conversation
	time.Sleep(500 * time.Millisecond)
	hangupData := `{"callType":"video","duration":18}`
	hangupSignal := protocol.WebRTCSignal{
		CallId:     callID,
		FromUserId: 1,
		ToUserId:   2,
		SignalType: "HANGUP",
		Data:       hangupData,
	}
	hangupPayload, _ := json.Marshal(&hangupSignal)
	hangupFrame := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, hangupPayload)
	hangupEncoded, _ := hangupFrame.Encode()
	_ = ws1.WriteMessage(websocket.BinaryMessage, hangupEncoded)
	log.Printf("[Desktop / User 1] Sent HANGUP signal to User 2")

	// Wait for User 2 to receive HANGUP
	select {
	case <-hangupReceivedChan:
		log.Println("[PASS] Step 4: User 2 successfully received HANGUP")
	case <-time.After(3 * time.Second):
		log.Fatalf("[FAIL] Timeout waiting for User 2 to receive HANGUP")
	}

	// Step 7: Verify Call History in Backend HTTP API
	time.Sleep(300 * time.Millisecond)
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080/api/calls/history", nil)
	req.Header.Set("Authorization", "Bearer "+token1)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Failed to query call history API: %v", err)
	}
	defer resp.Body.Close()

	var historyResp struct {
		Code int `json:"code"`
		Data []struct {
			ID        string `json:"id"`
			FromUser  string `json:"from_user"`
			ToUser    string `json:"to_user"`
			CallType  string `json:"call_type"`
			Status    string `json:"status"`
			Duration  int    `json:"duration"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&historyResp); err != nil {
		log.Fatalf("Failed to decode call history response: %v", err)
	}

	found := false
	for _, rec := range historyResp.Data {
		if rec.ID == callID {
			found = true
			log.Printf("[PASS] Step 5: Found Call Record in SQLite database!")
			log.Printf("       CallID: %s, From: %s, To: %s, Type: %s, Status: %s, Duration: %ds",
				rec.ID, rec.FromUser, rec.ToUser, rec.CallType, rec.Status, rec.Duration)
			break
		}
	}

	if !found {
		log.Fatalf("[FAIL] Call Record with CallID %s not found in history response", callID)
	}

	log.Println("==================================================")
	log.Println("🎉 ALL WEBRTC SIGNALING TESTS PASSED PERFECTLY! 🎉")
	log.Println("==================================================")
}
