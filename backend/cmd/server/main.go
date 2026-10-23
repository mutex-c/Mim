package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"mim/models"
	"mim/pkg/encrypt"
	"mim/pkg/jwtx"
	"mim/protocol"
	"mim/service/auth"
	"mim/service/chat"
	"mim/service/file"
	"mim/service/group"
	"mim/service/settings"
	"mim/service/user"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	port      = flag.Int("port", 8080, "server listen port")
	dbPath    = flag.String("db", "mim.db", "sqlite database file path")
	jwtSecret = "mim-enterprise-jwt-secret-key-2026"
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow cross-origin for local dev / desktop clients
		},
		ReadBufferSize:  1024 * 64,
		WriteBufferSize: 1024 * 64,
	}
)

type Server struct {
	db           *gorm.DB
	hub          *chat.Hub
	authService  *auth.Service
	userService  *user.Service
	chatService  *chat.Service
	fileService  *file.Service
	groupService *group.Service
	dfaFilter    *settings.DFAFilter
	cfg          auth.Config
	callLogs     []CallRecord
	callMu       sync.RWMutex
}

type CallRecord struct {
	ID        string `json:"id"`
	FromUser  string `json:"from_user"`
	ToUser    string `json:"to_user"`
	CallType  string `json:"call_type"` // audio or video
	Status    string `json:"status"`    // completed, missed, rejected
	Duration  int    `json:"duration"`  // seconds
	CreatedAt string `json:"created_at"`
}

func main() {
	flag.Parse()

	// Ensure DB directory exists if dbPath has parent
	if dir := filepath.Dir(*dbPath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	// 1. Initialize SQLite Database
	db, err := gorm.Open(sqlite.Open(*dbPath+"?_journal=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to initialize SQLite: %v", err)
	}

	// 2. Run Auto Migrations
	if err := models.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}

	// 3. Seed Initial Demo Users, Groups and Conversations
	seedInitialData(db)

	hub := chat.NewHub()
	authCfg := auth.Config{
		JwtSecret:     jwtSecret,
		AccessExpire:  2 * time.Hour,
		RefreshExpire: 14 * 24 * time.Hour,
	}

	fileCfg := file.Config{
		StorageDir: "uploads",
		SignSecret: "mim-file-sign-key-2026",
	}

	dfaWords := []string{"spam", "abuse", "malware", "phishing", "trojan", "exploit", "secret_leak"}
	dfa := settings.NewDFAFilter(dfaWords)

	srv := &Server{
		db:           db,
		hub:          hub,
		authService:  auth.NewService(db, authCfg),
		userService:  user.NewService(db),
		chatService:  chat.NewService(db, hub),
		fileService:  file.NewService(db, fileCfg),
		groupService: group.NewService(db),
		dfaFilter:    dfa,
		cfg:          authCfg,
		callLogs:     make([]CallRecord, 0),
	}

	// 4. Setup Routes
	mux := http.NewServeMux()

	// Auth APIs
	mux.HandleFunc("/api/auth/register", srv.handleRegister)
	mux.HandleFunc("/api/auth/login", srv.handleLogin)
	mux.HandleFunc("/api/auth/demo-users", srv.handleDemoUsers)

	// User & Friend APIs
	mux.HandleFunc("/api/user/profile", srv.authMiddleware(srv.handleProfile))
	mux.HandleFunc("/api/user/friends", srv.authMiddleware(srv.handleFriends))
	mux.HandleFunc("/api/user/search", srv.authMiddleware(srv.handleSearchUser))
	mux.HandleFunc("/api/user/add-friend", srv.authMiddleware(srv.handleAddFriend))

	// Group APIs
	mux.HandleFunc("/api/group/create", srv.authMiddleware(srv.handleCreateGroup))
	mux.HandleFunc("/api/group/members", srv.authMiddleware(srv.handleGroupMembers))

	// File Upload / Download APIs (SHA-256 Deduplication / 秒传)
	mux.HandleFunc("/api/file/upload", srv.authMiddleware(srv.handleFileUpload))
	mux.HandleFunc("/api/file/download", srv.handleFileDownload)

	// Calls History API
	mux.HandleFunc("/api/calls/history", srv.authMiddleware(srv.handleCallHistory))
	mux.HandleFunc("/api/calls/record", srv.authMiddleware(srv.handleRecordCall))

	// Chat APIs
	mux.HandleFunc("/api/chat/conversations", srv.authMiddleware(srv.handleConversations))
	mux.HandleFunc("/api/chat/messages", srv.authMiddleware(srv.handleMessages))
	mux.HandleFunc("/api/chat/revoke", srv.authMiddleware(srv.handleRevoke))

	// WebSocket Gateway Endpoint
	mux.HandleFunc("/ws", srv.handleWebSocket)

	// Wrap with CORS
	handler := srv.corsMiddleware(mux)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("==================================================")
	log.Printf("  MIM (M Instant Message) Cluster Backend Server  ")
	log.Printf("  HTTP API:        http://127.0.0.1:%d", *port)
	log.Printf("  WebSocket:       ws://127.0.0.1:%d/ws", *port)
	log.Printf("  SQLite Database: %s", *dbPath)
	log.Printf("  File Storage:    uploads/", )
	log.Printf("==================================================")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server exited with error: %v", err)
	}
}

