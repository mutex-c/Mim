package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusMetricsRecording(t *testing.T) {
	// 1. Record metrics
	OnlineConnections.WithLabelValues("web").Inc()
	OnlineConnections.WithLabelValues("android").Set(5)

	MessagesReceived.WithLabelValues("single", "text").Inc()
	MessagesPushed.WithLabelValues("single").Inc()
	DeliveryAckTotal.WithLabelValues("delivered").Inc()
	ReadReceiptTotal.WithLabelValues("single").Inc()
	KickedUsersTotal.WithLabelValues("web").Inc()

	timer := MessageLatency.WithLabelValues("send_message")
	timer.Observe(0.015)

	// 2. Query /metrics HTTP handler
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := Handler()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got: %d", rec.Code)
	}

	body := rec.Body.String()
	expectedMetrics := []string{
		"mim_gateway_online_connections",
		"mim_chat_messages_received_total",
		"mim_gateway_messages_pushed_total",
		"mim_chat_delivery_ack_total",
		"mim_chat_read_receipt_total",
		"mim_auth_kicked_users_total",
		"mim_chat_message_processing_duration_seconds",
	}

	for _, name := range expectedMetrics {
		if !strings.Contains(body, name) {
			t.Errorf("expected metric %s not found in /metrics output", name)
		}
	}
}
