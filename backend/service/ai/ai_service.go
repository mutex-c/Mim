package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AIConfig represents the configuration for LLM provider
type AIConfig struct {
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

// ChatMessageContext represents an extracted chat message for AI prompt context
type ChatMessageContext struct {
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	IsMe       bool   `json:"is_me"`
}

// ReplySuggestion represents a single quick-reply candidate
type ReplySuggestion struct {
	Type  string `json:"type"`  // positive | alternative | clarify
	Label string `json:"label"` // 标签名
	Text  string `json:"text"`  // 建议正文
}

// UserAIConfig is the database entity for persisting user AI settings
type UserAIConfig struct {
	UserID      uint64    `gorm:"primaryKey;column:user_id" json:"user_id"`
	Provider    string    `gorm:"type:varchar(64);not null;default:'deepseek';column:provider" json:"provider"`
	BaseURL     string    `gorm:"type:varchar(255);not null;default:'https://api.deepseek.com/v1';column:base_url" json:"base_url"`
	APIKey      string    `gorm:"type:varchar(255);not null;default:'';column:api_key" json:"api_key"`
	Model       string    `gorm:"type:varchar(128);not null;default:'deepseek-chat';column:model" json:"model"`
	Temperature float64   `gorm:"type:float;not null;default:0.7;column:temperature" json:"temperature"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (UserAIConfig) TableName() string {
	return "mim_user_ai_configs"
}

// DefaultReplySuggestions provides fallback suggestions when AI is unavailable
var DefaultReplySuggestions = []ReplySuggestion{
	{Type: "positive", Label: "确认同意", Text: "收到，这个方案没问题，我稍后整理好发你！"},
	{Type: "alternative", Label: "委婉补充", Text: "整体思路挺好，不过关于时间可能需要再对一下，稍后聊聊？"},
	{Type: "clarify", Label: "追问细节", Text: "好的！另外想确认下具体的细节和时间节点怎么安排？"},
}

// AIService manages AI interactions and configurations
type AIService struct {
	db     *gorm.DB
	client *http.Client
}

// NewAIService constructs a new AIService instance and auto-migrates UserAIConfig
func NewAIService(db *gorm.DB) *AIService {
	if db != nil {
		_ = db.AutoMigrate(&UserAIConfig{})
	}
	return &AIService{
		db: db,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// GetAIConfig returns user's stored AI config or sensible defaults
func (s *AIService) GetAIConfig(userID uint64) (*AIConfig, error) {
	cfg := &AIConfig{
		Provider:    "deepseek",
		BaseURL:     "https://api.deepseek.com/v1",
		APIKey:      "",
		Model:       "deepseek-chat",
		Temperature: 0.7,
	}

	if s.db == nil {
		return cfg, nil
	}

	var rec UserAIConfig
	if err := s.db.First(&rec, "user_id = ?", userID).Error; err == nil {
		if rec.Provider != "" {
			cfg.Provider = rec.Provider
		}
		if rec.BaseURL != "" {
			cfg.BaseURL = rec.BaseURL
		}
		cfg.APIKey = rec.APIKey
		if rec.Model != "" {
			cfg.Model = rec.Model
		}
		if rec.Temperature > 0 {
			cfg.Temperature = rec.Temperature
		}
	}
	return cfg, nil
}

// SaveAIConfig saves or updates user's AI settings
func (s *AIService) SaveAIConfig(userID uint64, cfg *AIConfig) error {
	if s.db == nil {
		return errors.New("database not initialized")
	}

	trimmedKey := strings.TrimSpace(cfg.APIKey)
	var rec UserAIConfig
	err := s.db.First(&rec, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		rec = UserAIConfig{
			UserID:      userID,
			Provider:    cfg.Provider,
			BaseURL:     cfg.BaseURL,
			APIKey:      trimmedKey,
			Model:       cfg.Model,
			Temperature: cfg.Temperature,
			UpdatedAt:   time.Now(),
		}
		return s.db.Create(&rec).Error
	} else if err != nil {
		return err
	}

	rec.Provider = cfg.Provider
	rec.BaseURL = cfg.BaseURL
	// Preserve existing API key if incoming key is empty
	if trimmedKey != "" {
		rec.APIKey = trimmedKey
	}
	rec.Model = cfg.Model
	if cfg.Temperature > 0 {
		rec.Temperature = cfg.Temperature
	}
	rec.UpdatedAt = time.Now()
	return s.db.Save(&rec).Error
}

func resolveChatEndpoint(baseURL string) string {
	u := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if u == "" {
		return "https://api.deepseek.com/v1/chat/completions"
	}
	if strings.HasSuffix(u, "/chat/completions") {
		return u
	}
	return u + "/chat/completions"
}

func cleanJSONResponse(raw string) string {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "```") {
		lines := strings.Split(text, "\n")
		if len(lines) >= 2 {
			lines = lines[1:]
			if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
				lines = lines[:len(lines)-1]
			}
			text = strings.TrimSpace(strings.Join(lines, "\n"))
		}
	}
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start != -1 && end != -1 && end >= start {
		text = text[start : end+1]
	}
	return text
}

