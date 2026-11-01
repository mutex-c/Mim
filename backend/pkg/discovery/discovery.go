package discovery

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const ServicePrefix = "/mim/services"

// Registrar handles service registration with lease and keepalive in Etcd
type Registrar struct {
	client      *clientv3.Client
	serviceName string
	instanceID  string
	addr        string
	ttl         int64
	leaseID     clientv3.LeaseID
	cancel      context.CancelFunc
	closed      atomic.Bool
	stopCh      chan struct{}
}

// NewRegistrar creates a registrar and registers the service instance
func NewRegistrar(client *clientv3.Client, serviceName, instanceID, addr string, ttlSeconds int64) (*Registrar, error) {
	if ttlSeconds <= 0 {
		ttlSeconds = 5
	}

	r := &Registrar{
		client:      client,
		serviceName: serviceName,
		instanceID:  instanceID,
		addr:        addr,
		ttl:         ttlSeconds,
		stopCh:      make(chan struct{}),
	}

	if client == nil {
		// Mock / standalone mode
		registerLocal(serviceName, instanceID, addr)
		return r, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	// Grant lease
	grantCtx, grantCancel := context.WithTimeout(ctx, 3*time.Second)
	leaseResp, err := client.Grant(grantCtx, ttlSeconds)
	grantCancel()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to grant lease: %w", err)
	}
	r.leaseID = leaseResp.ID

	// Put service key with lease
	key := fmt.Sprintf("%s/%s/%s", ServicePrefix, serviceName, instanceID)
	putCtx, putCancel := context.WithTimeout(ctx, 3*time.Second)
	_, err = client.Put(putCtx, key, addr, clientv3.WithLease(leaseResp.ID))
	putCancel()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to put service key: %w", err)
	}

	// Start lease keepalive loop
	keepAliveCh, err := client.KeepAlive(ctx, leaseResp.ID)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to keepalive lease: %w", err)
	}

	go func() {
		for {
			select {
			case _, ok := <-keepAliveCh:
				if !ok {
					log.Printf("[Etcd Registrar] Keepalive channel closed for %s/%s", serviceName, instanceID)
					return
				}
			case <-ctx.Done():
				return
			case <-r.stopCh:
				return
			}
		}
	}()

	log.Printf("[Etcd Registrar] Successfully registered service %s/%s at %s (LeaseID: %x)",
		serviceName, instanceID, addr, leaseResp.ID)
	return r, nil
}

