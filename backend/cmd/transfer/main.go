package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"

	"mim/pkg/discovery"
	"mim/pkg/kafkax"
	"mim/pkg/redisx"
	rpcfw "mim/rpc/client"
	"mim/service/transfer"
)

var (
	kafkaBrokers = flag.String("kafka-brokers", "127.0.0.1:9092", "comma-separated Kafka brokers (leave empty for memory mode)")
	groupID      = flag.String("group-id", "mim_transfer_group", "Kafka consumer group ID")
	redisAddr    = flag.String("redis-addr", "127.0.0.1:6379", "redis host:port")
	redisPass    = flag.String("redis-pass", "redis", "redis password")
	etcdAddr     = flag.String("etcd-addr", "127.0.0.1:2379", "etcd host:port (optional)")
	nodeID       = flag.String("node-id", "transfer-1", "transfer instance node id")
)

func main() {
	flag.Parse()

	log.Printf("[Transfer] Starting Mim Transfer Engine [%s]...", *nodeID)

	// 1. Connect Redis (to check active user routing)
	rdb, err := redisx.NewClient(redisx.Config{
		Addr:     *redisAddr,
		Password: *redisPass,
	})
	if err != nil {
		log.Printf("[Transfer Warning] Redis not available (%s): %v", *redisAddr, err)
		rdb = nil
	} else {
		log.Printf("[Transfer] Connected to Redis 7 at %s", *redisAddr)
	}

	// 2. Setup Etcd Gateway Discovery
	var gwDisc *discovery.Discovery
	if *etcdAddr != "" {
		etcdCli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{*etcdAddr},
			DialTimeout: 2 * time.Second,
		})
		if err == nil {
			disc, err := discovery.NewDiscovery(etcdCli, "gateway")
			if err == nil {
				gwDisc = disc
			}
		}
	}

	// 3. Setup Kafka Consumer & Producer
	var brokers []string
	if *kafkaBrokers != "" {
		for _, b := range strings.Split(*kafkaBrokers, ",") {
			b = strings.TrimSpace(b)
			if b != "" {
				brokers = append(brokers, b)
			}
		}
	}

	producer := kafkax.NewProducer(brokers)

	// Pre-create the push topic BEFORE joining the consumer group.
	// A group reader that subscribes to a non-existent topic is assigned zero partitions
	// and never recovers until restarted, which silently breaks the whole push pipeline
	// on a fresh deployment (the topic is normally created by the first publish, which
	// happens after this process has already joined).
	ensureCtx, ensureCancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := kafkax.EnsureTopic(ensureCtx, brokers, transfer.TopicPush, 3); err != nil {
		log.Printf("[Transfer Warning] Could not ensure topic %s: %v (consumer may stay idle until the topic exists)", transfer.TopicPush, err)
	} else if len(brokers) > 0 {
		log.Printf("[Transfer] Topic %s ready", transfer.TopicPush)
	}
	ensureCancel()

	consumer := kafkax.NewConsumer(brokers, *groupID, transfer.TopicPush)
	rpcMgr := rpcfw.NewManager()

	transferSvc := transfer.NewTransferService(producer, consumer, rdb, rpcMgr, gwDisc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Start consuming push events
	transferSvc.Start(ctx)
	log.Printf("[Transfer] Kafka push pipeline active. Subscribed to topic: %s", transfer.TopicPush)

	// 5. Watchdog: a group member with zero assigned partitions consumes nothing while
	// still looking perfectly healthy from the outside. Fail loudly instead of idling.
	if len(brokers) > 0 {
		go watchConsumerAssignment(consumer, transfer.TopicPush)
	}

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[Transfer] Shutting down...")
	cancel()
	_ = consumer.Close()
	_ = producer.Close()
	rpcMgr.Close()
	if rdb != nil {
		_ = rdb.Close()
	}
	log.Printf("[Transfer] Exited cleanly.")
}

// watchConsumerAssignment verifies that the consumer group actually received partitions.
// Joining a group before the topic exists yields a permanently idle member: no error,
// no rebalance, no delivery. Silent idling is the worst possible failure mode for a push
// pipeline, so we escalate to a hard exit that is impossible to miss in `docker ps`.
func watchConsumerAssignment(consumer *kafkax.Consumer, topic string) {
	const (
		checkInterval = 30 * time.Second
		maxChecks     = 3
	)

	for i := 1; i <= maxChecks; i++ {
		time.Sleep(checkInterval)
		if consumer.IsFetching() {
			log.Printf("[Transfer] Consumer active: fetching from %s", topic)
			return
		}
		log.Printf("[Transfer Warning] Consumer has not fetched anything from %s after %s (check %d/%d): topic may be missing, or the consumer joined before it was created",
			topic, time.Duration(i)*checkInterval, i, maxChecks)
	}

	log.Printf("[Transfer FATAL] Consumer still has not fetched anything from %s after %s. The push pipeline is dead: every message will be persisted but never delivered. Restart mim-transfer once the topic exists.",
		topic, time.Duration(maxChecks)*checkInterval)
	os.Exit(1)
}
