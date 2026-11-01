package redisx

import (
	"context"
	"testing"
)

func TestRedisXNilSafe(t *testing.T) {
	var c *Client = nil

	if c.Underlying() != nil {
		t.Fatal("expected nil underlying client")
	}

	if err := c.Close(); err != nil {
		t.Fatalf("expected nil close error, got: %v", err)
	}

	ctx := context.Background()
	_, err := c.NextSeqID(ctx, "test_conv", 10)
	if err == nil {
		t.Fatal("expected error for nil client NextSeqID")
	}

	oldDev, err := c.RegisterDevice(ctx, 1, "web", "dev_1")
	if err != nil || oldDev != "" {
		t.Fatalf("expected empty result without error, got: %v, %s", err, oldDev)
	}

	if err := c.UnregisterDevice(ctx, 1, "web", "dev_1"); err != nil {
		t.Fatalf("unexpected unregister error: %v", err)
	}

	devs, err := c.GetUserDevices(ctx, 1)
	if err != nil || devs != nil {
		t.Fatalf("expected nil devices, got: %v, %v", err, devs)
	}

	if err := c.PublishKickout(ctx, KickoutPayload{}); err != nil {
		t.Fatalf("expected nil publish error, got: %v", err)
	}
}

func TestRedisXWithLiveServer(t *testing.T) {
	// Connect to local redis with password 'redis'
	c, err := NewClient(Config{
		Addr:     "127.0.0.1:6379",
		Password: "redis",
		DB:       15, // Test DB
	})
	if err != nil {
		t.Skipf("Live redis not available, skipping integration test: %v", err)
		return
	}
	defer c.Close()

	ctx := context.Background()
	testConv := "test_conv_seq_1"
	// Clean up key first
	_ = c.rdb.Del(ctx, "mim:conv:"+testConv+":seq").Err()

	s1, err := c.NextSeqID(ctx, testConv, 100)
	if err != nil {
		t.Fatalf("NextSeqID failed: %v", err)
	}
	if s1 != 101 {
		t.Fatalf("expected seq 101, got %d", s1)
	}

	s2, err := c.NextSeqID(ctx, testConv, 100)
	if err != nil {
		t.Fatalf("NextSeqID 2 failed: %v", err)
	}
	if s2 != 102 {
		t.Fatalf("expected seq 102, got %d", s2)
	}

	// Test Device Registration & Kickout
	uid := uint64(99999)
	oldDev, err := c.RegisterDevice(ctx, uid, "web", "device_A")
	if err != nil {
		t.Fatalf("RegisterDevice failed: %v", err)
	}
	if oldDev != "" {
		t.Fatalf("expected empty old device, got %s", oldDev)
	}

	// Register device B on same platform (web)
	oldDev2, err := c.RegisterDevice(ctx, uid, "web", "device_B")
	if err != nil {
		t.Fatalf("RegisterDevice 2 failed: %v", err)
	}
	if oldDev2 != "device_A" {
		t.Fatalf("expected old device 'device_A', got '%s'", oldDev2)
	}

	// Register device on mobile platform
	_, err = c.RegisterDevice(ctx, uid, "android", "device_Mobile")
	if err != nil {
		t.Fatalf("RegisterDevice mobile failed: %v", err)
	}

	devs, err := c.GetUserDevices(ctx, uid)
	if err != nil {
		t.Fatalf("GetUserDevices failed: %v", err)
	}
	if devs["web"] != "device_B" || devs["android"] != "device_Mobile" {
		t.Fatalf("unexpected devices: %+v", devs)
	}

	// Clean up test keys
	_ = c.rdb.Del(ctx, "mim:conv:"+testConv+":seq", "mim:user:99999:devices").Err()
}