// Seed initial test data
func seedInitialData(db *gorm.DB) {
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return
	}

	hash, _ := encrypt.HashPassword("123456")

	// Users: Alex Chen (1), Sarah Jenkins (2), Michael Zhang (3), David Miller (4)
	users := []models.User{
		{
			ID:           1,
			Account:      "alex",
			Nickname:     "Alex Chen",
			PasswordHash: hash,
			Avatar:       "AC",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleAdmin,
		},
		{
			ID:           2,
			Account:      "sarah",
			Nickname:     "Sarah Jenkins",
			PasswordHash: hash,
			Avatar:       "SJ",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleNormal,
		},
		{
			ID:           3,
			Account:      "michael",
			Nickname:     "Michael Zhang",
			PasswordHash: hash,
			Avatar:       "MZ",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleNormal,
		},
		{
			ID:           4,
			Account:      "david",
			Nickname:     "David Miller",
			PasswordHash: hash,
			Avatar:       "DM",
			Status:       models.UserStatusNormal,
			Role:         models.UserRoleNormal,
		},
	}

	for _, u := range users {
		db.Create(&u)
		db.Create(&models.UserConfig{
			UserID:           u.ID,
			SearchByAccount:  1,
			SearchByNickname: 1,
			FriendVerifyMode: models.VerifyModeDirect,
			RecallTimeLimit:  120,
		})
	}

	// Establish Friends
	friends := []models.Friend{
		{UserID: 1, FriendID: 2},
		{UserID: 2, FriendID: 1},
		{UserID: 1, FriendID: 3},
		{UserID: 3, FriendID: 1},
		{UserID: 1, FriendID: 4},
		{UserID: 4, FriendID: 1},
		{UserID: 2, FriendID: 3},
		{UserID: 3, FriendID: 2},
	}
	db.Create(&friends)

	// Seed Group 101: Core Infrastructure Team
	grp := models.Group{
		ID:           101,
		Title:        "Core Infrastructure Team",
		Avatar:       "CI",
		Notice:       "v1.4.2 canary deployed. All cluster nodes nominal.",
		OwnerID:      1,
		MemberCount:  4,
		MaxMembers:   500,
		IsSearchable: 1,
	}
	db.Create(&grp)

	grpMembers := []models.GroupMember{
		{GroupID: 101, UserID: 1, MemberNickname: "Alex Chen", Role: models.GroupRoleOwner},
		{GroupID: 101, UserID: 2, MemberNickname: "Sarah Jenkins", Role: models.GroupRoleAdmin},
		{GroupID: 101, UserID: 3, MemberNickname: "Michael Zhang", Role: models.GroupRoleMember},
		{GroupID: 101, UserID: 4, MemberNickname: "David Miller", Role: models.GroupRoleMember},
	}
	db.Create(&grpMembers)

	// Establish Initial Conversations
	now := time.Now()
	conv12 := models.Conversation{
		ConversationID: "1_2",
		UserID:         1,
		TargetID:       2,
		ConvType:       models.ConvTypeSingle,
		LastMsgID:      2,
		LastMsgSeq:     2,
		LastMsgPreview: "The Q3 architecture review is ready for sign-off.",
		LastMsgTime:    now.Add(-5 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	conv21 := models.Conversation{
		ConversationID: "1_2",
		UserID:         2,
		TargetID:       1,
		ConvType:       models.ConvTypeSingle,
		LastMsgID:      2,
		LastMsgSeq:     2,
		LastMsgPreview: "The Q3 architecture review is ready for sign-off.",
		LastMsgTime:    now.Add(-5 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	conv13 := models.Conversation{
		ConversationID: "1_3",
		UserID:         1,
		TargetID:       3,
		ConvType:       models.ConvTypeSingle,
		LastMsgID:      3,
		LastMsgSeq:     1,
		LastMsgPreview: "All integration test suites passing.",
		LastMsgTime:    now.Add(-1 * time.Hour),
		UnreadCount:    0,
		IsPinned:       0,
	}
	convGrp1 := models.Conversation{
		ConversationID: "group_101",
		UserID:         1,
		TargetID:       101,
		ConvType:       models.ConvTypeGroup,
		LastMsgID:      4,
		LastMsgSeq:     1,
		LastMsgPreview: "David: Deployed v1.4.2 to canary cluster.",
		LastMsgTime:    now.Add(-20 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	convGrp2 := models.Conversation{
		ConversationID: "group_101",
		UserID:         2,
		TargetID:       101,
		ConvType:       models.ConvTypeGroup,
		LastMsgID:      4,
		LastMsgSeq:     1,
		LastMsgPreview: "David: Deployed v1.4.2 to canary cluster.",
		LastMsgTime:    now.Add(-20 * time.Minute),
		UnreadCount:    0,
		IsPinned:       1,
	}
	db.Create(&conv12)
	db.Create(&conv21)
	db.Create(&conv13)
	db.Create(&convGrp1)
	db.Create(&convGrp2)

	// Seed Messages
	msgs := []models.Message{
		{
			ID:             1,
			ClientMsgID:    "init_c1",
			ConversationID: "1_2",
			ConvType:       models.ConvTypeSingle,
			SenderID:       2,
			ReceiverID:     1,
			SeqID:          1,
			MsgType:        models.MsgTypeText,
			Content:        "Hi Alex, did you review the distributed SeqId monotonic generator proposal?",
			CreatedAt:      now.Add(-10 * time.Minute),
		},
		{
			ID:             2,
			ClientMsgID:    "init_c2",
			ConversationID: "1_2",
			ConvType:       models.ConvTypeSingle,
			SenderID:       2,
			ReceiverID:     1,
			SeqID:          2,
			MsgType:        models.MsgTypeText,
			Content:        "The Q3 architecture review is ready for sign-off.",
			CreatedAt:      now.Add(-5 * time.Minute),
		},
		{
			ID:             3,
			ClientMsgID:    "init_c3",
			ConversationID: "1_3",
			ConvType:       models.ConvTypeSingle,
			SenderID:       3,
			ReceiverID:     1,
			SeqID:          1,
			MsgType:        models.MsgTypeText,
			Content:        "All integration test suites passing.",
			CreatedAt:      now.Add(-1 * time.Hour),
		},
		{
			ID:             4,
			ClientMsgID:    "init_c4",
			ConversationID: "group_101",
			ConvType:       models.ConvTypeGroup,
			SenderID:       4,
			ReceiverID:     0,
			SeqID:          1,
			MsgType:        models.MsgTypeText,
			Content:        "Deployed v1.4.2 to canary cluster. Observability metrics are nominal.",
			CreatedAt:      now.Add(-20 * time.Minute),
		},
	}
	db.Create(&msgs)
	log.Printf("Seeded 4 demo users, 1 group, and initial conversations.")
}

// CORS Middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Auth Middleware
func (s *Server) authMiddleware(next func(w http.ResponseWriter, r *http.Request, userID uint64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		} else if qToken := r.URL.Query().Get("token"); qToken != "" {
			tokenStr = qToken
		}

		// Also allow quick query ?user_id=1 for ease of testing
		if tokenStr == "" {
			if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
				uid, _ := strconv.ParseUint(uidStr, 10, 64)
				if uid > 0 {
					next(w, r, uid)
					return
				}
			}
			s.jsonError(w, http.StatusUnauthorized, "Missing authentication token")
			return
		}

		claims, err := jwtx.ParseToken(s.cfg.JwtSecret, tokenStr)
		if err != nil {
			s.jsonError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		next(w, r, claims.UserID)
	}
}

func (s *Server) jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": data,
	})
}

func (s *Server) jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code": status,
		"msg":  msg,
	})
}

