package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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
	"mim/service/ai"
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
	port       = flag.Int("port", 8080, "server listen port")
	dbPath     = flag.String("db", "mim.db", "sqlite database file path")
	storageDir = flag.String("storage", "uploads", "file upload storage directory")
	jwtSecret  = "mim-enterprise-jwt-secret-key-2026"
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
	aiService    *ai.AIService
	dfaFilter    *settings.DFAFilter
	cfg          auth.Config
}

func main() {
	flag.Parse()

	// Ensure DB directory exists if dbPath has parent
	if dir := filepath.Dir(*dbPath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	// 1. Initialize SQLite Database (WAL mode, NORMAL sync & single writer for zero lock contention)
	db, err := gorm.Open(sqlite.Open(*dbPath+"?_journal=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=-64000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to initialize SQLite: %v", err)
	}

	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
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
		StorageDir: *storageDir,
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
		aiService:    ai.NewAIService(db),
		dfaFilter:    dfa,
		cfg:          authCfg,
	}

	// 4. Setup Routes
	mux := http.NewServeMux()

	// Auth APIs
	mux.HandleFunc("/api/auth/register", srv.handleRegister)
	mux.HandleFunc("/api/auth/login", srv.handleLogin)
	mux.HandleFunc("/api/auth/demo-users", srv.handleDemoUsers)

	// User & Friend APIs
	mux.HandleFunc("/api/user/profile", srv.authMiddleware(srv.handleProfile))
	mux.HandleFunc("/api/user/update-profile", srv.authMiddleware(srv.handleUpdateProfile))
	mux.HandleFunc("/api/user/config", srv.authMiddleware(srv.handleUpdateConfig))
	mux.HandleFunc("/api/user/change-password", srv.authMiddleware(srv.handleChangePassword))
	mux.HandleFunc("/api/user/friends", srv.authMiddleware(srv.handleFriends))
	mux.HandleFunc("/api/user/delete-friend", srv.authMiddleware(srv.handleDeleteFriend))
	mux.HandleFunc("/api/user/search", srv.authMiddleware(srv.handleSearchUser))
	mux.HandleFunc("/api/user/add-friend", srv.authMiddleware(srv.handleAddFriend))
	mux.HandleFunc("/api/user/friend-requests", srv.authMiddleware(srv.handleFriendRequests))
	mux.HandleFunc("/api/user/handle-friend-request", srv.authMiddleware(srv.handleProcessFriendRequest))

	// Group APIs
	mux.HandleFunc("/api/group/create", srv.authMiddleware(srv.handleCreateGroup))
	mux.HandleFunc("/api/group/info", srv.authMiddleware(srv.handleGroupInfo))
	mux.HandleFunc("/api/group/members", srv.authMiddleware(srv.handleGroupMembers))
	mux.HandleFunc("/api/group/update", srv.authMiddleware(srv.handleUpdateGroup))
	mux.HandleFunc("/api/group/invite", srv.authMiddleware(srv.handleInviteGroup))
	mux.HandleFunc("/api/group/mute-member", srv.authMiddleware(srv.handleMuteMember))
	mux.HandleFunc("/api/group/kick-member", srv.authMiddleware(srv.handleKickMember))
	mux.HandleFunc("/api/group/disband", srv.authMiddleware(srv.handleDisbandGroup))
	mux.HandleFunc("/api/group/set-role", srv.authMiddleware(srv.handleSetGroupRole))
	mux.HandleFunc("/api/group/transfer-owner", srv.authMiddleware(srv.handleTransferGroupOwner))
	mux.HandleFunc("/api/group/quit", srv.authMiddleware(srv.handleQuitGroup))

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
	mux.HandleFunc("/api/chat/pin-conversation", srv.authMiddleware(srv.handlePinConversation))
	mux.HandleFunc("/api/chat/mute-conversation", srv.authMiddleware(srv.handleMuteConversation))
	mux.HandleFunc("/api/chat/read-conversation", srv.authMiddleware(srv.handleReadConversation))
	mux.HandleFunc("/api/chat/delete-conversation", srv.authMiddleware(srv.handleDeleteConversation))
	mux.HandleFunc("/api/chat/clear-messages", srv.authMiddleware(srv.handleClearMessages))
	mux.HandleFunc("/api/chat/delete-message", srv.authMiddleware(srv.handleDeleteMessage))

	// AI APIs
	mux.HandleFunc("/api/ai/config", srv.authMiddleware(srv.handleAIConfig))
	mux.HandleFunc("/api/ai/test", srv.authMiddleware(srv.handleAITest))
	mux.HandleFunc("/api/ai/reply", srv.authMiddleware(srv.handleAIReply))
	mux.HandleFunc("/api/ai/summary", srv.authMiddleware(srv.handleAISummary))

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

	// Seed Initial Call Records (Specifically for Alex Chen and Sarah Jenkins)
	calls := []models.CallRecord{
		{
			CallID:     "call_init_1",
			CallerID:   1,
			CallerName: "Alex Chen",
			CalleeID:   2,
			CalleeName: "Sarah Jenkins",
			CallType:   "audio",
			Status:     "Completed",
			Duration:   184,
			CreatedAt:  now.Add(-2 * time.Hour),
		},
		{
			CallID:     "call_init_2",
			CallerID:   1,
			CallerName: "Alex Chen",
			CalleeID:   3,
			CalleeName: "Michael Zhang",
			CallType:   "video",
			Status:     "Completed",
			Duration:   340,
			CreatedAt:  now.Add(-24 * time.Hour),
		},
	}
	db.Create(&calls)

	log.Printf("Seeded 4 demo users, 1 group, initial conversations and calls.")
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

func (s *Server) handleUpdateProfile(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		Nickname  string `json:"nickname"`
		Avatar    string `json:"avatar"`
		Signature string `json:"signature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := s.userService.UpdateProfile(userID, body.Nickname, body.Avatar, body.Signature); err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "profile_updated"})
}

func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request, userID uint64) {
	var cfg models.UserConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := s.userService.UpdateConfig(userID, &cfg); err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "config_updated"})
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	var u models.User
	if err := s.db.First(&u, userID).Error; err != nil {
		s.jsonError(w, http.StatusNotFound, "User not found")
		return
	}
	if !encrypt.CheckPasswordHash(body.OldPassword, u.PasswordHash) {
		s.jsonError(w, http.StatusBadRequest, "原密码不正确")
		return
	}
	newHash, err := encrypt.HashPassword(body.NewPassword)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	s.db.Model(&u).Update("password_hash", newHash)
	s.jsonResponse(w, map[string]string{"status": "password_changed"})
}

func (s *Server) handleDeleteFriend(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		FriendID uint64 `json:"friend_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	_ = s.userService.DeleteFriend(userID, body.FriendID)
	_ = s.userService.DeleteFriend(body.FriendID, userID)
	s.jsonResponse(w, map[string]string{"status": "friend_deleted"})
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

	isDirectFriend, err := s.userService.SendFriendRequest(userID, body.TargetID, body.Msg)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	if isDirectFriend {
		s.jsonResponse(w, map[string]interface{}{"status": "friend_added", "is_friend": true, "msg": "已成功添加为好友"})
	} else {
		s.jsonResponse(w, map[string]interface{}{"status": "request_sent", "is_friend": false, "msg": "好友申请已发送，等待对方通过"})
	}
}

func (s *Server) handleFriendRequests(w http.ResponseWriter, r *http.Request, userID uint64) {
	requests, err := s.userService.GetFriendRequests(userID)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.jsonResponse(w, requests)
}

func (s *Server) handleProcessFriendRequest(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		RequestID uint64 `json:"request_id"`
		Action    int8   `json:"action"` // 2 = Agreed, 3 = Rejected
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := s.userService.HandleFriendRequest(body.RequestID, userID, body.Action)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]interface{}{"status": "success", "action": body.Action})
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

