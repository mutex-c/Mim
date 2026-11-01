// Command cluster_smoke performs a real end-to-end delivery check against a running
// Mim microservice cluster (gateway + logic + transfer + Kafka + Redis).
//
// It mints two short-lived JWTs, opens two WebSocket connections against the gateway,
// sends one chat message from user A to user B, and asserts that B actually receives
// the pushed frame. It exits non-zero on failure, so it can gate CI or a deploy.
//
// Usage:
//
//	go run ./cmd/tools/cluster_smoke -gw ws://127.0.0.1:8081/ws -uid-a 101 -uid-b 202
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"mim/pkg/jwtx"
	"mim/protocol"
)

func main() {
	gwURL := flag.String("gw", "ws://127.0.0.1:8081/ws", "gateway websocket url")
	secret := flag.String("secret", "mim-enterprise-jwt-secret-key-2026", "jwt secret shared with the cluster")
	uidA := flag.Uint64("uid-a", 101, "sender user id")
	uidB := flag.Uint64("uid-b", 202, "receiver user id")
	group := flag.Bool("group", false, "treat uid-b as a group id and send a group message")
	wait := flag.Duration("timeout", 15*time.Second, "how long to wait for the push")
	flag.Parse()

	fail := func(format string, args ...interface{}) {
		fmt.Printf("SMOKE_FAIL: "+format+"\n", args...)
		os.Exit(1)
	}

	mint := func(uid uint64, dev string) string {
		tp, err := jwtx.GenerateTokenPair(*secret, time.Hour, 24*time.Hour, uid, fmt.Sprintf("user%d", uid), "web", dev)
		if err != nil {
			fail("mint token for user %d: %v", uid, err)
		}
		return tp.AccessToken
	}

	dial := func(token, dev string) *websocket.Conn {
		url := fmt.Sprintf("%s?token=%s&platform=web&device_id=%s", *gwURL, token, dev)
		conn, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			fail("dial gateway %s: %v", *gwURL, err)
		}
		return conn
	}

	receiver := dial(mint(*uidB, "smoke_receiver"), "smoke_receiver")
	defer receiver.Close()
	sender := dial(mint(*uidA, "smoke_sender"), "smoke_sender")
	defer sender.Close()

	// Give the gateway a moment to register both devices in Redis for routing.
	time.Sleep(500 * time.Millisecond)

	convID := fmt.Sprintf("%d_%d", *uidA, *uidB)
	if *uidA > *uidB {
		convID = fmt.Sprintf("%d_%d", *uidB, *uidA)
	}
	convType := protocol.ConversationType_SINGLE
	target := *uidB
	if *group {
		convID = fmt.Sprintf("group_%d", *uidB)
		convType = protocol.ConversationType_GROUP
	}

	msg := &protocol.ChatMessage{
		ClientMsgId:    fmt.Sprintf("smoke-%d", time.Now().UnixNano()),
		ConversationId: convID,
		ConvType:       convType,
		ReceiverId:     target,
		MsgType:        protocol.MessageType_TEXT,
		Content:        "cluster smoke test",
		SendTime:       time.Now().UnixMilli(),
	}
	payload, err := proto.Marshal(msg)
	if err != nil {
		fail("marshal message: %v", err)
	}
	frame := protocol.NewFrame(protocol.CmdChatMessageSend, 1, 0, payload)
	raw, err := frame.Encode()
	if err != nil {
		fail("encode frame: %v", err)
	}

	if err := sender.WriteMessage(websocket.BinaryMessage, raw); err != nil {
		fail("write message: %v", err)
	}

	_ = sender.SetReadDeadline(time.Now().Add(*wait))
	_, data, err := sender.ReadMessage()
	if err != nil {
		fail("no ACK from gateway: %v", err)
	}
	ackFrame, err := protocol.Decode(data)
	if err != nil {
		fail("decode ACK: %v", err)
	}
	var ack protocol.MessageAck
	_ = json.Unmarshal(ackFrame.Payload, &ack)
	if ackFrame.CmdType != protocol.CmdMessageAck || ack.Code != 0 {
		fail("unexpected ACK cmd=0x%04X code=%d err=%q", ackFrame.CmdType, ack.Code, ack.ErrorMsg)
	}
	fmt.Printf("ACK  cmd=0x%04X seq=%d\n", ackFrame.CmdType, ack.SeqId)

	_ = receiver.SetReadDeadline(time.Now().Add(*wait))
	_, data, err = receiver.ReadMessage()
	if err != nil {
		fail("message was persisted but never pushed to the receiver (%v). Check: mim-transfer partitions, Kafka topic, Redis device routing", err)
	}
	pushFrame, err := protocol.Decode(data)
	if err != nil {
		fail("decode push: %v", err)
	}
	if pushFrame.CmdType != protocol.CmdChatMessagePush {
		fail("expected CmdChatMessagePush (0x0101), got 0x%04X", pushFrame.CmdType)
	}

	fmt.Printf("PUSH cmd=0x%04X len=%d\n", pushFrame.CmdType, len(pushFrame.Payload))
	fmt.Println("SMOKE_PASS")
}
