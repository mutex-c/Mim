package log

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"
)

type LogEvent struct {
	TraceID     string    `json:"trace_id"`
	UserID      uint64    `json:"user_id"`
	Account     string    `json:"account"`
	Service     string    `json:"service"`
	Action      string    `json:"action"`
	ClientIP    string    `json:"client_ip"`
	GeoLocation string    `json:"geo_location"`
	StatusCode  int       `json:"status_code"`
	DurationMs  int64     `json:"duration_ms"`
	CreatedAt   time.Time `json:"created_at"`
}

type Service struct {
	queue      chan LogEvent
	bufferSize int
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewService(bufferSize int) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		queue:      make(chan LogEvent, bufferSize),
		bufferSize: bufferSize,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// ResolveGeoIP parses IP to Location (internal/external classification)
func ResolveGeoIP(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "未知IP"
	}

	if ip.IsLoopback() {
		return "本地回环 (127.0.0.1)"
	}
	if ip.IsPrivate() {
		return "内网局域网 (Private LAN)"
	}

	// Geolocation mapping mock / heuristics for enterprise audit
	if strings.HasPrefix(ipStr, "114.114") || strings.HasPrefix(ipStr, "223.5") {
		return "中国 (DNS公共节点)"
	}
	return "亚太地区"
}

// Produce pushes a log event to the asynchronous Kafka/channel queue
func (s *Service) Produce(event LogEvent) bool {
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	if event.GeoLocation == "" {
		event.GeoLocation = ResolveGeoIP(event.ClientIP)
	}

	select {
	case s.queue <- event:
		return true
	default:
		// Queue full (drop or alert in high-load scenarios)
		return false
	}
}

// StartConsumer runs background workers to consume audit logs
func (s *Service) StartConsumer(workers int, handler func(event LogEvent)) {
	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for {
				select {
				case <-s.ctx.Done():
					return
				case event, ok := <-s.queue:
					if !ok {
						return
					}
					handler(event)
				}
			}
		}()
	}
}

// Stop gracefully shuts down log consumer workers
func (s *Service) Stop() {
	s.cancel()
	close(s.queue)
	s.wg.Wait()
}