func (s *Server) handleGroupInfo(w http.ResponseWriter, r *http.Request, userID uint64) {
	grpIDStr := r.URL.Query().Get("group_id")
	grpID, _ := strconv.ParseUint(grpIDStr, 10, 64)
	if grpID == 0 {
		s.jsonError(w, http.StatusBadRequest, "group_id is required")
		return
	}
	grp, err := s.groupService.GetGroupDetails(grpID)
	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Group not found")
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

	var grp models.Group
	s.db.First(&grp, grpID)

	type MemberInfo struct {
		UserID     uint64     `json:"user_id"`
		Account    string     `json:"account"`
		Nickname   string     `json:"nickname"`
		Avatar     string     `json:"avatar"`
		Role       int8       `json:"role"`
		MutedUntil *time.Time `json:"muted_until"`
		IsMuted    bool       `json:"is_muted"`
	}

	res := make([]MemberInfo, 0, len(members))
	now := time.Now()
	for _, m := range members {
		var u models.User
		s.db.Select("id, account, nickname, avatar").First(&u, m.UserID)
		isMuted := false
		if m.Role != models.GroupRoleOwner && m.Role != models.GroupRoleAdmin {
			if grp.IsAllMuted == 1 {
				isMuted = true
			} else if m.MutedUntil != nil && now.Before(*m.MutedUntil) {
				isMuted = true
			}
		}
		res = append(res, MemberInfo{
			UserID:     m.UserID,
			Account:    u.Account,
			Nickname:   u.Nickname,
			Avatar:     u.Avatar,
			Role:       m.Role,
			MutedUntil: m.MutedUntil,
			IsMuted:    isMuted,
		})
	}
	s.jsonResponse(w, res)
}

