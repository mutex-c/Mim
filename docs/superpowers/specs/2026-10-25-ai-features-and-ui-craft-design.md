# Mim AI 协同助手与 6 大顶级 UI 工艺规范设计规约

- **创建日期**: 2026-10-25
- **状态**: 已规划 (Approved Design)
- **文档路径**: `docs/superpowers/specs/2026-10-25-ai-features-and-ui-craft-design.md`

---

## 1. 概述与目标

本项目为 Mim（即时通讯系统）引入两项核心原生 AI 协同能力：
1. **AI 智能帮回 (Smart Reply)**：利用 Mim 软件内部直接读取的真实上下文流水，无需外部截图 OCR，一键生成 3 条不同风格的快捷回复候选卡片，并带有 90ms 模糊晕开流式微动效。
2. **AI 会话智能总结 (Smart Summary)**：抽取当前单聊/群聊近 30~80 条真实发言记录，调用用户配置的大模型生成包含核心议题、关键共识与待办事项的结构化纪要面板。
3. **用户自主模型配置**：支持用户在设置中心任意配置第三方及私有模型（Base URL、API Key、Model Name），预置 DeepSeek、Kimi、OpenAI、Ollama 等常用模板。
4. **全端落地 6 大顶级 UI 交互规范**：
   - 中文输入法选字回车防误触（isComposing / keyCode 229 + 500ms 冷却）；
   - 智能滚动吸底（$\le$ 40px 跟随，用户滚动打断，$>$ 120px 浮出跳至最新药丸）；
   - 单层边框与 inset box-shadow 描边，根除布局抖动；
   - 同心圆角几何规律（$R_{外} = R_{内} + P_{内边距}$）；
   - 分栏拖拽极致防卡（全屏无感遮罩 + Esc 复原 + 双击 50/50 + 82% 自动折叠）；
   - 长文件名保留后缀的智能中间截断。

---

## 2. 总体架构与数据流（方案一：后端上下文装配与中转代理）

```
[前端: ChatStage / SettingsModal]
       │
       │  (1) GET/POST /api/ai/config (读取/更新用户 AI 配置: BaseURL, APIKey, Model)
       │  (2) POST /api/ai/reply      (发送 {conversation_id, count: 3})
       │  (3) POST /api/ai/summary    (发送 {conversation_id, range: 50, lang: 'zh'|'en'})
       ▼
[Mim Go 后端: HTTP Server]
       │
       ├─► SQLite `mim_messages` / `mim_users`: 读取真实上下文流水（清洗格式/过滤撤回与杂音）
       ├─► 组装 OpenAI 兼容协议 Payload (`/chat/completions`)
       ├─► 根据用户保存在 `mim_user_configs` 的 Base URL 与 Key 代为发起外部请求
       ▼
[目标大模型] (DeepSeek / Kimi / OpenAI / Ollama / 公司内部网关)
       │
       ▼ (响应流或 JSON 数据返回后端)
[Mim Go 后端代理] ──► 返回标准 JSON / SSE 给前端
```

### 架构优势
1. **100% 免疫浏览器跨域（CORS）**：彻底杜绝前端直接发起第三方 API 产生的 `TypeError: Failed to fetch`。
2. **API 密钥本地化安全存储**：API Key 仅保存在客户端本地 SQLite 中，不经由非信任云端转发。
3. **精确上下文提取**：从数据库底层直接按时序提取准确的发送人姓名、内容及角色，规避前端翻页截断。

---

## 3. 详细接口设计

### 3.1 用户 AI 配置接口
- `GET /api/ai/config?user_id={uid}`
  - 返回：
    ```json
    {
      "code": 0,
      "data": {
        "provider": "deepseek",
        "base_url": "https://api.deepseek.com/v1",
        "api_key": "sk-***",
        "model": "deepseek-chat",
        "temperature": 0.7
      }
    }
    ```
- `POST /api/ai/config?user_id={uid}`
  - 入参：`{ provider, base_url, api_key, model, temperature }`
  - 逻辑：更新或创建在 `mim_user_configs` 表中。

### 3.2 智能帮回接口
- `POST /api/ai/reply?user_id={uid}`
  - 入参：
    ```json
    {
      "conversation_id": "1_2",
      "count": 3,
      "tone": "auto"
    }
    ```
  - 后端处理：
    - 读取该会话最近 8 条未撤回的文字消息。
    - 组装 System Prompt，严格要求输出 JSON 格式包含 3 个选项：
      1. `positive`：赞同、确认、高效推进；
      2. `alternative`：委婉推迟、补充顾虑或提出替代建议；
      3. `clarify`：追问细节、幽默或友好回应。
  - 返回格式：
    ```json
    {
      "code": 0,
      "data": {
        "suggestions": [
          { "type": "positive", "label": "确认同意", "text": "收到，这个方案没问题，我稍后整理好发你！" },
          { "type": "alternative", "label": "委婉补充", "text": "整体思路挺好，不过关于上线时间可能需要再对一下，明天上午聊聊？" },
          { "type": "clarify", "label": "追问细节", "text": "好的！另外想确认下第三步提到的设计稿由哪边最终定稿呢？" }
        ]
      }
    }
    ```

### 3.3 智能总结接口（升级现有端点）
- `POST /api/ai/summary?user_id={uid}`
  - 入参：`{ conversation_id: "...", lang: "zh"|"en", limit: 50 }`
  - 若用户未配置有效 API Key：平滑降级至现有的本地规则聚合摘要；
  - 若用户配置了有效 API Key：调用大模型提炼核心议题、关键结论、行动清单并返回 Markdown 结构化文本。

