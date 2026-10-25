# MIM 沃雅妮莎专属定制主题包实施计划 (Vodyanitsa Bespoke Theme Pack Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 MIM 客户端打造全套高保真「幽歌萦渊 · 沃雅妮莎」专属主题包，涵盖深渊剧院与水妖色彩系统、专属剧情对话底板、AI 生成的高清五星大立绘与 6 张微表情贴纸、限定祈愿抽卡系统及「金鳍座」首席角色档案展柜。

**Architecture:** 
1. 注册 `vodyanitsa` 调色盘变量，扩展 `style.css` 提供深渊剧院夜空、水纹暗斑与高对比石板剧情气泡；
2. 调度图像生成 Subagent 生成 1 张超清大立绘与 6 张 Q版微表情贴纸；
3. 在 `ChatStage.vue` 中集成专属性对话框排版、水之眼徽记、头衔映射及贴纸渲染器；
4. 构建专属「幽歌萦渊」祈愿弹窗与个人资料「金鳍座」角色档案面板。

**Tech Stack:** Vue 3 (Composition API), TypeScript, Tailwind CSS, Pinia, Chrome DevTools MCP, AI Image Generation.

---

### Task 1: AI 专属艺术资产生成与资源清单注册

**Files:**
- Create: `clients/desktop-web/src/assets/vodyanitsa/vodyanitsaStickers.ts`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/splash_vodyanitsa.png`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/vodyanitsa_sing.png`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/vodyanitsa_pout.png`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/vodyanitsa_horn.png`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/vodyanitsa_dive.png`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/vodyanitsa_tea.png`
- Create: `clients/desktop-web/src/assets/vodyanitsa/images/vodyanitsa_bow.png`

- [ ] **Step 1: 调度图像生成 Subagent 生成 1 张五星专属抽卡大立绘**
  - Prompt: 沃雅妮莎官方级全景立绘，至冬科洛列夫茨基剧团首席女高音，离群水妖，深蓝色水流微卷长发，高贵优雅的歌剧院深海礼服，身旁环绕波光粼粼的水流、冰霜晶体与金色音符。
  - Output: `clients/desktop-web/src/assets/vodyanitsa/images/splash_vodyanitsa.png`

- [ ] **Step 2: 调度图像生成 Subagent 并发生成 6 款专属 Q版微表情贴纸**
  - 表情 1：`vodyanitsa_sing.png`（高歌咏叹 · 手捧乐谱、欢快陶醉闭眼高歌、飘浮八分音符）
  - 表情 2：`vodyanitsa_pout.png`（委屈鼓脸 · 脸颊鼓鼓、水汪汪大眼睛、带点小泪珠）
  - 表情 3：`vodyanitsa_horn.png`（唤春角笛 · 吹奏金色小海螺角笛、喷涌水花特效）
  - 表情 4：`vodyanitsa_dive.png`（水妖探头 · 趴在水面冒出小脑袋、背后晃动水妖金鳍尾巴）
  - 表情 5：`vodyanitsa_tea.png`（优雅红茶 · 剧院首席端着至冬热红茶优雅微笑）
  - 表情 6：`vodyanitsa_bow.png`（华丽谢幕 · 舞台鞠躬、俏皮傲娇单眼眨眼）
  - Output: 存入 `clients/desktop-web/src/assets/vodyanitsa/images/`

- [ ] **Step 3: 编写 `vodyanitsaStickers.ts` 贴纸数据源**
  - 定义 `VodyanitsaSticker` 接口与 6 款贴纸配置（id、name、caption、imagePath）。

- [ ] **Step 4: 提交资产与配置文件**
  - `git add clients/desktop-web/src/assets/vodyanitsa/`
  - `git commit -m "feat(vodyanitsa): add bespoke artworks and chibi sticker assets"`

---

### Task 2: 调色盘注册与 CSS 视觉基质构建

**Files:**
- Modify: `clients/desktop-web/src/stores/theme.ts`
- Modify: `clients/desktop-web/src/style.css`

- [ ] **Step 1: 在 `stores/theme.ts` 中注册 `vodyanitsa` 调色盘**
  - 增加 `ColorPalette = ... | 'vodyanitsa'`。
  - 在 `PALETTES` 数组中添加：
    ```ts
    {
      id: 'vodyanitsa',
      nameZh: '幽歌萦渊 · 沃雅妮莎',
      nameEn: 'Opera Siren · Vodyanitsa',
      previewColor: '#0284c7',
      descZh: '至冬科洛列夫茨基剧团首席女高音，深渊曜夜与水妖冰泉流金',
      descEn: 'Snezhnaya Grand Opera diva, abyssal starlight with sirens crystal azure'
    }
    ```

- [ ] **Step 2: 在 `style.css` 中注入调色盘变量与舞台暗纹**
  - 浅色模式 `:root[data-palette='vodyanitsa']`
  - 深色模式 `.dark[data-palette='vodyanitsa']`
  - 水纹微光舞台背景 `.vodyanitsa-stage-bg`（纯 CSS 渐变实现深渊水波与星辉光斑）