func (s *Server) handleUpdateGroup(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID      uint64 `json:"group_id"`
		Title        string `json:"title"`
		Notice       string `json:"notice"`
		Avatar       string `json:"avatar"`
		IsSearchable int8   `json:"is_searchable"`
		JoinMode     int8   `json:"join_mode"`
		IsAllMuted   int8   `json:"is_all_muted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.UpdateGroupSettings(userID, body.GroupID, body.Title, body.Notice, body.Avatar, body.IsSearchable, body.JoinMode, body.IsAllMuted)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "updated"})
}

func (s *Server) handleInviteGroup(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID   uint64   `json:"group_id"`
		MemberIDs []uint64 `json:"member_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.InviteMembers(userID, body.GroupID, body.MemberIDs)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "invited"})
}

func (s *Server) handleMuteMember(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID  uint64 `json:"group_id"`
		TargetID uint64 `json:"target_id"`
		Duration int    `json:"duration"` // duration in seconds, 0 to unmute
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.SetMemberMute(userID, body.GroupID, body.TargetID, time.Duration(body.Duration)*time.Second)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "success"})
}

func (s *Server) handleKickMember(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID  uint64 `json:"group_id"`
		TargetID uint64 `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.KickMember(userID, body.GroupID, body.TargetID)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "kicked"})
}

func (s *Server) handleDisbandGroup(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID uint64 `json:"group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.DisbandGroup(userID, body.GroupID)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "disbanded"})
}

func (s *Server) handleSetGroupRole(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID  uint64 `json:"group_id"`
		TargetID uint64 `json:"target_id"`
		Role     int8   `json:"role"` // 2: Admin, 3: Member
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.SetMemberRole(userID, body.GroupID, body.TargetID, body.Role)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]interface{}{"status": "success", "role": body.Role})
}

func (s *Server) handleTransferGroupOwner(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID    uint64 `json:"group_id"`
		NewOwnerID uint64 `json:"new_owner_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.TransferOwner(userID, body.GroupID, body.NewOwnerID)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "transferred"})
}

func (s *Server) handleQuitGroup(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		GroupID uint64 `json:"group_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request")
		return
	}
	err := s.groupService.QuitGroup(userID, body.GroupID)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "quit"})
}

