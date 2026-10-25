# Mim AI 协同助手与 6 大顶级 UI 工艺规范落地实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 Mim 接入基于内部上下文流水的原生 AI 协同能力（智能帮回 + 会话总结 + 用户自定义大模型配置），并全端落地 6 项顶级 UI 交互规范（输入法回车防误触、智能滚动吸底、90ms 模糊流式字效、单层 inset 描边、同心几何圆角、分栏防卡与长文件名中间截断）。

**Architecture:** 
- 后端新增 `backend/service/ai` 模块，提供 OpenAI-compatible 协议的本地中转，从 SQLite 读取真实会话上下文生成提示词，避免前端跨域并保障 API Key 安全；
- 前端扩展 `SettingsModal.vue` 提供多厂商模型配置，在 `ChatStage.vue` 输入框工具栏植入「AI 帮回」浮层及 90ms 模糊晕开流式微动效，顶部支持抽屉式「AI 总结」；
- 全局优化聊天舞台的输入法事件、滚动跟踪器与分栏拖拽手势。

**Tech Stack:** Go 1.22, GORM, SQLite, Vue 3, TypeScript, Tailwind CSS, Tauri v2.

---

### Task 1: 后端 AI 配置与中转服务模块 (`backend/service/ai`)

**Files:**
- Create: `backend/service/ai/ai_service.go`
- Create: `backend/service/ai/ai_test.go`
- Modify: `backend/cmd/server/main.go:168-185`, `backend/cmd/server/main.go:1407-1510`

- [ ] **Step 1: 编写后端 AI 服务的单元测试**

在 `backend/service/ai/ai_test.go` 中，构建测试用例，使用 `httptest.Server` 模拟兼容 OpenAI 的 `/v1/chat/completions` 返回，验证提示词拼接与 JSON 提取：

```go
package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAIService_GenerateReply(t *testing.T) {
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
						"content": `[
							{"type": "positive", "label": "赞同确认", "text": "没问题，我稍后处理！"},
							{"type": "alternative", "label": "委婉推迟", "text": "今天日程较满，明天上午沟通可以吗？"},
							{"type": "clarify", "label": "追问细节", "text": "好的，想先确认下预计交付时间？"}
						]`,
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	svc := NewAIService(nil)
	cfg := AIConfig{
		BaseURL: mockServer.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}

	messages := []ChatMessageContext{
		{SenderName: "Alice", Content: "方案写好了，麻烦过目一下"},
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
}
```

- [ ] **Step 2: 运行测试以确认失败**

Run: `cd backend; go test ./service/ai -v`  
Expected: FAIL ("cannot find package" or undefined `NewAIService`)

- [ ] **Step 3: 实现 `backend/service/ai/ai_service.go`**

实现配置读取/存储、OpenAI-compatible HTTP 客户端转发、上下文抽取与 JSON 规范化解析：

```go
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

type AIConfig struct {
	Provider    string  `json:"provider"`
	BaseURL     string  `json:"base_url"`
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

type ChatMessageContext struct {
	SenderName string `json:"sender_name"`
	Content    string `json:"content"`
	IsMe       bool   `json:"is_me"`
}

type ReplySuggestion struct {
	Type  string `json:"type"`  // positive | alternative | clarify
	Label string `json:"label"` // 标签名
	Text  string `json:"text"`  // 建议正文
}

type AIService struct {
	db     *gorm.DB
	client *http.Client
}

func NewAIService(db *gorm.DB) *AIService {
	return &AIService{
		db: db,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (s *AIService) CallSmartReply(cfg AIConfig, contexts []ChatMessageContext) ([]ReplySuggestion, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	endpoint := baseURL
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = baseURL + "/chat/completions"
	}

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
		"temperature": 0.7,
	}

	reqBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(context.Background(), "POST", endpoint, bytes.NewReader(reqBytes))
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

	bodyBytes, err := io.ReadAll(resp.Body)
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

	rawContent := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	rawContent = strings.TrimPrefix(rawContent, "```json")
	rawContent = strings.TrimPrefix(rawContent, "```")
	rawContent = strings.TrimSuffix(rawContent, "```")
	rawContent = strings.TrimSpace(rawContent)

	var suggestions []ReplySuggestion
	if err := json.Unmarshal([]byte(rawContent), &suggestions); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON response: %w, raw: %s", err, rawContent)
	}
	return suggestions, nil
}