- [ ] **Step 3: 在 `style.css` 中注入剧情底板与凯尔特水妖包角样式**
  - `.vodyanitsa-dialogue-plate`
  - `.vodyanitsa-plate-peer`（深海曜蓝石板材质，冰晶高亮内发光）
  - `.vodyanitsa-plate-self`（深曜暖金石板材质，璨金流光）
  - 水晶包角 `.vodyanitsa-corner-trim` 与晶莹水滴分割线 `.vodyanitsa-divider-strip`
  - 45° 旋转水元素神之眼切割宝石 `.vodyanitsa-vision-rhombus`

- [ ] **Step 4: 运行 `npm run build` 确保样式语法无误并提交**
  - `npm --prefix clients/desktop-web run build`
  - `git commit -m "feat(vodyanitsa): register theme palette and dialogue plate css"`

---

### Task 3: 聊天主舞台与气泡渲染逻辑集成

**Files:**
- Create: `clients/desktop-web/src/components/VodyanitsaStickerPicker.vue`
- Modify: `clients/desktop-web/src/components/ChatStage.vue`

- [ ] **Step 1: 构建 `VodyanitsaStickerPicker.vue` 贴纸抽屉组件**
  - 支持网格展示 6 款微表情；
  - 鼠标悬停放大预览与名称显示；
  - 点击向外部发射 `select` 事件。

- [ ] **Step 2: 在 `ChatStage.vue` 中集成剧情气泡渲染分支**
  - 当 `themeStore.currentPalette === 'vodyanitsa'` 时：
    - Peer 气泡渲染：水元素切割徽章（`水`）+ `沃雅妮莎` + `★★★★★` + `✦ 幽歌萦渊 · 首席女高音` + `—— 💧 ——` 分割线；
    - Self 气泡渲染：岩元素切割徽章（`岩`）+ `旅行者` + `★★★★★` + `✦ 异乡的旅人 · 首席乐迷` + `—— ◆ ——` 分割线 + 右侧专属金边「**旅**」头像；
    - 针对 `[VODYANITSA_STICKER]:<id>` 消息渲染专属水妖大贴纸卡片。

- [ ] **Step 3: 聊天工具栏挂载专属贴纸按钮**
  - 在聊天输入框操作栏添加「💧 沃雅妮莎」专属贴纸弹出面板按钮，仅在当前主题激活时呈现（或常驻可切换）。

- [ ] **Step 4: 编译验证与提交**
  - `npm --prefix clients/desktop-web run build`
  - `git commit -m "feat(vodyanitsa): integrate bespoke dialogue plates and sticker drawer"`

---

### Task 4: 专属祈愿卡池与「金鳍座」角色档案展柜

**Files:**
- Create: `clients/desktop-web/src/components/VodyanitsaWishModal.vue`
- Modify: `clients/desktop-web/src/components/UserProfileModal.vue`
- Modify: `clients/desktop-web/src/components/ChatStage.vue`

- [ ] **Step 1: 创建 `VodyanitsaWishModal.vue` 专属角色祈愿活动弹窗**
  - 卡池主题：角色活动祈愿 ·「幽歌萦渊」；
  - 核心 UP 角色：五星限定水元素法器角色「沃雅妮莎」；
  - 十连抽流星动效、出金金色光柱全屏大立绘展示；
  - 「✦ 分享战报到聊天」按钮，生成专属水纹战报卡片投递至当前会话。

- [ ] **Step 2: 在 `UserProfileModal.vue` 中为沃雅妮莎定制角色展柜**
  - 当主题为 `vodyanitsa` 时，展示至冬科洛列夫茨基剧团首席女高音背景板；
  - 展示大立绘、专属命之座「金鳍座」、生命值 65,280、水冰抗性 -35%、治疗加成 +75.4% 属性面板。

- [ ] **Step 3: 编译验证与提交**
  - `npm --prefix clients/desktop-web run build`
  - `git commit -m "feat(vodyanitsa): add exclusive wish banner and persona showcase"`

---

### Task 5: 全面端到端验证与视觉质检

**Files:**
- Verify: Chrome DevTools MCP Viewport Inspection (`http://localhost:4173/chats/1_2`)

- [ ] **Step 1: 启动桌面 Web 预览服务**
  - `npm --prefix clients/desktop-web run preview -- --port 4173`

- [ ] **Step 2: 切换至 `vodyanitsa` 主题并验证交互**
  - 验证左侧导航栏、会话列表与主舞台背景的水波暗纹；
  - 发送普通文本消息，验证旅行者（Self）与沃雅妮莎（Peer）气泡的高对比度与精致名牌；
  - 发送 6 款微表情贴纸，验证卡片渲染；
  - 点击打开「幽歌萦渊」抽卡，抽取十连抽并分享战报；
  - 打开个人资料弹窗，验证「金鳍座」角色档案展柜。

- [ ] **Step 3: 浅色与深色模式双向质检**
  - 分别切换至浅色模式与深色模式，使用 `take_screenshot` 截图对比，确保 0 透底、文字 100% 清晰锐利。

- [ ] **Step 4: 清理预览服务与最终提交**
