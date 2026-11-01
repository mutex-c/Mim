package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"mim/pkg/database"
	"mim/pkg/discovery"
	"mim/pkg/kafkax"
	"mim/pkg/redisx"
	"mim/protocol"
	rpcfw "mim/rpc/client"
	rpclogic "mim/rpc/logic"
	"mim/service/auth"
	"mim/service/chat"
	"mim/service/group"
	"mim/service/transfer"
	"mim/service/user"
)

var (
	grpcPort     = flag.Int("port", 9091, "gRPC server listen port")
	dbDriver     = flag.String("db-driver", "sqlite", "database driver: sqlite or mysql")
	dbDSN        = flag.String("db", "mim.db", "database DSN or file path")
	redisAddr    = flag.String("redis-addr", "127.0.0.1:6379", "redis host:port")
	redisPass    = flag.String("redis-pass", "redis", "redis password")
	kafkaBrokers = flag.String("kafka-brokers", "127.0.0.1:9092", "comma-separated Kafka brokers (optional)")
	etcdAddr     = flag.String("etcd-addr", "127.0.0.1:2379", "etcd host:port (leave empty for standalone)")
	nodeID       = flag.String("node-id", "logic-1", "service instance node id")
	jwtSecret    = flag.String("jwt-secret", "mim-enterprise-jwt-secret-key-2026", "jwt secret")
)

func main() {
	flag.Parse()

	if envSecret := os.Getenv("MIM_JWT_SECRET"); envSecret != "" {
		*jwtSecret = envSecret
	}

	log.Printf("[Logic] Starting Mim Logic Cluster Service [%s] on port %d...", *nodeID, *grpcPort)

	// 1. Initialize Database (SQLite WAL or MySQL 8)
	db, err := database.InitDB(database.Config{
		Driver: *dbDriver,
		DSN:    *dbDSN,
	})
	if err != nil {
		log.Fatalf("[Logic] Failed to open DB: %v", err)
	}

	// 2. Initialize Redis Client
	rdb, err := redisx.NewClient(redisx.Config{
		Addr:     *redisAddr,
		Password: *redisPass,
	})
	if err != nil {
		log.Printf("[Logic Warning] Redis not available: %v. Running in standalone mode.", err)
		rdb = nil
	} else {
		log.Printf("[Logic] Connected to Redis 7 at %s", *redisAddr)
	}

	// 3. Setup Etcd Client & Gateway Discovery
	var etcdClient *clientv3.Client
	var gwDisc *discovery.Discovery
	if *etcdAddr != "" {
		cli, err := clientv3.New(clientv3.Config{
			Endpoints:   []string{*etcdAddr},
			DialTimeout: 2 * time.Second,
		})
		if err == nil {
			etcdClient = cli
			disc, err := discovery.NewDiscovery(etcdClient, "gateway")
			if err == nil {
				gwDisc = disc
			}
		} else {
			log.Printf("[Logic Warning] Etcd connect failed: %v (falling back to direct dial)", err)
		}
	}

	// 4. Setup Kafka Producer for Asynchronous Push Pipeline
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
	defer producer.Close()

	// 5. Initialize ClusterPusher & Domain Services
	rpcMgr := rpcfw.NewManager()
	defer rpcMgr.Close()

	clusterPusher := transfer.NewClusterPusher(producer, rpcMgr, gwDisc, "127.0.0.1:9092", rdb)
	chatSvc := chat.NewService(db, clusterPusher, rdb)
	authSvc := auth.NewService(db, auth.Config{
		JwtSecret:     *jwtSecret,
		AccessExpire:  2 * time.Hour,
		RefreshExpire: 14 * 24 * time.Hour,
	})
	_ = user.NewService(db)
	_ = group.NewService(db)

	// 6. Start gRPC Server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Fatalf("[Logic] Failed to listen on port %d: %v", *grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	logicServer := rpclogic.NewServer(chatSvc, authSvc, *jwtSecret, producer)
	protocol.RegisterLogicServiceServer(grpcServer, logicServer)

	// 7. Register with Etcd (if configured)
	var registrar *discovery.Registrar
	if etcdClient != nil {
		localIP := "127.0.0.1"
		regAddr := fmt.Sprintf("%s:%d", localIP, *grpcPort)
		reg, err := discovery.NewRegistrar(etcdClient, "logic", *nodeID, regAddr, 5)
		if err == nil {
			registrar = reg
		} else {
			log.Printf("[Logic Warning] Etcd register failed: %v", err)
		}
	}

	go func() {
		log.Printf("[Logic] gRPC service ready. Listening on 127.0.0.1:%d", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("[Logic] Server terminated: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("[Logic] Shutting down...")
	if registrar != nil {
		_ = registrar.Close()
	}
	grpcServer.GracefulStop()
	if rdb != nil {
		_ = rdb.Close()
	}
	log.Printf("[Logic] Exited successfully.")
}