// Handlers

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req auth.RegisterReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := s.authService.Register(&req)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	tokens, err := jwtx.GenerateTokenPair(s.cfg.JwtSecret, s.cfg.AccessExpire, s.cfg.RefreshExpire, user.ID, user.Account, "web", "browser-main")
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to generate tokens")
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"user":  user,
		"token": tokens,
	})
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req auth.LoginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Platform == "" {
		req.Platform = "web"
	}
	if req.DeviceID == "" {
		req.DeviceID = "browser-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	resp, err := s.authService.Login(&req)
	if err != nil {
		s.jsonError(w, http.StatusUnauthorized, err.Error())
		return
	}

	s.jsonResponse(w, resp)
}

func (s *Server) handleDemoUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	s.db.Select("id, account, nickname, avatar, role").Find(&users)

	type DemoUserItem struct {
		User  models.User     `json:"user"`
		Token *jwtx.TokenPair `json:"token"`
	}

	list := make([]DemoUserItem, 0, len(users))
	for _, u := range users {
		tokens, _ := jwtx.GenerateTokenPair(s.cfg.JwtSecret, s.cfg.AccessExpire, s.cfg.RefreshExpire, u.ID, u.Account, "web", "demo-device")
		list = append(list, DemoUserItem{
			User:  u,
			Token: tokens,
		})
	}

	s.jsonResponse(w, list)
}