func (s *AIService) CallSmartSummary(cfg AIConfig, contexts []ChatMessageContext, lang string) (string, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.deepseek.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "deepseek-chat"
	}

	baseURL := strings.TrimSuffix(cfg.BaseURL, "/")
	endpoint := baseURL
	if !strings.HasSuffix(endpoint, "/chat/completions") {
		endpoint = baseURL + "/chat/completions"
	}

	var historyLines []string
	for _, m := range contexts {
		historyLines = append(historyLines, fmt.Sprintf("%s: %s", m.SenderName, m.Content))
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
		"temperature": 0.5,
	}

	reqBytes, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(context.Background(), "POST", endpoint, bytes.NewReader(reqBytes))
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

	bodyBytes, err := io.ReadAll(resp.Body)
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
```

- [ ] **Step 4: 重新运行测试以确认通过**

Run: `cd backend; go test ./service/ai -v`  
Expected: PASS

- [ ] **Step 5: 在 `backend/cmd/server/main.go` 中注册 HTTP 路由**

修改 `backend/cmd/server/main.go`：
1. 注册 `s.aiService = ai.NewAIService(s.db)`；
2. 增加 `/api/ai/config` 处理方法（读取和写入当前用户的 `mim_user_configs`）；
3. 增加 `/api/ai/reply` 处理方法（从 `mim_messages` 拉取最新 8 条文字消息后调用 `s.aiService.CallSmartReply`）；
4. 升级 `/api/ai/summary`：若用户配置了 `api_key` 则调用真实 LLM，若无则平滑回退至现有本地规则摘要。

- [ ] **Step 6: 编译并验证 Go 后端**

Run: `cd backend; go build -o server.exe ./cmd/server`  
Expected: 编译成功无报错。

- [ ] **Step 7: 提交代码**

```bash
git add backend/service/ai backend/cmd/server/main.go
git commit -m "feat(ai): implement AI service for smart reply and summary with OpenAI-compatible backend proxy"
```

---

### Task 2: 前端 AI 配置面板实现 (`SettingsModal.vue`)

**Files:**
- Modify: `clients/desktop-web/src/components/SettingsModal.vue`
- Modify: `clients/desktop-web/src/stores/chat.ts`

- [ ] **Step 1: 在 `chat.ts` 或新 Store 中提供 AI 配置请求与持久化管理**

增加 `fetchAIConfig()` 与 `saveAIConfig(config)`：
- 支持预设：
  - DeepSeek (`https://api.deepseek.com/v1`, `deepseek-chat`)
  - Kimi (`https://api.moonshot.cn/v1`, `moonshot-v1-8k`)
  - OpenAI (`https://api.openai.com/v1`, `gpt-4o-mini`)
  - Ollama (`http://localhost:11434/v1`, `qwen2.5:7b`)
  - 自定义

- [ ] **Step 2: 在 `SettingsModal.vue` 中集成「AI 引擎」标签页**

遵循 UI 工艺准则：
- 容器子按钮无边框，采用悬停变底色；
- 描边使用 `box-shadow: inset 0 0 0 1px ...`；
- 同心圆角几何：外容器 `rounded-2xl`，内边距 `p-3`，输入框与卡片采用 `rounded-xl`；
- 提供「连通性测试」按钮，点击向 `/api/ai/reply` 发送一次健康探测并实时显示响应耗时与连通状态。

- [ ] **Step 3: 运行前端构建测试**

Run: `npm --prefix clients/desktop-web run build`  
Expected: 编译通过，无类型错误。

- [ ] **Step 4: 提交代码**

```bash
git add clients/desktop-web/src/components/SettingsModal.vue clients/desktop-web/src/stores/chat.ts
git commit -m "feat(ai): add AI model configuration settings with presets and connection tester"
```

---

### Task 3: 前端「AI 帮回」浮层与 90ms 模糊流式字效

**Files:**
- Create: `clients/desktop-web/src/components/AISmartReplyBar.vue`
- Modify: `clients/desktop-web/src/components/ChatStage.vue`