// CallSmartReply invokes the OpenAI-compatible endpoint to generate 3 reply suggestions
func (s *AIService) CallSmartReply(ctx context.Context, cfg AIConfig, contexts []ChatMessageContext) ([]ReplySuggestion, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	if cfg.Temperature <= 0 || cfg.Temperature > 2.0 {
		cfg.Temperature = 0.7
	}

	endpoint := resolveChatEndpoint(cfg.BaseURL)

	var historyLines []string
	for _, m := range contexts {
		prefix := m.SenderName
		if m.IsMe {
			prefix = "我"
		}
		historyLines = append(historyLines, fmt.Sprintf("%s: %s", prefix, m.Content))
	}
	historyText := strings.Join(historyLines, "\n")

	systemPrompt := `你是一位专业、情商极高的高效即时通讯助手。
根据用户提供的最新真实聊天记录，请为“我”生成 3 条精简、恰当、符合口语聊天习惯的不同回复候选。
请严格输出合法的 JSON 数组，格式如下：
[
  {"type": "positive", "label": "赞同/高效确认", "text": "回复内容"},
  {"type": "alternative", "label": "委婉推迟/补充考虑", "text": "回复内容"},
  {"type": "clarify", "label": "追问细节/幽默化解", "text": "回复内容"}
]
不要输出任何其他解释文字或 Markdown 标记。`

	reqBody := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "最新聊天记录流水如下：\n" + historyText},
		},
		"temperature": cfg.Temperature,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI upstream request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("AI upstream error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return nil, err
	}
	if len(chatResp.Choices) == 0 {
		return nil, errors.New("empty choices from AI provider")
	}

	cleaned := cleanJSONResponse(chatResp.Choices[0].Message.Content)

	var suggestions []ReplySuggestion
	if err := json.Unmarshal([]byte(cleaned), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w, raw: %s", err, cleaned)
	}
	return suggestions, nil
}

// CallSmartSummary invokes the OpenAI-compatible endpoint to generate structured meeting/chat minutes
func (s *AIService) CallSmartSummary(ctx context.Context, cfg AIConfig, contexts []ChatMessageContext, lang string) (string, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}
	if cfg.Temperature <= 0 || cfg.Temperature > 2.0 {
		cfg.Temperature = 0.5
	}

	endpoint := resolveChatEndpoint(cfg.BaseURL)

	var historyLines []string
	for _, m := range contexts {
		prefix := m.SenderName
		if m.IsMe {
			prefix = "我"
		}
		historyLines = append(historyLines, fmt.Sprintf("%s: %s", prefix, m.Content))
	}
	historyText := strings.Join(historyLines, "\n")

	promptLang := "中文"
	if lang == "en" {
		promptLang = "English"
	}

	systemPrompt := fmt.Sprintf(`你是一位高阶会议与会话纪要秘书。
请对以下聊天会话记录进行深度提炼，使用 %s 输出结构清晰、条理分明的 Markdown 纪要，包含：
1. 📌 **核心议题与讨论脉络**（简明扼要说明讨论了什么）
2. 💡 **关键共识与决策事项**（列出各方达成的结论）
3. 📋 **待办事项与行动清单 (Action Items)**（包含明确的任务与责任人）
语言保持专业、干练、精炼。`, promptLang)

	reqBody := map[string]interface{}{
		"model": cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "聊天记录如下：\n" + historyText},
		},
		"temperature": cfg.Temperature,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	if ctx == nil {
		ctx = context.Background()
	}
	reqCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", endpoint, bytes.NewReader(reqBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("AI summary request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("AI summary error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", err
	}
	if len(chatResp.Choices) == 0 {
		return "", errors.New("empty choices from AI provider")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
