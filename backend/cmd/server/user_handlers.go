package main

import (
	"encoding/json"
	"net/http"

	"mim/models"
	"mim/pkg/encrypt"
	"mim/service/user"
)

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
