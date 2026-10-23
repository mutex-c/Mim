package chat

import (
	"sync"
)

// Hub manages all active WebSocket connections across platforms
type Hub struct {
	// Map: UserID -> map[DeviceID]*Client
	clients map[uint64]map[string]*Client
	mu      sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uint64]map[string]*Client),
	}
}

// Register adds a client connection to the hub
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	devices, ok := h.clients[c.UserID]
	if !ok {
		devices = make(map[string]*Client)
		h.clients[c.UserID] = devices
	}

	// If device already connected, close old and replace
	if old, exists := devices[c.DeviceID]; exists {
		old.Close()
	}

	devices[c.DeviceID] = c
}

// Unregister removes a client connection from the hub
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if devices, ok := h.clients[c.UserID]; ok {
		if curr, exists := devices[c.DeviceID]; exists && curr == c {
			delete(devices, c.DeviceID)
			c.Close()
		}
		if len(devices) == 0 {
			delete(h.clients, c.UserID)
		}
	}
}

// SendToUser dispatches a raw binary packet to all active devices of a user
func (h *Hub) SendToUser(userID uint64, data []byte) bool {
	h.mu.RLock()
	devices, ok := h.clients[userID]
	if !ok || len(devices) == 0 {
		h.mu.RUnlock()
		return false
	}

	// Copy clients to avoid holding read lock while channel sending
	targets := make([]*Client, 0, len(devices))
	for _, client := range devices {
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	for _, client := range targets {
		select {
		case client.SendChan <- data:
		default:
			// Client send channel full, drop or handle gracefully
		}
	}
	return true
}

// SendToUserExclude sends data to a user's other devices (e.g. multi-terminal sync)
func (h *Hub) SendToUserExclude(userID uint64, excludeDeviceID string, data []byte) {
	h.mu.RLock()
	devices, ok := h.clients[userID]
	if !ok || len(devices) == 0 {
		h.mu.RUnlock()
		return
	}

	targets := make([]*Client, 0, len(devices))
	for devID, client := range devices {
		if devID != excludeDeviceID {
			targets = append(targets, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range targets {
		select {
		case client.SendChan <- data:
		default:
		}
	}
}

// IsOnline checks if a user has at least one active connection
func (h *Hub) IsOnline(userID uint64) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	devices, ok := h.clients[userID]
	return ok && len(devices) > 0
}

// OnlineCount returns total active online users
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetUserPlatforms returns a list of online platforms for a user (web, windows, android)
func (h *Hub) GetUserPlatforms(userID uint64) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	devices, ok := h.clients[userID]
	if !ok {
		return nil
	}

	platforms := make([]string, 0, len(devices))
	for _, client := range devices {
		platforms = append(platforms, client.Platform)
	}
	return platforms
}
