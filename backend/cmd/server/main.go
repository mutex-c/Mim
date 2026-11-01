package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"mim/pkg/database"
	"mim/pkg/redisx"
	"mim/service/ai"
	"mim/service/auth"
	"mim/service/chat"
	"mim/service/file"
	"mim/service/group"
	"mim/service/settings"
	"mim/service/user"
)

var (
	port       = flag.Int("port", 8080, "server listen port")
	dbDriver   = flag.String("db-driver", "sqlite", "database driver: sqlite or mysql")
	dbDSN      = flag.String("db", "mim.db", "database DSN or file path")
	storageDir = flag.String("storage", "uploads", "file upload storage directory")
	redisAddr  = flag.String("redis-addr", "127.0.0.1:6379", "redis host:port")
	redisPass  = flag.String("redis-pass", "redis", "redis password")
	jwtSecret  = "mim-enterprise-jwt-secret-key-2026"
)

type Server struct {
	db           *gorm.DB
	hub          *chat.Hub
	rdb          *redisx.Client
	authService  *auth.Service
	userService  *user.Service
	chatService  *chat.Service
	fileService  *file.Service
	groupService *group.Service
	aiService    *ai.AIService
	dfaFilter    *settings.DFAFilter
	cfg          auth.Config
}

func main() {
	flag.Parse()

	if envSecret := os.Getenv("MIM_JWT_SECRET"); envSecret != "" {
		jwtSecret = envSecret
	}
	if envDriver := os.Getenv("MIM_DB_DRIVER"); envDriver != "" {
		*dbDriver = envDriver
	}
	if envDSN := os.Getenv("MIM_DB_DSN"); envDSN != "" {
		*dbDSN = envDSN
	}
	if envAddr := os.Getenv("MIM_REDIS_ADDR"); envAddr != "" {
		*redisAddr = envAddr
	}
	if envPass := os.Getenv("MIM_REDIS_PASSWORD"); envPass != "" {
		*redisPass = envPass
	}

	// Ensure DB directory exists if dbDSN is a file path
	if *dbDriver == "sqlite" {
		if dir := filepath.Dir(*dbDSN); dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}
	}

	// 1. Initialize Database (SQLite WAL or MySQL 8)
	db, err := database.InitDB(database.Config{
		Driver: *dbDriver,
		DSN:    *dbDSN,
	})
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 2. Seed Initial Demo Users, Groups and Conversations
	seedInitialData(db)

	// 3. Initialize Redis Client for Distributed Routing & Kickout (FR-AUTH-03)
	rdb, err := redisx.NewClient(redisx.Config{
		Addr:     *redisAddr,
		Password: *redisPass,
		DB:       0,
	})
	redisStatus := "Connected"
	if err != nil {
		log.Printf("[Warning] Redis connection failed (%s): %v. Falling back to local mode.", *redisAddr, err)
		redisStatus = "Offline (Local Fallback)"
		rdb = nil
	}

	hub := chat.NewHub(rdb)
	authCfg := auth.Config{
		JwtSecret:     jwtSecret,
		AccessExpire:  2 * time.Hour,
		RefreshExpire: 14 * 24 * time.Hour,
	}

	fileCfg := file.Config{
		StorageDir: *storageDir,
		SignSecret: "mim-file-sign-key-2026",
	}

	dfaWords := []string{"spam", "abuse", "malware", "phishing", "trojan", "exploit", "secret_leak"}
	dfa := settings.NewDFAFilter(dfaWords)

	srv := &Server{
		db:           db,
		hub:          hub,
		rdb:          rdb,
		authService:  auth.NewService(db, authCfg),
		userService:  user.NewService(db),
		chatService:  chat.NewService(db, hub, rdb),
		fileService:  file.NewService(db, fileCfg),
		groupService: group.NewService(db),
		aiService:    ai.NewAIService(db),
		dfaFilter:    dfa,
		cfg:          authCfg,
	}

	// 4. Setup Routes, Prometheus Metrics & Middleware
	handler := srv.setupRoutes()

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("==================================================")
	log.Printf("  MIM (M Instant Message) Cluster Backend Server  ")
	log.Printf("  HTTP API:        http://127.0.0.1:%d", *port)
	log.Printf("  WebSocket:       ws://127.0.0.1:%d/ws", *port)
	log.Printf("  Prometheus:      http://127.0.0.1:%d/metrics", *port)
	log.Printf("  Database (%s):   %s", strings.ToUpper(*dbDriver), *dbDSN)
	log.Printf("  Redis 7 Route:   %s (%s)", *redisAddr, redisStatus)
	log.Printf("  File Storage:    %s/", *storageDir)
	log.Printf("==================================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}
