package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"mim/models"
)

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
