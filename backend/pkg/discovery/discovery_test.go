package discovery

import (
	"testing"
)

func TestLocalDiscoveryFallback(t *testing.T) {
	serviceName := "logic"

	// Register 2 mock nodes
	r1, err := NewRegistrar(nil, serviceName, "node-1", "127.0.0.1:9001", 5)
	if err != nil {
		t.Fatalf("NewRegistrar failed: %v", err)
	}
	defer r1.Close()

	r2, err := NewRegistrar(nil, serviceName, "node-2", "127.0.0.1:9002", 5)
	if err != nil {
		t.Fatalf("NewRegistrar 2 failed: %v", err)
	}
	defer r2.Close()

	disc, err := NewDiscovery(nil, serviceName)
	if err != nil {
		t.Fatalf("NewDiscovery failed: %v", err)
	}
	defer disc.Close()

	endpoints := disc.GetAllEndpoints()
	if len(endpoints) != 2 {
		t.Fatalf("expected 2 endpoints, got %d (%v)", len(endpoints), endpoints)
	}

	ep1, err := disc.GetEndpoint()
	if err != nil {
		t.Fatalf("GetEndpoint failed: %v", err)
	}
	ep2, err := disc.GetEndpoint()
	if err != nil {
		t.Fatalf("GetEndpoint 2 failed: %v", err)
	}

	if ep1 == "" || ep2 == "" {
		t.Fatalf("empty endpoint returned")
	}

	// Test Unregister node-1
	_ = r1.Close()

	disc2, _ := NewDiscovery(nil, serviceName)
	endpointsAfter := disc2.GetAllEndpoints()
	if len(endpointsAfter) != 1 || endpointsAfter[0] != "127.0.0.1:9002" {
		t.Fatalf("expected 1 endpoint (node-2) after unregister, got %v", endpointsAfter)
	}
}
