package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"mim/protocol"
)

// Manager manages pooled gRPC connections to service endpoints
type Manager struct {
	conns map[string]*grpc.ClientConn
	mu    sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		conns: make(map[string]*grpc.ClientConn),
	}
}

// GetConn returns an existing connection or dials a new one to target
func (m *Manager) GetConn(target string) (*grpc.ClientConn, error) {
	m.mu.RLock()
	if conn, ok := m.conns[target]; ok {
		m.mu.RUnlock()
		return conn, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.conns[target]; ok {
		return conn, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC target (%s): %w", target, err)
	}

	m.conns[target] = conn
	return conn, nil
}

// GetLogicClient dials or reuses connection to a LogicService endpoint
func (m *Manager) GetLogicClient(target string) (protocol.LogicServiceClient, error) {
	conn, err := m.GetConn(target)
	if err != nil {
		return nil, err
	}
	return protocol.NewLogicServiceClient(conn), nil
}

// GetGatewayClient dials or reuses connection to a GatewayService endpoint
func (m *Manager) GetGatewayClient(target string) (protocol.GatewayServiceClient, error) {
	conn, err := m.GetConn(target)
	if err != nil {
		return nil, err
	}
	return protocol.NewGatewayServiceClient(conn), nil
}

// Close closes all open connections
func (m *Manager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, conn := range m.conns {
		_ = conn.Close()
	}
	m.conns = make(map[string]*grpc.ClientConn)
}