// Close revokes the lease and unregisters the service instance
func (r *Registrar) Close() error {
	if r.closed.Swap(true) {
		return nil
	}

	if r.client == nil {
		unregisterLocal(r.serviceName, r.instanceID)
		return nil
	}

	close(r.stopCh)
	if r.cancel != nil {
		r.cancel()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	key := fmt.Sprintf("%s/%s/%s", ServicePrefix, r.serviceName, r.instanceID)
	_, _ = r.client.Delete(ctx, key)
	if r.leaseID > 0 {
		_, _ = r.client.Revoke(ctx, r.leaseID)
	}

	log.Printf("[Etcd Registrar] Unregistered service %s/%s", r.serviceName, r.instanceID)
	return nil
}

// Discovery watches service instances in Etcd and provides client load balancing
type Discovery struct {
	client      *clientv3.Client
	serviceName string
	endpoints   map[string]string // instanceID -> addr
	mu          sync.RWMutex
	cancel      context.CancelFunc
	index       uint32
}

// NewDiscovery creates a discovery instance for a given service name
func NewDiscovery(client *clientv3.Client, serviceName string) (*Discovery, error) {
	d := &Discovery{
		client:      client,
		serviceName: serviceName,
		endpoints:   make(map[string]string),
	}

	if client == nil {
		// Mock / standalone mode: pull from local registry
		d.endpoints = getLocalEndpoints(serviceName)
		return d, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel

	prefix := fmt.Sprintf("%s/%s/", ServicePrefix, serviceName)

	// 1. Initial fetch of active endpoints
	getCtx, getCancel := context.WithTimeout(ctx, 3*time.Second)
	resp, err := client.Get(getCtx, prefix, clientv3.WithPrefix())
	getCancel()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to fetch service instances: %w", err)
	}

	d.mu.Lock()
	for _, kv := range resp.Kvs {
		key := string(kv.Key)
		instanceID := key[len(prefix):]
		d.endpoints[instanceID] = string(kv.Value)
	}
	d.mu.Unlock()

	// 2. Watch for changes (PUT / DELETE)
	go func() {
		watchCh := client.Watch(ctx, prefix, clientv3.WithPrefix())
		for wresp := range watchCh {
			for _, ev := range wresp.Events {
				key := string(ev.Kv.Key)
				if len(key) <= len(prefix) {
					continue
				}
				instanceID := key[len(prefix):]

				d.mu.Lock()
				switch ev.Type {
				case clientv3.EventTypePut:
					d.endpoints[instanceID] = string(ev.Kv.Value)
					log.Printf("[Etcd Discovery] Service instance updated: %s/%s -> %s", serviceName, instanceID, string(ev.Kv.Value))
				case clientv3.EventTypeDelete:
					delete(d.endpoints, instanceID)
					log.Printf("[Etcd Discovery] Service instance removed: %s/%s", serviceName, instanceID)
				}
				d.mu.Unlock()
			}
		}
	}()

	return d, nil
}

// GetEndpoint returns an active endpoint address using round-robin load balancing
func (d *Discovery) GetEndpoint() (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.endpoints) == 0 {
		// If in fallback mode, check local registry again
		if d.client == nil {
			local := getLocalEndpoints(d.serviceName)
			if len(local) > 0 {
				d.endpoints = local
			}
		}
		if len(d.endpoints) == 0 {
			return "", fmt.Errorf("no available instance for service: %s", d.serviceName)
		}
	}

	list := make([]string, 0, len(d.endpoints))
	for _, addr := range d.endpoints {
		list = append(list, addr)
	}

	idx := atomic.AddUint32(&d.index, 1) % uint32(len(list))
	return list[idx], nil
}

// GetRandomEndpoint returns a random active endpoint address
func (d *Discovery) GetRandomEndpoint() (string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.endpoints) == 0 {
		return "", fmt.Errorf("no available instance for service: %s", d.serviceName)
	}

	list := make([]string, 0, len(d.endpoints))
	for _, addr := range d.endpoints {
		list = append(list, addr)
	}

	return list[rand.Intn(len(list))], nil
}

// GetAllEndpoints returns all active endpoint addresses
func (d *Discovery) GetAllEndpoints() []string {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.client == nil {
		local := getLocalEndpoints(d.serviceName)
		if len(local) > 0 {
			d.endpoints = local
		}
	}

	list := make([]string, 0, len(d.endpoints))
	for _, addr := range d.endpoints {
		list = append(list, addr)
	}
	return list
}

// Close stops the watcher
func (d *Discovery) Close() {
	if d.cancel != nil {
		d.cancel()
	}
}

// ----------------------------------------------------
// Local In-Memory Registry for Fallback and Unit Tests
// ----------------------------------------------------

var (
	localRegistry = make(map[string]map[string]string)
	localMu       sync.RWMutex
)

func registerLocal(serviceName, instanceID, addr string) {
	localMu.Lock()
	defer localMu.Unlock()
	instances, ok := localRegistry[serviceName]
	if !ok {
		instances = make(map[string]string)
		localRegistry[serviceName] = instances
	}
	instances[instanceID] = addr
}

func unregisterLocal(serviceName, instanceID string) {
	localMu.Lock()
	defer localMu.Unlock()
	if instances, ok := localRegistry[serviceName]; ok {
		delete(instances, instanceID)
	}
}

func getLocalEndpoints(serviceName string) map[string]string {
	localMu.RLock()
	defer localMu.RUnlock()
	instances, ok := localRegistry[serviceName]
	if !ok {
		return make(map[string]string)
	}
	res := make(map[string]string, len(instances))
	for k, v := range instances {
		res[k] = v
	}
	return res
}
