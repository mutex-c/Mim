package chat

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"mim/protocol"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10 * 1024 * 1024 // 10MB
)

// Client represents a single active WebSocket connection
type Client struct {
	UserID    uint64
	Platform  string // web, windows, android
	DeviceID  string
	Conn      *websocket.Conn
	SendChan  chan []byte
	Hub       *Hub
	closed    atomic.Bool
	sendMu    sync.RWMutex
	closeOnce sync.Once
}

func NewClient(userID uint64, platform, deviceID string, conn *websocket.Conn, hub *Hub) *Client {
	return &Client{
		UserID:   userID,
		Platform: platform,
		DeviceID: deviceID,
		Conn:     conn,
		SendChan: make(chan []byte, 256),
		Hub:      hub,
	}
}

// Send attempts to send binary data to client. Thread-safe against concurrent Close()
func (c *Client) Send(data []byte) bool {
	if c.closed.Load() {
		return false
	}
	c.sendMu.RLock()
	defer c.sendMu.RUnlock()

	if c.closed.Load() {
		return false
	}

	select {
	case c.SendChan <- data:
		return true
	default:
		return false
	}
}

// Close gracefully closes the client
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.sendMu.Lock()
		close(c.SendChan)
		c.sendMu.Unlock()
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	})
}

// Kick sends a CmdKickOut notification frame to the client and closes the connection
func (c *Client) Kick(reason, newDeviceID string) {
	notify := &protocol.KickOutNotify{
		Code:        4001,
		Reason:      reason,
		NewDeviceId: newDeviceID,
		KickTime:    time.Now().UnixMilli(),
	}
	payload, _ := json.Marshal(notify)
	frame := protocol.NewFrame(protocol.CmdKickOut, 0, protocol.FlagJSON, payload)
	if bytes, err := frame.Encode(); err == nil {
		c.Send(bytes)
	}

	// Graceful wait to let the frame flush out of WebSocket buffer before closing
	time.AfterFunc(200*time.Millisecond, func() {
		c.Close()
	})
}

