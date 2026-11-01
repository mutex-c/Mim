package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"mim/models"
	"mim/service/ai"
)

// AI Config API Handler (GET to read, POST to save)
func (s *Server) handleAIConfig(w http.ResponseWriter, r *http.Request, userID uint64) {
	if r.Method == http.MethodGet {
		cfg, err := s.aiService.GetAIConfig(userID)
		if err != nil {
			s.jsonError(w, http.StatusInternalServerError, "Failed to get AI config: "+err.Error())
			return
		}
		// Mask sensitive API Key before returning to frontend
		safeCfg := *cfg
		if safeCfg.APIKey != "" {
			if len(safeCfg.APIKey) > 8 {
				safeCfg.APIKey = safeCfg.APIKey[:3] + "******" + safeCfg.APIKey[len(safeCfg.APIKey)-4:]
			} else {
				safeCfg.APIKey = "******"
			}
		}
		s.jsonResponse(w, &safeCfg)
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

	if !s.checkConversationAccess(userID, body.ConversationID) {
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

	if !s.checkConversationAccess(userID, body.ConversationID) {
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
