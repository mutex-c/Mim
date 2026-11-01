package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"mim/models"
)

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
	if !s.checkConversationAccess(userID, convID) {
		s.jsonError(w, http.StatusForbidden, "Access denied: you are not a participant in this conversation")
		return
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
	if !s.checkConversationAccess(userID, body.ConversationID) {
		s.jsonError(w, http.StatusForbidden, "Access denied: you are not a participant in this conversation")
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
	if !s.checkConversationAccess(userID, body.ConversationID) {
		s.jsonError(w, http.StatusForbidden, "Access denied: you are not a participant in this conversation")
		return
	}

	var msg models.Message
	var err error
	if body.MsgID > 0 {
		err = s.db.Where("id = ? AND conversation_id = ?", body.MsgID, body.ConversationID).First(&msg).Error
	} else if body.ClientMsgID != "" {
		err = s.db.Where("client_msg_id = ? AND conversation_id = ?", body.ClientMsgID, body.ConversationID).First(&msg).Error
	} else {
		s.jsonError(w, http.StatusBadRequest, "msg_id or client_msg_id required")
		return
	}

	if err != nil {
		s.jsonError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Verify message ownership or group admin privilege
	if msg.SenderID != userID {
		canDelete := false
		if strings.HasPrefix(body.ConversationID, "group_") {
			groupID, _ := strconv.ParseUint(strings.TrimPrefix(body.ConversationID, "group_"), 10, 64)
			var mem models.GroupMember
			if err := s.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&mem).Error; err == nil {
				if mem.Role == 2 || mem.Role == 3 {
					canDelete = true
				}
			}
		}
		if !canDelete {
			s.jsonError(w, http.StatusForbidden, "Access denied: cannot delete other user's message")
			return
		}
	}

	s.db.Delete(&msg)
	s.jsonResponse(w, map[string]string{"status": "deleted"})
}