func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request, userID uint64) {
	profile, err := s.userService.GetProfile(userID)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "User not found")
		return
	}
	s.jsonResponse(w, profile)
}

func (s *Server) handleFriends(w http.ResponseWriter, r *http.Request, userID uint64) {
	friends, err := s.userService.GetFriendList(userID)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type FriendWithOnline struct {
		user.FriendItem
		IsOnline bool `json:"is_online"`
	}

	res := make([]FriendWithOnline, 0, len(friends))
	for _, f := range friends {
		res = append(res, FriendWithOnline{
			FriendItem: *f,
			IsOnline:   s.hub.IsOnline(f.FriendID),
		})
	}

	s.jsonResponse(w, res)
}

func (s *Server) handleSearchUser(w http.ResponseWriter, r *http.Request, userID uint64) {
	kw := r.URL.Query().Get("keyword")
	results, err := s.userService.SearchUser(kw, userID)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.jsonResponse(w, results)
}

func (s *Server) handleAddFriend(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		TargetID uint64 `json:"target_id"`
		Msg      string `json:"msg"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := s.userService.SendFriendRequest(userID, body.TargetID, body.Msg)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "request_sent"})
}

func (s *Server) handleCreateGroup(w http.ResponseWriter, r *http.Request, userID uint64) {
	var req group.CreateGroupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	req.OwnerID = userID

	grp, err := s.groupService.CreateGroup(&req)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, grp)
}

func (s *Server) handleGroupMembers(w http.ResponseWriter, r *http.Request, userID uint64) {
	grpIDStr := r.URL.Query().Get("group_id")
	grpID, _ := strconv.ParseUint(grpIDStr, 10, 64)
	if grpID == 0 {
		s.jsonError(w, http.StatusBadRequest, "group_id is required")
		return
	}

	var members []models.GroupMember
	if err := s.db.Where("group_id = ?", grpID).Find(&members).Error; err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, members)
}

// File Upload Handler (SHA-256 Instant Upload / 秒传)
func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 32MB max in memory
	_ = r.ParseMultipartForm(32 << 20)
	fileHeader, _, err := r.FormFile("file")
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, "Missing file parameter")
		return
	}
	defer fileHeader.Close()

	data, err := io.ReadAll(fileHeader)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to read file bytes")
		return
	}

	fileName := r.FormValue("name")
	if fileName == "" {
		fileName = "upload_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	record, err := s.fileService.SaveFile(userID, fileName, http.DetectContentType(data), data)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"file_id":   record.ID,
		"file_name": record.FileName,
		"file_size": record.FileSize,
		"file_hash": record.FileHash,
		"file_url":  fmt.Sprintf("/api/file/download?file_id=%d", record.ID),
	})
}

// File Download Handler
func (s *Server) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("file_id")
	fileID, _ := strconv.ParseUint(fileIDStr, 10, 64)

	var record models.FileRecord
	if err := s.db.First(&record, fileID).Error; err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", record.FileName))
	http.ServeFile(w, r, record.StoragePath)
}

// Call history & recording
func (s *Server) handleCallHistory(w http.ResponseWriter, r *http.Request, userID uint64) {
	s.callMu.RLock()
	defer s.callMu.RUnlock()
	s.jsonResponse(w, s.callLogs)
}

func (s *Server) handleRecordCall(w http.ResponseWriter, r *http.Request, userID uint64) {
	var rec CallRecord
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	rec.ID = "call_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	rec.CreatedAt = time.Now().Format("15:04:05")

	s.callMu.Lock()
	s.callLogs = append([]CallRecord{rec}, s.callLogs...)
	if len(s.callLogs) > 50 {
		s.callLogs = s.callLogs[:50]
	}
	s.callMu.Unlock()

	s.jsonResponse(w, rec)
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request, userID uint64) {
	var convs []models.Conversation
	s.db.Where("user_id = ?", userID).Order("is_pinned DESC, last_msg_time DESC").Find(&convs)

	type ConvItemResp struct {
		models.Conversation
		Title      string `json:"title"`
		AvatarText string `json:"avatar_text"`
		IsOnline   bool   `json:"is_online"`
	}

	list := make([]ConvItemResp, 0, len(convs))
	for _, c := range convs {
		title := ""
		avatar := ""
		isOnline := false

		if c.ConvType == models.ConvTypeSingle {
			var target models.User
			if err := s.db.First(&target, c.TargetID).Error; err == nil {
				title = target.Nickname
				avatar = target.Avatar
				if avatar == "" && len(target.Nickname) > 0 {
					avatar = strings.ToUpper(target.Nickname[:2])
				}
				isOnline = s.hub.IsOnline(c.TargetID)
			}
		} else {
			var group models.Group
			if err := s.db.First(&group, c.TargetID).Error; err == nil {
				title = group.Title
				avatar = "CI"
			}
		}

		list = append(list, ConvItemResp{
			Conversation: c,
			Title:        title,
			AvatarText:   avatar,
			IsOnline:     isOnline,
		})
	}

	s.jsonResponse(w, list)
}

func (s *Server) handleMessages(w http.ResponseWriter, r *http.Request, userID uint64) {
	convID := r.URL.Query().Get("conv_id")
	if convID == "" {
		s.jsonError(w, http.StatusBadRequest, "conv_id is required")
		return
	}

	lastSeqStr := r.URL.Query().Get("last_seq_id")
	lastSeq, _ := strconv.ParseUint(lastSeqStr, 10, 64)

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 50
	}

	var msgs []models.Message
	s.db.Where("conversation_id = ? AND seq_id > ?", convID, lastSeq).
		Order("seq_id ASC").
		Limit(limit).
		Find(&msgs)

	s.jsonResponse(w, msgs)
}

func (s *Server) handleRevoke(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		MsgID uint64 `json:"msg_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.chatService.RevokeMessage(userID, body.MsgID); err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.jsonResponse(w, map[string]string{"status": "revoked"})
}

