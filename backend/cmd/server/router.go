package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"mim/pkg/metrics"
)

// setupRoutes registers all HTTP and WebSocket endpoints and applies CORS middleware.
func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Auth APIs
	mux.HandleFunc("/api/auth/register", s.handleRegister)
	mux.HandleFunc("/api/auth/login", s.handleLogin)
	mux.HandleFunc("/api/auth/demo-users", s.handleDemoUsers)

	// User & Friend APIs
	mux.HandleFunc("/api/user/profile", s.authMiddleware(s.handleProfile))
	mux.HandleFunc("/api/user/update-profile", s.authMiddleware(s.handleUpdateProfile))
	mux.HandleFunc("/api/user/config", s.authMiddleware(s.handleUpdateConfig))
	mux.HandleFunc("/api/user/change-password", s.authMiddleware(s.handleChangePassword))
	mux.HandleFunc("/api/user/friends", s.authMiddleware(s.handleFriends))
	mux.HandleFunc("/api/user/delete-friend", s.authMiddleware(s.handleDeleteFriend))
	mux.HandleFunc("/api/user/search", s.authMiddleware(s.handleSearchUser))
	mux.HandleFunc("/api/user/add-friend", s.authMiddleware(s.handleAddFriend))
	mux.HandleFunc("/api/user/friend-requests", s.authMiddleware(s.handleFriendRequests))
	mux.HandleFunc("/api/user/handle-friend-request", s.authMiddleware(s.handleProcessFriendRequest))

	// Group APIs
	mux.HandleFunc("/api/group/create", s.authMiddleware(s.handleCreateGroup))
	mux.HandleFunc("/api/group/info", s.authMiddleware(s.handleGroupInfo))
	mux.HandleFunc("/api/group/members", s.authMiddleware(s.handleGroupMembers))
	mux.HandleFunc("/api/group/update", s.authMiddleware(s.handleUpdateGroup))
	mux.HandleFunc("/api/group/invite", s.authMiddleware(s.handleInviteGroup))
	mux.HandleFunc("/api/group/mute-member", s.authMiddleware(s.handleMuteMember))
	mux.HandleFunc("/api/group/kick-member", s.authMiddleware(s.handleKickMember))
	mux.HandleFunc("/api/group/disband", s.authMiddleware(s.handleDisbandGroup))
	mux.HandleFunc("/api/group/set-role", s.authMiddleware(s.handleSetGroupRole))
	mux.HandleFunc("/api/group/transfer-owner", s.authMiddleware(s.handleTransferGroupOwner))
	mux.HandleFunc("/api/group/quit", s.authMiddleware(s.handleQuitGroup))

	// File Upload / Download APIs (SHA-256 Deduplication)
	mux.HandleFunc("/api/file/upload", s.authMiddleware(s.handleFileUpload))
	mux.HandleFunc("/api/file/download", s.handleFileDownload)

	// Calls History API
	mux.HandleFunc("/api/calls/history", s.authMiddleware(s.handleCallHistory))
	mux.HandleFunc("/api/calls/record", s.authMiddleware(s.handleRecordCall))

	// Chat APIs
	mux.HandleFunc("/api/chat/conversations", s.authMiddleware(s.handleConversations))
	mux.HandleFunc("/api/chat/messages", s.authMiddleware(s.handleMessages))
	mux.HandleFunc("/api/chat/revoke", s.authMiddleware(s.handleRevoke))
	mux.HandleFunc("/api/chat/pin-conversation", s.authMiddleware(s.handlePinConversation))
	mux.HandleFunc("/api/chat/mute-conversation", s.authMiddleware(s.handleMuteConversation))
	mux.HandleFunc("/api/chat/read-conversation", s.authMiddleware(s.handleReadConversation))
	mux.HandleFunc("/api/chat/delete-conversation", s.authMiddleware(s.handleDeleteConversation))
	mux.HandleFunc("/api/chat/clear-messages", s.authMiddleware(s.handleClearMessages))
	mux.HandleFunc("/api/chat/delete-message", s.authMiddleware(s.handleDeleteMessage))

	// AI APIs
	mux.HandleFunc("/api/ai/config", s.authMiddleware(s.handleAIConfig))
	mux.HandleFunc("/api/ai/test", s.authMiddleware(s.handleAITest))
	mux.HandleFunc("/api/ai/reply", s.authMiddleware(s.handleAIReply))
	mux.HandleFunc("/api/ai/summary", s.authMiddleware(s.handleAISummary))

	// WebSocket Gateway Endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Prometheus Metrics Endpoint
	mux.Handle("/metrics", metrics.Handler())

	// Web Frontend static hosting
	if _, err := os.Stat("dist"); err == nil {
		fs := http.FileServer(http.Dir("dist"))
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/ws" {
				http.NotFound(w, r)
				return
			}
			filePath := filepath.Join("dist", filepath.Clean(r.URL.Path))
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
			http.ServeFile(w, r, filepath.Join("dist", "index.html"))
		})
	}

	return s.corsMiddleware(mux)
}