- [ ] **Step 1: 编写 `AISmartReplyBar.vue`**

实现具备 3 条候选气泡的轻量面板：
1. 接收 `suggestions: Array<{ type: string; label: string; text: string }>`；
2. 每一个字符渲染为 `<span class="ai-blur-char" :style="{ '--char-index': idx }">{{ char }}</span>`；
3. CSS 规则：
   ```css
   @keyframes aiBlurFadeIn {
     0% {
       filter: blur(6px);
       opacity: 0;
       transform: translateY(2px);
     }
     100% {
       filter: blur(0);
       opacity: 1;
       transform: translateY(0);
     }
   }
   .ai-blur-char {
     display: inline-block;
     animation: aiBlurFadeIn 90ms cubic-bezier(.215, .61, .355, 1) both;
     animation-delay: calc(var(--char-index) * 30ms);
   }
   ```
4. 交互：
   - 点击气泡文本：触发 `emit('select', text)`，将文本填入输入框并保持光标在末尾；
   - 点击气泡右侧微型发送图标：触发 `emit('send', text)`，直接一键发出。

- [ ] **Step 2: 在 `ChatStage.vue` 工具栏挂载 `[ AI 帮回 ]` 触发按钮**

在输入框上方工具栏加入按钮：
- 点击后若正在生成显示骨架微光态，调用 `/api/ai/reply` 成功后滑出 `AISmartReplyBar`；
- 支持按 `Esc` 或点击外部自动收起候选条。

- [ ] **Step 3: 编译与样式验证**

Run: `npm --prefix clients/desktop-web run build`  
Expected: 编译成功。

- [ ] **Step 4: 提交代码**

```bash
git add clients/desktop-web/src/components/AISmartReplyBar.vue clients/desktop-web/src/components/ChatStage.vue
git commit -m "feat(ai): add AI Smart Reply bar with 90ms blur-in typewriter animation"
```

---

### Task 4: 前端「AI 总结」结构化纪要抽屉面板

**Files:**
- Create: `clients/desktop-web/src/components/AISummaryDrawer.vue`
- Modify: `clients/desktop-web/src/components/ChatStage.vue`

- [ ] **Step 1: 编写 `AISummaryDrawer.vue`**

实现右侧滑出抽屉：
- 标题栏：带有会话名称、分析条数、时间戳与关闭按钮；
- 内容区：渲染 Markdown 格式的提炼纪要（议题脉络、核心结论、Action Items 任务清单）；
- 底部操作栏：
  - 「复制为 Markdown」：一键写入剪贴板；
  - 「发送到群聊/当前会话」：将提炼好的纪要直接作为富文本卡片发送到聊天中。

- [ ] **Step 2: 在 `ChatStage.vue` 顶部操作栏接驳 `AISummaryDrawer`**

点击顶部现有的「AI 总结」图标时打开抽屉，并触发后端 `/api/ai/summary` 动态抽取。

- [ ] **Step 3: 编译验证**

Run: `npm --prefix clients/desktop-web run build`  
Expected: 编译成功。

- [ ] **Step 4: 提交代码**

```bash
git add clients/desktop-web/src/components/AISummaryDrawer.vue clients/desktop-web/src/components/ChatStage.vue
git commit -m "feat(ai): add structured AI summary drawer with copy and share actions"
```

---

### Task 5: 落地 6 大顶级 UI 交互规范 (`ChatStage.vue`, `FileMessageCard.vue`)

**Files:**
- Modify: `clients/desktop-web/src/components/ChatStage.vue`
- Modify: `clients/desktop-web/src/components/FileMessageCard.vue`

- [ ] **Step 1: 规范 1——中文输入法选字回车防误触**

在 `ChatStage.vue` 的 `textarea` 输入处理中完善：
```ts
let isComposing = false;
let compositionCooldown = false;
let cooldownTimer: any = null;

const onCompositionStart = () => {
  isComposing = true;
  compositionCooldown = true;
};

const onCompositionEnd = () => {
  isComposing = false;
  if (cooldownTimer) clearTimeout(cooldownTimer);
  cooldownTimer = setTimeout(() => {
    compositionCooldown = false;
  }, 500); // 选字结束后 500ms 内的回车彻底拦截
};

const onInputKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    if (isComposing || compositionCooldown || e.keyCode === 229) {
      // 严禁发送，阻止默认行为
      e.preventDefault();
      return;
    }
    e.preventDefault();
    sendMessage();
  }
};
```

