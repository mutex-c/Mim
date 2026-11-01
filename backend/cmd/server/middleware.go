package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mim/models"
	"mim/pkg/jwtx"
)

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

		if tokenStr == "" {
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

// checkConversationAccess verifies whether userID is a legitimate participant of convID
func (s *Server) checkConversationAccess(userID uint64, convID string) bool {
	if convID == "" || userID == 0 {
		return false
	}
	if strings.HasPrefix(convID, "group_") {
		groupID, err := strconv.ParseUint(strings.TrimPrefix(convID, "group_"), 10, 64)
		if err != nil {
			return false
		}
		var count int64
		s.db.Model(&models.GroupMember{}).Where("group_id = ? AND user_id = ?", groupID, userID).Count(&count)
		return count > 0
	}
	// Single chat: minID_maxID
	parts := strings.Split(convID, "_")
	if len(parts) == 2 {
		u1, err1 := strconv.ParseUint(parts[0], 10, 64)
		u2, err2 := strconv.ParseUint(parts[1], 10, 64)
		if err1 == nil && err2 == nil {
			return userID == u1 || userID == u2
		}
	}
	return false
}
