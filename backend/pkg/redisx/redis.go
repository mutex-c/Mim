package redisx

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr     string
	Password string
	DB       int
}

type Client struct {
	rdb *redis.Client
}

// NewClient initializes and pings a Redis connection
func NewClient(cfg Config) (*Client, error) {
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     64,
		MinIdleConns: 8,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed (%s): %w", cfg.Addr, err)
	}

	return &Client{rdb: rdb}, nil
}

// Underlying returns the raw go-redis client
func (c *Client) Underlying() *redis.Client {
	if c == nil {
		return nil
	}
	return c.rdb
}

// Close gracefully closes the Redis connection pool
func (c *Client) Close() error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Close()
}

// NextSeqID allocates a strictly monotonically increasing sequence number using Redis INCR.
// fallback can be a uint64 or a lazy evaluation func() uint64.
// If the key was newly created by this Incr (val == 1), fallback is invoked to seed the key
// with the existing DB max seq_id. Subsequent calls purely INCR without touching the database.
func (c *Client) NextSeqID(ctx context.Context, convID string, fallback any) (uint64, error) {
	if c == nil || c.rdb == nil {
		return 0, fmt.Errorf("redis client is nil")
	}

	key := fmt.Sprintf("mim:conv:%s:seq", convID)

	val, err := c.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	// If val == 1, key was newly created by this Incr. Seed with existing DB max seq if available.
	if val == 1 && fallback != nil {
		var initVal uint64
		switch fn := fallback.(type) {
		case func() uint64:
			initVal = fn()
		case uint64:
			initVal = fn
		case int:
			initVal = uint64(fn)
		case int64:
			initVal = uint64(fn)
		}

		if initVal > 0 {
			reseedScript := redis.NewScript(`
				local curr = redis.call('GET', KEYS[1])
				if curr == '1' then
					redis.call('SET', KEYS[1], ARGV[1])
					return tonumber(ARGV[1])
				else
					return tonumber(curr)
				end
			`)
			res, err := reseedScript.Run(ctx, c.rdb, []string{key}, initVal+1).Int64()
			if err == nil {
				return uint64(res), nil
			}
		}
	}

	return uint64(val), nil
}

// DeviceSession represents an active device session stored in Redis
type DeviceSession struct {
	DeviceID    string    `json:"device_id"`
	Platform    string    `json:"platform"`
	GatewayAddr string    `json:"gateway_addr,omitempty"`
	ConnectedAt time.Time `json:"connected_at"`
}

// RegisterDevice registers an online device for a user under a specific platform.
// Returns previous deviceID on the same platform if one existed.
func (c *Client) RegisterDevice(ctx context.Context, userID uint64, platform, deviceID string) (string, error) {
	return c.RegisterDeviceWithGateway(ctx, userID, platform, deviceID, "")
}

// RegisterDeviceWithGateway registers an online device and stores the gateway node address for targeted routing.
func (c *Client) RegisterDeviceWithGateway(ctx context.Context, userID uint64, platform, deviceID, gatewayAddr string) (string, error) {
	if c == nil || c.rdb == nil {
		return "", nil
	}

	key := fmt.Sprintf("mim:user:%d:devices", userID)
	oldData, err := c.rdb.HGet(ctx, key, platform).Result()
	var oldDeviceID string
	if err == nil && oldData != "" {
		var oldSess DeviceSession
		if err := json.Unmarshal([]byte(oldData), &oldSess); err == nil {
			oldDeviceID = oldSess.DeviceID
		} else {
			oldDeviceID = oldData
		}
	}

	sess := DeviceSession{
		DeviceID:    deviceID,
		Platform:    platform,
		GatewayAddr: gatewayAddr,
		ConnectedAt: time.Now(),
	}
	bytes, _ := json.Marshal(sess)

	pipe := c.rdb.Pipeline()
	pipe.HSet(ctx, key, platform, string(bytes))
	pipe.Expire(ctx, key, 7*24*time.Hour)
	_, err = pipe.Exec(ctx)
	return oldDeviceID, err
}

// UnregisterDevice removes a device session only if the registered deviceID matches.
func (c *Client) UnregisterDevice(ctx context.Context, userID uint64, platform, deviceID string) error {
	if c == nil || c.rdb == nil {
		return nil
	}

	key := fmt.Sprintf("mim:user:%d:devices", userID)
	data, err := c.rdb.HGet(ctx, key, platform).Result()
	if err != nil {
		return nil
	}

	var sess DeviceSession
	if err := json.Unmarshal([]byte(data), &sess); err == nil {
		if sess.DeviceID != deviceID {
			return nil // Another newer device already registered on this platform
		}
	} else if data != deviceID {
		return nil
	}

	return c.rdb.HDel(ctx, key, platform).Err()
}

// GetUserDevices returns all active platform -> deviceID mappings for a user
func (c *Client) GetUserDevices(ctx context.Context, userID uint64) (map[string]string, error) {
	if c == nil || c.rdb == nil {
		return nil, nil
	}

	key := fmt.Sprintf("mim:user:%d:devices", userID)
	data, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	res := make(map[string]string, len(data))
	for plat, val := range data {
		var sess DeviceSession
		if err := json.Unmarshal([]byte(val), &sess); err == nil {
			res[plat] = sess.DeviceID
		} else {
			res[plat] = val
		}
	}
	return res, nil
}

// GetUserDeviceSessions returns all active platform -> DeviceSession records for a user
func (c *Client) GetUserDeviceSessions(ctx context.Context, userID uint64) (map[string]DeviceSession, error) {
	if c == nil || c.rdb == nil {
		return nil, nil
	}

	key := fmt.Sprintf("mim:user:%d:devices", userID)
	data, err := c.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	res := make(map[string]DeviceSession, len(data))
	for plat, val := range data {
		var sess DeviceSession
		if err := json.Unmarshal([]byte(val), &sess); err == nil {
			res[plat] = sess
		} else {
			res[plat] = DeviceSession{DeviceID: val, Platform: plat}
		}
	}
	return res, nil
}

// KickoutPayload is published over Redis PubSub for distributed multi-gateway kickout
type KickoutPayload struct {
	UserID      uint64 `json:"user_id"`
	Platform    string `json:"platform"`
	NewDeviceID string `json:"new_device_id"`
	Reason      string `json:"reason"`
	KickTime    int64  `json:"kick_time"`
}

const KickoutChannel = "mim:channel:kickout"

// PublishKickout broadcasts a kickout event across all gateway instances
func (c *Client) PublishKickout(ctx context.Context, payload KickoutPayload) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return c.rdb.Publish(ctx, KickoutChannel, data).Err()
}

// SubscribeKickout listens for kickout events and calls the callback
func (c *Client) SubscribeKickout(ctx context.Context, onKick func(p KickoutPayload)) {
	if c == nil || c.rdb == nil {
		return
	}
	pubsub := c.rdb.Subscribe(ctx, KickoutChannel)
	go func() {
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
			var p KickoutPayload
			if err := json.Unmarshal([]byte(msg.Payload), &p); err == nil {
				onKick(p)
			} else {
				log.Printf("[Redis Kickout] Parse error: %v", err)
			}
		}
	}()
}