---

## 4. 前端组件与交互规范

### 4.1 AI 模型配置面板（`SettingsModal.vue`）
- 增加「AI 助手」设置标签页。
- 提供厂商快速预设切换：
  - **DeepSeek**：`https://api.deepseek.com/v1` | `deepseek-chat`
  - **Kimi (Moonshot)**：`https://api.moonshot.cn/v1` | `moonshot-v1-8k`
  - **OpenAI**：`https://api.openai.com/v1` | `gpt-4o-mini`
  - **Ollama 本地**：`http://localhost:11434/v1` | `qwen2.5:7b` (免 Key)
  - **自定义 (Custom)**：自由输入
- 连通性测试按钮：直接测试连通状态并反馈延迟。

### 4.2 AI 帮回输入建议条（`ChatStage.vue`）
- **UI 布局**：在消息输入框（Textarea）顶部工具栏居左放置 `[ AI 帮回 ]` 触发按钮。
- **候选浮层**：点击后在输入框上方滑出弹性卡片容器，横向展示 3 个建议气泡。
- **流式字效微动效**（90ms 模糊晕开）：
  - 每一个字符应用 CSS 动画：`from { filter: blur(6px); opacity: 0; } to { filter: blur(0); opacity: 1; }`。
  - 动画时长 `90ms`，缓动曲线 `cubic-bezier(.215, .61, .355, 1)`。
  - 邻近字符设置 `animation-delay: calc(var(--char-index) * 30ms)`。
- **操作反馈**：
  - 点击卡片文本：瞬间填入输入框并选中聚焦，用户可直接二次修改。
  - 卡片右侧微型快速发送图标：直接将选中文本作为消息发送。

### 4.3 AI 会话纪要侧边抽屉（`AISummaryDrawer.vue`）
- 点击顶部「AI 总结」图标，右侧滑出抽屉。
- 卡片内部清晰排版：
  1. 📌 **核心议题与讨论背景**
  2. 💡 **阶段共识与决议**
  3. 📋 **待办事项 (Action Items)** 与责任人
- 底部操作栏：一键复制全文、一键发送卡片至当前聊天。

---

## 5. 6 大顶级 UI 工艺规范落地清单

### 规范 1：中文输入法选字回车防误触
- 在输入框绑定 `@compositionstart`、`@compositionend` 与 `@keydown.enter`。
- 维护 `isComposing: boolean` 与 `compositionCooldown: boolean`。
- 当 `isComposing === true` 或 `e.keyCode === 229` 时，拦截回车；
- 在 `compositionend` 触发后，启动 500ms 定时器锁：在 500ms 内按回车继续拦截，超时后解锁。

### 规范 2：智能滚动吸底与跳至最新药丸
- 实时计算滚动条距离底部高度：`distanceFromBottom = scrollHeight - scrollTop - clientHeight`。
- 若 `distanceFromBottom <= 40`，新消息到达自动平滑吸底；
- 用户主动上滑时，自动锁定当前视窗，不被新消息强制拉到底部；
- 当 `distanceFromBottom > 120` 时，右下角平滑淡入「跳至最新」悬浮药丸（带未读新消息数量提示）。点击瞬间滚动至底部。

### 规范 3：单层边框与 inset box-shadow 描边
- 容器内子按钮一律 `border: none`，以背景色变化（`hover:bg-black/5 dark:hover:bg-white/5`）表达悬浮态。
- 选中/激活态描边使用 `box-shadow: inset 0 0 0 1px var(--accent-color)`，杜绝 `border` 切换导致的 1px 抖动与重排。

### 规范 4：同心圆角几何规律
- 严格贯彻 $R_{外} = R_{内} + P_{内边距}$。
- 例如：外容器 `rounded-2xl` (16px)，内边距 `p-2` (8px)，内部元素必须采用 `rounded-lg` (8px)。

### 规范 5：分栏拖拽极致防卡
- 在拖拽分栏（Splitter）事件 `mousedown` 时，在整个右侧视窗挂载一层全屏透明遮罩 `fixed inset-0 cursor-col-resize z-50`，解决鼠标移动被内层元素捕获丢失事件的问题。
- 监听 `Esc` 键：拖拽未松开或变动后按 Esc 还原初始比例。
- 双击拖拽手柄：复位至 50/50 居中平分。
- 拖拽比例检测：左栏宽度超过 82% 时，平滑将右栏收起至折叠态。

### 规范 6：文件名智能中间截断
- 提取文件名并拆分扩展名：`const ext = filename.split('.').pop()`。
- 当文件名超出长度时，截断格式为：`前缀字符串 + ... + 后缀名`（如 `2026年第四季度研发立项汇报...v2.xlsx`），始终完整保障文件类型与关键版本号辨识。

---

## 6. 验证与测试标准
1. **AI 配置连通性**：配置 DeepSeek / Ollama 后，测试连通返回 200，API Key 隐蔽显示。
2. **帮回准确性**：对真实对话点击「AI 帮回」，3 个选项必须在 1.5s 内返回，并呈现 90ms 模糊流式文字动画。
3. **输入法测试**：在 Windows 微软拼音/搜狗输入法打字打拼音时按回车，消息严禁被发出；选字完成后 300ms 内狂按回车，同样不误发。
4. **滚动吸底测试**：上滑浏览历史消息时，来新消息页面不动；点击「跳至最新」平滑吸底。
5. **分栏拖拽测试**：快速拖动无掉帧失焦，按 Esc 能恢复原比例。