// WebSocket Gateway Connection Handler
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	tokenStr := r.URL.Query().Get("token")
	var userID uint64
	platform := r.URL.Query().Get("platform")
	if platform == "" {
		platform = "web"
	}
	deviceID := r.URL.Query().Get("device_id")
	if deviceID == "" {
		deviceID = fmt.Sprintf("dev_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
	}

	if tokenStr != "" {
		claims, err := jwtx.ParseToken(s.cfg.JwtSecret, tokenStr)
		if err == nil {
			userID = claims.UserID
			if claims.Platform != "" {
				platform = claims.Platform
			}
			if claims.DeviceID != "" {
				deviceID = claims.DeviceID
			}
		}
	}

	// Allow query param user_id for quick sandbox tests
	if userID == 0 {
		if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
			userID, _ = strconv.ParseUint(uidStr, 10, 64)
		}
	}

	if userID == 0 {
		http.Error(w, "Unauthorized WebSocket handshake", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	client := chat.NewClient(userID, platform, deviceID, conn, s.hub)
	s.hub.Register(client)
	log.Printf("[WS] User %d connected on %s (Device: %s). Online users: %d", userID, platform, deviceID, s.hub.OnlineCount())

	// Start pump routines
	var wg sync.WaitGroup
	wg.Add(2)

	// Write pump
	go func() {
		defer wg.Done()
		for data := range client.SendChan {
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				break
			}
		}
	}()

	// Read pump
	go func() {
		defer func() {
			s.hub.Unregister(client)
			client.Close()
			log.Printf("[WS] User %d disconnected. Online users: %d", userID, s.hub.OnlineCount())
			wg.Done()
		}()

		conn.SetReadLimit(10 * 1024 * 1024)
		_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
			return nil
		})

		for {
			msgType, rawData, err := conn.ReadMessage()
			if err != nil {
				break
			}

			_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))

			// Process incoming frame
			if msgType == websocket.BinaryMessage {
				s.processBinaryFrame(client, rawData)
			} else if msgType == websocket.TextMessage {
				s.processTextFrame(client, rawData)
			}
		}
	}()

	wg.Wait()
}

