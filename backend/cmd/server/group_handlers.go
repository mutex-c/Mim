package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"mim/models"
	"mim/service/group"
)

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
	if grp.IsSearchable == 0 {
		var myMember models.GroupMember
		if err := s.db.Where("group_id = ? AND user_id = ?", grpID, userID).First(&myMember).Error; err != nil {
			s.jsonError(w, http.StatusForbidden, "You are not a member of this private group")
			return
		}
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

	var myMember models.GroupMember
	if err := s.db.Where("group_id = ? AND user_id = ?", grpID, userID).First(&myMember).Error; err != nil {
		s.jsonError(w, http.StatusForbidden, "You are not a member of this group")
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
