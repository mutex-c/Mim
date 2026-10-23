package chat

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 10 * 1024 * 1024 // 10MB
)

// Client represents a single active WebSocket connection
type Client struct {
	UserID   uint64
	Platform string // web, windows, android
	DeviceID string
	Conn     *websocket.Conn
	SendChan chan []byte
	Hub      *Hub
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

// Close gracefully closes the client
func (c *Client) Close() {
	c.closeOnce.Do(func() {
		close(c.SendChan)
		if c.Conn != nil {
			_ = c.Conn.Close()
		}
	})
}