// processBinaryFrame parses the 24-byte protocol frame and dispatches actions
func (s *Server) processBinaryFrame(client *chat.Client, data []byte) {
	frame, err := protocol.Decode(data)
	if err != nil {
		log.Printf("[WS] Invalid binary frame from user %d: %v", client.UserID, err)
		return
	}

	switch frame.CmdType {
	case protocol.CmdHeartbeatPing:
		pong := protocol.NewFrame(protocol.CmdHeartbeatPong, frame.SeqID, 0, nil)
		if bytes, err := pong.Encode(); err == nil {
			client.SendChan <- bytes
		}

	case protocol.CmdChatMessageSend:
		s.handleFrameChatMessage(client, frame)

	case protocol.CmdSyncReq:
		s.handleFrameSync(client, frame)

	case protocol.CmdWebRTCSignal:
		s.handleFrameWebRTC(client, frame)

	default:
		log.Printf("[WS] Unhandled CmdType: 0x%04X", frame.CmdType)
	}
}

// handleFrameChatMessage processes a ChatMessageSend packet
func (s *Server) handleFrameChatMessage(client *chat.Client, frame *protocol.Frame) {
	var chatMsg protocol.ChatMessage
	err := proto.Unmarshal(frame.Payload, &chatMsg)
	if err != nil {
		// Fallback to JSON payload
		err = json.Unmarshal(frame.Payload, &chatMsg)
	}
	if err != nil {
		log.Printf("[WS] Failed to decode chat message payload: %v", err)
		return
	}

	// 1. DFA Sensitive Word Filter
	filteredContent := s.dfaFilter.Filter(chatMsg.Content, '*')
	chatMsg.Content = filteredContent

	// 2. Verified Sender ID
	chatMsg.SenderId = client.UserID

	sendReq := &chat.SendMessageReq{
		ClientMsgID:    chatMsg.ClientMsgId,
		SenderID:       client.UserID,
		ReceiverID:     chatMsg.ReceiverId,
		GroupID:        0,
		ConvType:       int8(chatMsg.ConvType),
		MsgType:        int8(chatMsg.MsgType),
		Content:        filteredContent,
		SenderDeviceID: client.DeviceID,
	}
	if chatMsg.Quote != nil {
		sendReq.QuotedMsgID = chatMsg.Quote.QuotedMsgId
		sendReq.QuotedContent = chatMsg.Quote.PreviewContent
	}

	ack, err := s.chatService.SendMessage(sendReq)
	if err != nil {
		log.Printf("[WS] SendMessage error: %v", err)
		ack = &protocol.MessageAck{
			ClientMsgId: chatMsg.ClientMsgId,
			Code:        500,
			ErrorMsg:    err.Error(),
			ServerTime:  time.Now().UnixMilli(),
		}
	}

	// Reply ACK to sender
	ackPayload, _ := json.Marshal(ack)
	ackFrame := protocol.NewFrame(protocol.CmdMessageAck, ack.SeqId, protocol.FlagJSON, ackPayload)
	if ackBytes, err := ackFrame.Encode(); err == nil {
		client.SendChan <- ackBytes
	}
}

