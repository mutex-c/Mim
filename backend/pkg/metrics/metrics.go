package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// OnlineConnections tracks current active WebSocket connections per platform
	OnlineConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "mim",
			Subsystem: "gateway",
			Name:      "online_connections",
			Help:      "Current number of active WebSocket connections.",
		},
		[]string{"platform"},
	)

	// MessagesReceived counts inbound messages sent by clients
	MessagesReceived = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mim",
			Subsystem: "chat",
			Name:      "messages_received_total",
			Help:      "Total count of messages received from clients.",
		},
		[]string{"conv_type", "msg_type"},
	)

	// MessagesPushed counts outbound messages pushed to client connections
	MessagesPushed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mim",
			Subsystem: "gateway",
			Name:      "messages_pushed_total",
			Help:      "Total count of messages pushed down to clients.",
		},
		[]string{"conv_type"},
	)

	// DeliveryAckTotal counts message delivery acknowledgements
	DeliveryAckTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mim",
			Subsystem: "chat",
			Name:      "delivery_ack_total",
			Help:      "Total delivery acknowledgements processed.",
		},
		[]string{"status"},
	)

	// ReadReceiptTotal counts read receipts processed
	ReadReceiptTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mim",
			Subsystem: "chat",
			Name:      "read_receipt_total",
			Help:      "Total read receipts processed.",
		},
		[]string{"conv_type"},
	)

	// KickedUsersTotal counts same-platform user kick events
	KickedUsersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mim",
			Subsystem: "auth",
			Name:      "kicked_users_total",
			Help:      "Total users kicked out due to same-platform login.",
		},
		[]string{"platform"},
	)

	// MessageLatency measures message processing latency in seconds
	MessageLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "mim",
			Subsystem: "chat",
			Name:      "message_processing_duration_seconds",
			Help:      "Duration of message processing and delivery in seconds.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(
		OnlineConnections,
		MessagesReceived,
		MessagesPushed,
		DeliveryAckTotal,
		ReadReceiptTotal,
		KickedUsersTotal,
		MessageLatency,
	)
}

// Handler returns HTTP handler exposing Prometheus /metrics endpoint
func Handler() http.Handler {
	return promhttp.Handler()
}
