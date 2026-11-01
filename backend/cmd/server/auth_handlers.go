package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"mim/models"
	"mim/pkg/jwtx"
	"mim/service/auth"
)

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
	// Security: Fail-closed. Disabled by default unless explicitly allowed.
	if os.Getenv("MIM_ALLOW_DEMO_USERS") != "true" && os.Getenv("MIM_ENV") != "development" {
		s.jsonError(w, http.StatusForbidden, "Demo users endpoint is disabled for security (requires MIM_ALLOW_DEMO_USERS=true)")
		return
	}

	var users []models.User
	// Restrict strictly to predefined demo seed accounts; never expose real registered users
	s.db.Select("id, account, nickname, avatar, role").
		Where("account IN (?)", []string{"alex", "bob", "carol", "david"}).
		Find(&users)

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
