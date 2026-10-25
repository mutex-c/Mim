package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAIService_GenerateReply(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" && r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "```json\n[\n" +
							`  {"type": "positive", "label": "赞同确认", "text": "没问题，我稍后处理！"},` + "\n" +
							`  {"type": "alternative", "label": "委婉推迟", "text": "今天日程较满，明天上午沟通可以吗？"},` + "\n" +
							`  {"type": "clarify", "label": "追问细节", "text": "好的，想先确认下预计交付时间？"}` + "\n" +
							"]\n```",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	svc := NewAIService(nil)
	cfg := AIConfig{
		BaseURL: mockServer.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}

	messages := []ChatMessageContext{
		{SenderName: "Alice", Content: "方案写好了，麻烦过目一下", IsMe: false},
	}

	suggestions, err := svc.CallSmartReply(cfg, messages)
	if err != nil {
		t.Fatalf("CallSmartReply failed: %v", err)
	}
	if len(suggestions) != 3 {
		t.Fatalf("expected 3 suggestions, got %d", len(suggestions))
	}
	if suggestions[0].Label != "赞同确认" {
		t.Errorf("unexpected label: %s", suggestions[0].Label)
	}
	if suggestions[0].Type != "positive" {
		t.Errorf("unexpected type: %s", suggestions[0].Type)
	}
	if suggestions[1].Type != "alternative" {
		t.Errorf("unexpected type: %s", suggestions[1].Type)
	}
	if suggestions[2].Type != "clarify" {
		t.Errorf("unexpected type: %s", suggestions[2].Type)
	}
}

func TestAIService_SmartSummary(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" && r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]string{
						"content": "### 📌 核心议题与讨论脉络\n- 讨论了客户端性能优化方案。\n\n### 💡 关键共识与决策事项\n- 同意采用虚拟滚动。\n\n### 📋 待办事项 (Action Items)\n- Alice: 编写设计文档。",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	svc := NewAIService(nil)
	cfg := AIConfig{
		BaseURL: mockServer.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}

	messages := []ChatMessageContext{
		{SenderName: "Alice", Content: "我们下午讨论下虚拟滚动方案", IsMe: false},
		{SenderName: "Bob", Content: "好的，赞成", IsMe: true},
	}

	summary, err := svc.CallSmartSummary(cfg, messages, "zh")
	if err != nil {
		t.Fatalf("CallSmartSummary failed: %v", err)
	}
	if !strings.Contains(summary, "核心议题与讨论脉络") {
		t.Errorf("summary missing expected heading: %s", summary)
	}

	summaryEn, err := svc.CallSmartSummary(cfg, messages, "en")
	if err != nil {
		t.Fatalf("CallSmartSummary (en) failed: %v", err)
	}
	if summaryEn == "" {
		t.Error("expected non-empty summary for en")
	}
}

func TestAIService_DBConfig(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite memory: %v", err)
	}

	svc := NewAIService(db)

	// Test default config
	defaultCfg, err := svc.GetAIConfig(1001)
	if err != nil {
		t.Fatalf("GetAIConfig failed: %v", err)
	}
	if defaultCfg.Provider != "deepseek" {
		t.Errorf("expected default provider 'deepseek', got %s", defaultCfg.Provider)
	}
	if defaultCfg.APIKey != "" {
		t.Errorf("expected empty default api key, got %s", defaultCfg.APIKey)
	}

	// Test save config
	toSave := AIConfig{
		Provider:    "openai",
		BaseURL:     "https://api.openai.com/v1",
		APIKey:      "sk-custom-secret-token",
		Model:       "gpt-4o-mini",
		Temperature: 0.8,
	}
	if err := svc.SaveAIConfig(1001, &toSave); err != nil {
		t.Fatalf("SaveAIConfig failed: %v", err)
	}

	loaded, err := svc.GetAIConfig(1001)
	if err != nil {
		t.Fatalf("GetAIConfig after save failed: %v", err)
	}
	if loaded.Provider != "openai" {
		t.Errorf("expected provider 'openai', got %s", loaded.Provider)
	}
	if loaded.APIKey != "sk-custom-secret-token" {
		t.Errorf("expected api_key 'sk-custom-secret-token', got %s", loaded.APIKey)
	}
	if loaded.Model != "gpt-4o-mini" {
		t.Errorf("expected model 'gpt-4o-mini', got %s", loaded.Model)
	}
	if loaded.Temperature != 0.8 {
		t.Errorf("expected temperature 0.8, got %f", loaded.Temperature)
	}
}

func TestAIService_CleanJSONResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fenced json",
			input:    "```json\n[{\"type\":\"positive\"}]\n```",
			expected: "[{\"type\":\"positive\"}]",
		},
		{
			name:     "fenced raw",
			input:    "```\n[{\"type\":\"positive\"}]\n```",
			expected: "[{\"type\":\"positive\"}]",
		},
		{
			name:     "surrounding prose",
			input:    "Here is the result:\n[{\"type\":\"positive\"}]\nHope this helps!",
			expected: "[{\"type\":\"positive\"}]",
		},
		{
			name:     "already clean",
			input:    "[{\"type\":\"positive\"}]",
			expected: "[{\"type\":\"positive\"}]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanJSONResponse(tt.input)
			if got != tt.expected {
				t.Errorf("cleanJSONResponse(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestAIService_ResolveChatEndpoint(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "https://api.deepseek.com/v1/chat/completions"},
		{"https://api.deepseek.com/v1", "https://api.deepseek.com/v1/chat/completions"},
		{"https://api.deepseek.com/v1/", "https://api.deepseek.com/v1/chat/completions"},
		{"https://custom.ai.com/chat/completions", "https://custom.ai.com/chat/completions"},
	}

	for _, tt := range tests {
		got := resolveChatEndpoint(tt.input)
		if got != tt.expected {
			t.Errorf("resolveChatEndpoint(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestAIService_UpstreamError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"quota exceeded"}`, http.StatusTooManyRequests)
	}))
	defer mockServer.Close()

	svc := NewAIService(nil)
	cfg := AIConfig{
		BaseURL: mockServer.URL,
		APIKey:  "test-key",
	}

	_, err := svc.CallSmartReply(cfg, []ChatMessageContext{{SenderName: "User", Content: "Hi"}})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("expected 429 in error message, got %v", err)
	}

	_, err = svc.CallSmartSummary(cfg, []ChatMessageContext{{SenderName: "User", Content: "Hi"}}, "zh")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "429") {
		t.Errorf("expected 429 in error message, got %v", err)
	}
}