// File Upload Handler (SHA-256 Instant Upload / 秒传)
func (s *Server) handleFileUpload(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 32MB max in memory
	_ = r.ParseMultipartForm(32 << 20)
	fileHeader, handler, err := r.FormFile("file")
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
	if fileName == "" && handler != nil && handler.Filename != "" {
		fileName = handler.Filename
	}
	if fileName == "" {
		fileName = "upload_" + strconv.FormatInt(time.Now().Unix(), 10)
	}

	record, err := s.fileService.SaveFile(userID, fileName, http.DetectContentType(data), data)
	if err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}

	encodedName := url.QueryEscape(record.FileName)
	fileUrl := fmt.Sprintf("/api/file/download?file_id=%d&name=%s", record.ID, encodedName)
	s.jsonResponse(w, map[string]interface{}{
		"file_id":   record.ID,
		"file_name": record.FileName,
		"file_size": record.FileSize,
		"file_hash": record.FileHash,
		"file_url":  fileUrl,
		"url":       fileUrl,
		"name":      record.FileName,
		"size":      record.FileSize,
		"mime":      record.FileType,
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

	dlName := r.URL.Query().Get("name")
	if dlName == "" {
		dlName = record.FileName
	}
	encodedName := url.PathEscape(dlName)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"; filename*=UTF-8''%s", dlName, encodedName))
	if record.FileType != "" {
		w.Header().Set("Content-Type", record.FileType)
	}
	http.ServeFile(w, r, record.StoragePath)
}