// handleFrameSync processes incremental sync request
func (s *Server) handleFrameSync(client *chat.Client, frame *protocol.Frame) {
	var syncReq protocol.SyncRequest
	if err := proto.Unmarshal(frame.Payload, &syncReq); err != nil {
		_ = json.Unmarshal(frame.Payload, &syncReq)
	}

	resp, err := s.chatService.SyncMessages(syncReq.ConversationId, syncReq.LastSeqId, int(syncReq.Limit))
	if err != nil {
		return
	}

	respPayload, _ := json.Marshal(resp)
	respFrame := protocol.NewFrame(protocol.CmdSyncResp, resp.LatestSeqId, protocol.FlagJSON, respPayload)
	if respBytes, err := respFrame.Encode(); err == nil {
		client.SendChan <- respBytes
	}
}

// handleFrameWebRTC processes ephemeral WebRTC signaling (strictly memory-only)
func (s *Server) handleFrameWebRTC(client *chat.Client, frame *protocol.Frame) {
	var signal protocol.WebRTCSignal
	if err := proto.Unmarshal(frame.Payload, &signal); err != nil {
		_ = json.Unmarshal(frame.Payload, &signal)
	}

	signal.FromUserId = client.UserID

	// Relay signal to target peer
	payload, _ := json.Marshal(signal)
	f := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, payload)
	if encoded, err := f.Encode(); err == nil {
		s.hub.SendToUser(signal.ToUserId, encoded)
	}

	// If HANGUP, record call history
	if signal.SignalType == "HANGUP" {
		s.callMu.Lock()
		s.callLogs = append([]CallRecord{
			{
				ID:        signal.CallId,
				FromUser:  fmt.Sprintf("User %d", signal.FromUserId),
				ToUser:    fmt.Sprintf("User %d", signal.ToUserId),
				CallType:  "Audio/Video",
				Status:    "Completed",
				Duration:  42,
				CreatedAt: time.Now().Format("15:04:05"),
			},
		}, s.callLogs...)
		s.callMu.Unlock()
	}
}

// processTextFrame supports developer debugging
func (s *Server) processTextFrame(client *chat.Client, data []byte) {
	var jsonReq struct {
		Action  string          `json:"action"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &jsonReq); err != nil {
		return
	}

	if jsonReq.Action == "ping" {
		_ = client.Conn.WriteJSON(map[string]interface{}{
			"action": "pong",
			"time":   time.Now().UnixMilli(),
		})
	}
}
