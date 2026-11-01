package chat

import (
	"context"
	"log"
	"sync"
	"time"

	"mim/pkg/redisx"
)

// Hub manages all active WebSocket connections across platforms with Redis distributed routing & single-platform kickout
type Hub struct {
	// Map: UserID -> map[*Client]bool
	clients     map[uint64]map[*Client]bool
	mu          sync.RWMutex
	redisClient *redisx.Client
	gatewayAddr string
}

func NewHub(rdb *redisx.Client) *Hub {
	h := &Hub{
		clients:     make(map[uint64]map[*Client]bool),
		redisClient: rdb,
	}

	// In multi-gateway cluster mode, subscribe to Redis kickout channel
	if rdb != nil {
		rdb.SubscribeKickout(context.Background(), func(p redisx.KickoutPayload) {
			h.KickLocalDevice(p.UserID, p.Platform, p.NewDeviceID, p.Reason)
		})
	}

	return h
}

// SetGatewayAddr configures the internal gRPC address of this gateway instance
func (h *Hub) SetGatewayAddr(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.gatewayAddr = addr
}

func (h *Hub) GetGatewayAddr() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.gatewayAddr
}

// KickLocalDevice kicks any locally connected client matching (userID, platform) except newDeviceID
func (h *Hub) KickLocalDevice(userID uint64, platform, newDeviceID, reason string) {
	h.mu.Lock()
	devices, ok := h.clients[userID]
	if !ok || len(devices) == 0 {
		h.mu.Unlock()
		return
	}

	var toKick []*Client
	for c := range devices {
		if c.Platform == platform && c.DeviceID != newDeviceID {
			toKick = append(toKick, c)
			delete(devices, c)
		}
	}
	if len(devices) == 0 {
		delete(h.clients, userID)
	}
	h.mu.Unlock()

	for _, c := range toKick {
		log.Printf("[Hub Kickout] Kicking user %d device %s on %s (New device: %s)", userID, c.DeviceID, platform, newDeviceID)
		c.Kick(reason, newDeviceID)
	}
}

// Register adds a client connection to the hub, enforcing same-platform kickout (FR-AUTH-03)
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	devices, ok := h.clients[c.UserID]
	if !ok {
		devices = make(map[*Client]bool)
		h.clients[c.UserID] = devices
	}

	// 1. Same-process single-platform kick: find any existing client on the same platform
	var localOldClients []*Client
	for oldClient := range devices {
		if oldClient.Platform == c.Platform && oldClient.DeviceID != c.DeviceID {
			localOldClients = append(localOldClients, oldClient)
			delete(devices, oldClient)
		}
	}
	devices[c] = true
	h.mu.Unlock()

	// Kick local previous clients
	for _, oldClient := range localOldClients {
		log.Printf("[Hub Register] Kicking local previous connection for user %d platform %s (old: %s, new: %s)",
			c.UserID, c.Platform, oldClient.DeviceID, c.DeviceID)
		oldClient.Kick("Account logged in on another device", c.DeviceID)
	}

	// 2. Distributed single-platform kick via Redis
	if h.redisClient != nil {
		go func(uid uint64, plat, dev, gwAddr string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			oldDevID, err := h.redisClient.RegisterDeviceWithGateway(ctx, uid, plat, dev, gwAddr)
			if err == nil && oldDevID != "" && oldDevID != dev {
				// Broadcast kickout to all gateway nodes in the cluster
				_ = h.redisClient.PublishKickout(ctx, redisx.KickoutPayload{
					UserID:      uid,
					Platform:    plat,
					NewDeviceID: dev,
					Reason:      "Account logged in on another device",
					KickTime:    time.Now().UnixMilli(),
				})
			}
		}(c.UserID, c.Platform, c.DeviceID, h.GetGatewayAddr())
	}
}

// Unregister removes a client connection from the hub
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if devices, ok := h.clients[c.UserID]; ok {
		delete(devices, c)
		if len(devices) == 0 {
			delete(h.clients, c.UserID)
		}
	}
	h.mu.Unlock()

	// Remove from Redis device registry if still matching
	if h.redisClient != nil {
		go func(uid uint64, plat, dev string) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = h.redisClient.UnregisterDevice(ctx, uid, plat, dev)
		}(c.UserID, c.Platform, c.DeviceID)
	}
}

// SendToUser dispatches a raw binary packet to all active devices/tabs of a user
func (h *Hub) SendToUser(userID uint64, data []byte) bool {
	h.mu.RLock()
	devices, ok := h.clients[userID]
	if !ok || len(devices) == 0 {
		h.mu.RUnlock()
		return false
	}

	targets := make([]*Client, 0, len(devices))
	for client := range devices {
		targets = append(targets, client)
	}
	h.mu.RUnlock()

	sentCount := 0
	for _, client := range targets {
		if client.Send(data) {
			sentCount++
		}
	}
	return sentCount > 0
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
	for client := range devices {
		if excludeDeviceID == "" || client.DeviceID != excludeDeviceID {
			targets = append(targets, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range targets {
		client.Send(data)
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
	for client := range devices {
		platforms = append(platforms, client.Platform)
	}
	return platforms
}