- [ ] **Step 2: 规范 2——智能滚动吸底与「跳至最新」悬浮药丸**

在 `ChatStage.vue` 消息滚动容器中：
1. 监听 `@scroll` 事件：
   ```ts
   const onMessagesScroll = (e: Event) => {
     const target = e.target as HTMLElement;
     const dist = target.scrollHeight - target.scrollTop - target.clientHeight;
     // 离底部 40px 以内才自动跟随新消息
     autoFollowBottom.value = dist <= 40;
     // 离底部超过 120px 时显示「跳至最新」
     showJumpToBottom.value = dist > 120;
   };
   ```
2. 新消息推入时：只有当 `autoFollowBottom.value === true` 才调用 `scrollToBottom()`；
3. 当 `showJumpToBottom.value === true` 时，右下角渲染悬浮药丸按钮 `[ ↓ 跳至最新 ]`，点击瞬间平滑滚动到底部。

- [ ] **Step 3: 规范 3 & 4——单层边框与 inset box-shadow、同心几何圆角**

检查并重构 `ChatStage.vue` 与相关卡片：
- 按钮消除外层独立实体 `border`，采用 `hover:bg-black/5 dark:hover:bg-white/5` 与 `shadow-[inset_0_0_0_1px_var(--color-primary)]`；
- 圆角严格遵循 $R_{外} = R_{内} + P_{内边距}$。

- [ ] **Step 4: 规范 5——分栏拖拽极致防卡**

在拖动分栏事件处理逻辑中：
1. `onMouseDown` 时在 `body` 挂载 `fixed inset-0 z-50 cursor-col-resize select-none` 透明遮罩；
2. 监听 `Esc` 恢复拖拽前的初始比例；
3. 双击拖拽条复原至 50/50；
4. 当左栏占比超过 82% 时，自动收起右侧面板。

- [ ] **Step 5: 规范 6——长文件名中间省略号截断**

在 `FileMessageCard.vue` 与文件上传预览中，编写工具函数 `formatMiddleEllipsis(fileName: string, maxLen = 22)`：
```ts
export const formatMiddleEllipsis = (name: string, maxLen = 22): string => {
  if (!name || name.length <= maxLen) return name;
  const lastDot = name.lastIndexOf('.');
  const ext = lastDot !== -1 ? name.slice(lastDot) : '';
  const base = lastDot !== -1 ? name.slice(0, lastDot) : name;
  const keepEnd = Math.max(ext.length + 3, 6);
  const keepStart = maxLen - keepEnd - 3;
  if (keepStart <= 2) return name.slice(0, 10) + '...' + ext;
  return base.slice(0, keepStart) + '...' + base.slice(-3) + ext;
};
```

- [ ] **Step 6: 前端编译与类型校验**

Run: `npm --prefix clients/desktop-web run build`  
Expected: 0 报错，编译通过。

- [ ] **Step 7: 提交代码**

```bash
git add clients/desktop-web/src/components/ChatStage.vue clients/desktop-web/src/components/FileMessageCard.vue
git commit -m "feat(ui): enforce 6 craft guidelines for IME enter guard, smart scroll, splitter and filenames"
```

---

### Task 6: 联调测试、回归验证与桌面打包

**Files:**
- All modified backend and frontend files

- [ ] **Step 1: 运行全量后端测试**

Run: `cd backend; go test ./...`  
Expected: 全部测试通过。

- [ ] **Step 2: 启动后端并测试 API 端点**

Run: 启动 `server.exe`，执行测试脚本向 `/api/ai/config`、`/api/ai/reply` 与 `/api/ai/summary` 发送请求，验证 JSON 协议。

- [ ] **Step 3: 构建桌面端 Release 安装包**

Run: `npm --prefix clients/desktop-web run tauri build`  
Expected: 重新生成最新的 `mim-desktop.exe` 及 NSIS/MSI 安装包。

- [ ] **Step 4: 最终提交与标签**

```bash
git add .
git commit -m "chore: release AI assistant and UI craft polish v1.1.0"
```