// Call history & recording (Strict per-user isolation: user can only see calls they participated in)
func (s *Server) handleCallHistory(w http.ResponseWriter, r *http.Request, userID uint64) {
	var records []models.CallRecord
	s.db.Where("caller_id = ? OR callee_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(50).
		Find(&records)

	type CallHistoryItem struct {
		ID        string `json:"id"`
		FromUser  string `json:"from_user"`
		ToUser    string `json:"to_user"`
		CallType  string `json:"call_type"`
		Status    string `json:"status"`
		Duration  int    `json:"duration"`
		CreatedAt string `json:"created_at"`
		IsCaller  bool   `json:"is_caller"`
	}

	res := make([]CallHistoryItem, 0, len(records))
	for _, rec := range records {
		res = append(res, CallHistoryItem{
			ID:        rec.CallID,
			FromUser:  rec.CallerName,
			ToUser:    rec.CalleeName,
			CallType:  rec.CallType,
			Status:    rec.Status,
			Duration:  rec.Duration,
			CreatedAt: rec.CreatedAt.Format("15:04"),
			IsCaller:  rec.CallerID == userID,
		})
	}

	s.jsonResponse(w, res)
}

func (s *Server) handleRecordCall(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		CallID     string `json:"call_id"`
		CallerID   uint64 `json:"caller_id"`
		CallerName string `json:"caller_name"`
		CalleeID   uint64 `json:"callee_id"`
		CalleeName string `json:"callee_name"`
		CallType   string `json:"call_type"`
		Status     string `json:"status"`
		Duration   int    `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	callerID := body.CallerID
	if callerID == 0 {
		callerID = userID
	}
	callerName := body.CallerName
	if callerName == "" {
		var u models.User
		if err := s.db.First(&u, callerID).Error; err == nil {
			callerName = u.Nickname
		}
	}
	calleeName := body.CalleeName
	if calleeName == "" && body.CalleeID > 0 {
		var u models.User
		if err := s.db.First(&u, body.CalleeID).Error; err == nil {
			calleeName = u.Nickname
		}
	}

	callID := body.CallID
	if callID == "" {
		callID = "call_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	rec := models.CallRecord{
		CallID:     callID,
		CallerID:   callerID,
		CallerName: callerName,
		CalleeID:   body.CalleeID,
		CalleeName: calleeName,
		CallType:   body.CallType,
		Status:     body.Status,
		Duration:   body.Duration,
		CreatedAt:  time.Now(),
	}
	s.db.Create(&rec)

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

	// Strict permission check: user MUST be a participant of the conversation!
	if strings.HasPrefix(convID, "group_") {
		groupID, _ := strconv.ParseUint(strings.TrimPrefix(convID, "group_"), 10, 64)
		var member models.GroupMember
		if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
			s.jsonError(w, http.StatusForbidden, "Access denied: you are not a member of this group")
			return
		}
	} else {
		// Single chat: format is "minID_maxID"
		parts := strings.Split(convID, "_")
		if len(parts) == 2 {
			u1, _ := strconv.ParseUint(parts[0], 10, 64)
			u2, _ := strconv.ParseUint(parts[1], 10, 64)
			if userID != u1 && userID != u2 {
				s.jsonError(w, http.StatusForbidden, "Access denied: you are not a participant in this conversation")
				return
			}
		}
	}

	lastSeqStr := r.URL.Query().Get("last_seq_id")
	lastSeq, _ := strconv.ParseUint(lastSeqStr, 10, 64)

	beforeSeqStr := r.URL.Query().Get("before_seq_id")
	beforeSeq, _ := strconv.ParseUint(beforeSeqStr, 10, 64)

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var msgs []models.Message
	if beforeSeq > 0 {
		// History paging: pull messages before beforeSeq (older messages)
		var revMsgs []models.Message
		s.db.Where("conversation_id = ? AND seq_id < ?", convID, beforeSeq).
			Order("seq_id DESC").
			Limit(limit).
			Find(&revMsgs)
		msgs = make([]models.Message, len(revMsgs))
		for i, m := range revMsgs {
			msgs[len(revMsgs)-1-i] = m
		}
	} else if lastSeq > 0 {
		// Incremental sync: pull messages after lastSeq (newer messages)
		s.db.Where("conversation_id = ? AND seq_id > ?", convID, lastSeq).
			Order("seq_id ASC").
			Limit(limit).
			Find(&msgs)
	} else {
		// Initial pull: fetch latest limit messages in chronological order
		var revMsgs []models.Message
		s.db.Where("conversation_id = ?", convID).
			Order("seq_id DESC").
			Limit(limit).
			Find(&revMsgs)
		msgs = make([]models.Message, len(revMsgs))
		for i, m := range revMsgs {
			msgs[len(revMsgs)-1-i] = m
		}
	}

	s.jsonResponse(w, msgs)
}

func (s *Server) handleRevoke(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		MsgID       uint64 `json:"msg_id"`
		ClientMsgID string `json:"client_msg_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.chatService.RevokeMessage(userID, body.MsgID, body.ClientMsgID); err != nil {
		s.jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	s.jsonResponse(w, map[string]string{"status": "revoked"})
}

func (s *Server) handlePinConversation(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ConversationID string `json:"conversation_id"`
		IsPinned       int8   `json:"is_pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := s.db.Model(&models.Conversation{}).
		Where("user_id = ? AND conversation_id = ?", userID, body.ConversationID).
		Update("is_pinned", body.IsPinned).Error; err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.jsonResponse(w, map[string]interface{}{"status": "success", "is_pinned": body.IsPinned})
}

func (s *Server) handleMuteConversation(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ConversationID string `json:"conversation_id"`
		IsMuted        int8   `json:"is_muted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := s.db.Model(&models.Conversation{}).
		Where("user_id = ? AND conversation_id = ?", userID, body.ConversationID).
		Update("is_muted", body.IsMuted).Error; err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.jsonResponse(w, map[string]interface{}{"status": "success", "is_muted": body.IsMuted})
}

func (s *Server) handleReadConversation(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ConversationID string `json:"conversation_id"`
		UnreadCount    int32  `json:"unread_count"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := s.db.Model(&models.Conversation{}).
		Where("user_id = ? AND conversation_id = ?", userID, body.ConversationID).
		Update("unread_count", body.UnreadCount).Error; err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.jsonResponse(w, map[string]interface{}{"status": "success", "unread_count": body.UnreadCount})
}

func (s *Server) handleDeleteConversation(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ConversationID string `json:"conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := s.db.Where("user_id = ? AND conversation_id = ?", userID, body.ConversationID).
		Delete(&models.Conversation{}).Error; err != nil {
		s.jsonError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.jsonResponse(w, map[string]string{"status": "deleted"})
}

func (s *Server) handleClearMessages(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ConversationID string `json:"conversation_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	s.db.Where("conversation_id = ?", body.ConversationID).Delete(&models.Message{})
	s.db.Model(&models.Conversation{}).
		Where("conversation_id = ?", body.ConversationID).
		Updates(map[string]interface{}{
			"last_msg_preview": "",
			"unread_count":     0,
		})
	s.jsonResponse(w, map[string]string{"status": "cleared"})
}

func (s *Server) handleDeleteMessage(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		ConversationID string `json:"conversation_id"`
		MsgID          uint64 `json:"msg_id"`
		ClientMsgID    string `json:"client_msg_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if body.MsgID > 0 {
		s.db.Where("id = ? AND conversation_id = ?", body.MsgID, body.ConversationID).Delete(&models.Message{})
	} else if body.ClientMsgID != "" {
		s.db.Where("client_msg_id = ? AND conversation_id = ?", body.ClientMsgID, body.ConversationID).Delete(&models.Message{})
	}
	s.jsonResponse(w, map[string]string{"status": "deleted"})
}

// AI Config API Handler (GET to read, POST to save)
func (s *Server) handleAIConfig(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method == http.MethodGet {
		cfg, err := s.aiService.GetAIConfig(userID)
		if err != nil {
			s.jsonError(w, http.StatusInternalServerError, "Failed to get AI config: "+err.Error())
			return
		}
		s.jsonResponse(w, cfg)
		return
	}
	if r.Method == http.MethodPost {
		var cfg ai.AIConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			s.jsonError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if err := s.aiService.SaveAIConfig(userID, &cfg); err != nil {
			s.jsonError(w, http.StatusInternalServerError, "Failed to save AI config: "+err.Error())
			return
		}
		s.jsonResponse(w, map[string]string{"status": "config_saved"})
		return
	}
	s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

// AI Test Connection API Handler (POST to test upstream connectivity without saving)
func (s *Server) handleAITest(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	var body struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
		Model   string `json:"model"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if strings.TrimSpace(body.BaseURL) == "" {
		s.jsonError(w, http.StatusBadRequest, "base_url is required")
		return
	}

	latencyMs, err := s.aiService.TestConnection(r.Context(), body.BaseURL, body.APIKey, body.Model)
	if err != nil {
		s.jsonError(w, http.StatusBadRequest, "Connection test failed: "+err.Error())
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"latency_ms": latencyMs,
	})
}

// AI Smart Reply API Handler
func (s *Server) handleAIReply(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method != http.MethodPost {
		s.jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var body struct {
		ConversationID string `json:"conversation_id"`
		Count          int    `json:"count"`
		Tone           string `json:"tone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.ConversationID == "" {
		s.jsonError(w, http.StatusBadRequest, "conversation_id required")
		return
	}

	var conv models.Conversation
	if err := s.db.Where("user_id = ? AND conversation_id = ?", userID, body.ConversationID).First(&conv).Error; err != nil {
		s.jsonError(w, http.StatusForbidden, "Access denied: you are not a participant of this conversation")
		return
	}

	// Pull last 8 non-revoked messages
	var rawMsgs []models.Message
	s.db.Where("conversation_id = ? AND is_revoked = 0", body.ConversationID).
		Order("id DESC").
		Limit(8).
		Find(&rawMsgs)

	if len(rawMsgs) == 0 {
		s.jsonResponse(w, map[string]interface{}{
			"suggestions": ai.DefaultReplySuggestions,
		})
		return
	}

	msgs := make([]models.Message, len(rawMsgs))
	for i, m := range rawMsgs {
		msgs[len(rawMsgs)-1-i] = m
	}

	cfg, err := s.aiService.GetAIConfig(userID)
	if err != nil || cfg == nil || cfg.APIKey == "" {
		s.jsonResponse(w, map[string]interface{}{
			"suggestions": ai.DefaultReplySuggestions,
		})
		return
	}

	userMap := make(map[uint64]string)
	var contexts []ai.ChatMessageContext
	for _, m := range msgs {
		name, ok := userMap[m.SenderID]
		if !ok {
			var u models.User
			if err := s.db.First(&u, m.SenderID).Error; err == nil && u.Nickname != "" {
				name = u.Nickname
			} else {
				name = fmt.Sprintf("User_%d", m.SenderID)
			}
			userMap[m.SenderID] = name
		}

		content := strings.TrimSpace(m.Content)
		if m.MsgType == models.MsgTypeImage {
			content = "[图片文件]"
		} else if m.MsgType == models.MsgTypeVoice {
			content = "[语音留言]"
		} else if m.MsgType == models.MsgTypeFile {
			content = "[文档附件]"
		}
		contexts = append(contexts, ai.ChatMessageContext{
			SenderName: name,
			Content:    content,
			IsMe:       m.SenderID == userID,
		})
	}

	suggestions, err := s.aiService.CallSmartReply(r.Context(), *cfg, contexts)
	if err != nil || len(suggestions) == 0 {
		log.Printf("[AIReply] upstream AI call failed: %v, falling back", err)
		s.jsonResponse(w, map[string]interface{}{
			"suggestions": ai.DefaultReplySuggestions,
		})
		return
	}

	s.jsonResponse(w, map[string]interface{}{
		"suggestions": suggestions,
	})
}

// AI Summary API Handler (Upgraded with real LLM when API key configured, falling back to rule summary)
func (s *Server) handleAISummary(w http.ResponseWriter, r *http.Request, userID uint64) {
	var body struct {
		ConversationID string `json:"conversation_id"`
		Lang           string `json:"lang"`
		Limit          int    `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.ConversationID == "" {
		s.jsonError(w, http.StatusBadRequest, "conversation_id required")
		return
	}

	var conv models.Conversation
	if err := s.db.Where("user_id = ? AND conversation_id = ?", userID, body.ConversationID).First(&conv).Error; err != nil {
		s.jsonError(w, http.StatusForbidden, "Access denied: you are not a participant of this conversation")
		return
	}

	limit := body.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var rawMsgs []models.Message
	s.db.Where("conversation_id = ? AND is_revoked = 0", body.ConversationID).
		Order("id DESC").
		Limit(limit).
		Find(&rawMsgs)

	if len(rawMsgs) == 0 {
		emptyMsg := "当前会话暂无足够的消息记录可供智能提炼。请在会话中发送几条消息后再试。"
		if body.Lang == "en" {
			emptyMsg = "No sufficient conversation history available for AI synthesis. Please exchange messages before summarizing."
		}
		s.jsonResponse(w, map[string]interface{}{
			"summary": emptyMsg,
			"count":   0,
		})
		return
	}

	msgs := make([]models.Message, len(rawMsgs))
	for i, m := range rawMsgs {
		msgs[len(rawMsgs)-1-i] = m
	}

	userMap := make(map[uint64]string)
	var lines []string
	var contexts []ai.ChatMessageContext
	for _, m := range msgs {
		name, ok := userMap[m.SenderID]
		if !ok {
			var u models.User
			if err := s.db.First(&u, m.SenderID).Error; err == nil && u.Nickname != "" {
				name = u.Nickname
			} else {
				name = fmt.Sprintf("User_%d", m.SenderID)
			}
			userMap[m.SenderID] = name
		}

		content := strings.TrimSpace(m.Content)
		if m.MsgType == models.MsgTypeImage {
			content = "[图片文件]"
		} else if m.MsgType == models.MsgTypeVoice {
			content = "[语音留言]"
		} else if m.MsgType == models.MsgTypeFile {
			content = "[文档附件]"
		}
		contexts = append(contexts, ai.ChatMessageContext{
			SenderName: name,
			Content:    content,
			IsMe:       m.SenderID == userID,
		})

		if len(content) > 120 {
			content = content[:120] + "..."
		}
		lines = append(lines, fmt.Sprintf("%s: %s", name, content))
	}

	// Try upstream LLM summary if APIKey configured
	aiCfg, err := s.aiService.GetAIConfig(userID)
	if err == nil && aiCfg != nil && aiCfg.APIKey != "" {
		aiSummary, err := s.aiService.CallSmartSummary(r.Context(), *aiCfg, contexts, body.Lang)
		if err == nil && aiSummary != "" {
			s.jsonResponse(w, map[string]interface{}{
				"summary":       aiSummary,
				"message_count": len(msgs),
				"participants":  len(userMap),
			})
			return
		}
		log.Printf("[AISummary] upstream call failed: %v, falling back to rule summary", err)
	}

	// Fallback to local rule summary
	names := make([]string, 0, len(userMap))
	for _, n := range userMap {
		names = append(names, n)
	}

	var summary string
	if body.Lang == "en" {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📊 **Mim Intelligence Briefing** (Analyzed %d messages)\n", len(msgs)))
		sb.WriteString(fmt.Sprintf("• **Key Contributors**: %s\n", strings.Join(names, ", ")))
		sb.WriteString("• **Discussion Timeline & Core Highlights**:\n")
		take := 5
		if len(lines) < take {
			take = len(lines)
		}
		for i := 0; i < take; i++ {
			sb.WriteString(fmt.Sprintf("  - %s\n", lines[len(lines)-take+i]))
		}
		sb.WriteString("• **Synthesis & Takeaways**: Active discussion across shared threads; all data records committed to Mim monotonic cluster.")
		summary = sb.String()
	} else {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("📊 **Mim 智能协同提炼总结**（已深度分析近 %d 条真实发言）\n", len(msgs)))
		sb.WriteString(fmt.Sprintf("• **发言参与人员**：%s\n", strings.Join(names, "、")))
		sb.WriteString("• **核心交流脉络提炼**：\n")
		take := 5
		if len(lines) < take {
			take = len(lines)
		}
		for i := 0; i < take; i++ {
			sb.WriteString(fmt.Sprintf("  - %s\n", lines[len(lines)-take+i]))
		}
		sb.WriteString("• **执行结论**：上下文已达成阶段性共识，关键要点与文件数据已由 Mim 统一落盘闭环。")
		summary = sb.String()
	}

	s.jsonResponse(w, map[string]interface{}{
		"summary":       summary,
		"message_count": len(msgs),
		"participants":  len(userMap),
	})
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

	if tokenStr != "" {
		claims, err := jwtx.ParseToken(s.cfg.JwtSecret, tokenStr)
		if err == nil {
			userID = claims.UserID
			if claims.Platform != "" && platform == "" {
				platform = claims.Platform
			}
			if deviceID == "" && claims.DeviceID != "" {
				deviceID = claims.DeviceID
			}
		}
	}

	if deviceID == "" {
		deviceID = fmt.Sprintf("dev_%d_%d", time.Now().Unix(), time.Now().Nanosecond())
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
				log.Printf("[WS] User %d ReadMessage error: %v", userID, err)
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

	groupID := uint64(0)
	receiverID := chatMsg.ReceiverId
	if chatMsg.ConvType == 1 { // Group
		groupID = chatMsg.ReceiverId
		receiverID = 0
	}

	sendReq := &chat.SendMessageReq{
		ClientMsgID:    chatMsg.ClientMsgId,
		SenderID:       client.UserID,
		ReceiverID:     receiverID,
		GroupID:        groupID,
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
	payload, _ := json.Marshal(&signal)
	f := protocol.NewFrame(protocol.CmdWebRTCSignal, 0, protocol.FlagJSON, payload)
	if encoded, err := f.Encode(); err == nil {
		s.hub.SendToUser(signal.ToUserId, encoded)
	}

	// If HANGUP or REJECT, record call history into database for caller & callee
	if signal.SignalType == "HANGUP" || signal.SignalType == "REJECT" {
		var fromUser, toUser models.User
		_ = s.db.First(&fromUser, signal.FromUserId)
		_ = s.db.First(&toUser, signal.ToUserId)

		fromName := fromUser.Nickname
		if fromName == "" {
			fromName = fmt.Sprintf("User %d", signal.FromUserId)
		}
		toName := toUser.Nickname
		if toName == "" {
			toName = fmt.Sprintf("User %d", signal.ToUserId)
		}

		callID := signal.CallId
		if callID == "" {
			callID = fmt.Sprintf("call_%d_%d", time.Now().Unix(), signal.FromUserId)
		}

		status := "Completed"
		if signal.SignalType == "REJECT" {
			status = "Rejected"
		}

		callType := "video"
		duration := 0
		if signal.Data != "" {
			var d struct {
				CallType string `json:"callType"`
				Duration int    `json:"duration"`
			}
			if err := json.Unmarshal([]byte(signal.Data), &d); err == nil {
				if d.CallType != "" {
					callType = d.CallType
				}
				if d.Duration > 0 {
					duration = d.Duration
				}
			}
		}

		rec := models.CallRecord{
			CallID:     callID,
			CallerID:   signal.FromUserId,
			CallerName: fromName,
			CalleeID:   signal.ToUserId,
			CalleeName: toName,
			CallType:   callType,
			Status:     status,
			Duration:   duration,
			CreatedAt:  time.Now(),
		}
		s.db.Create(&rec)
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
