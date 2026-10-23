package log_test

import (
	"sync/atomic"
	"testing"
	"time"

	"mim/service/log"
)

func TestAsyncLogProductionAndConsumption(t *testing.T) {
	svc := log.NewService(100)

	var consumedCount int64
	svc.StartConsumer(2, func(event log.LogEvent) {
		atomic.AddInt64(&consumedCount, 1)
	})

	// Produce 10 events
	for i := 0; i < 10; i++ {
		ok := svc.Produce(log.LogEvent{
			TraceID:    "trace-test-uuid",
			UserID:     uint64(i + 1),
			Account:    "user_test",
			Service:    "mim-gateway",
			Action:     "HTTP /api/v1/auth/login",
			ClientIP:   "127.0.0.1",
			StatusCode: 200,
			DurationMs: 15,
		})
		if !ok {
			t.Errorf("failed to produce event %d", i)
		}
	}

	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadInt64(&consumedCount)
	if count != 10 {
		t.Errorf("expected 10 events consumed, got: %d", count)
	}

	// Test GeoIP
	locLoopback := log.ResolveGeoIP("127.0.0.1")
	if locLoopback != "本地回环 (127.0.0.1)" {
		t.Errorf("unexpected geo ip: %s", locLoopback)
	}

	locLAN := log.ResolveGeoIP("192.168.1.100")
	if locLAN != "内网局域网 (Private LAN)" {
		t.Errorf("unexpected geo ip: %s", locLAN)
	}

	svc.Stop()
}
