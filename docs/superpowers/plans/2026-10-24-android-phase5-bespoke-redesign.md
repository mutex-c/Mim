# Android Phase 5 高定美学与核心交互重塑实施计划 (Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 重构 Mim Android 原生客户端的设计系统与核心交互，全量落地 3 套高定调色板（松风墨翠、沧浪深黛、黑曜暖琥）、微明度分层质感界面，与桌面端轻奢视觉高度统一，并保证底层 24B 二进制网络协议与 Room 数据库 100% 稳定运行。

**Architecture:** 
1. 在 `ui/theme/` 建立基于 Jetpack Compose `CompositionLocal` 的高定设计令牌系统（`MimColors`、`ColorPalette`、`ThemeMode`、`MimTheme`），并在 `PreferencesHelper` 中建立持久化与响应式状态流。
2. 在 `ui/components/` 封装高复用原子组件：高对比度字母头像 `MimAvatar`、QQ 式两阶段下载打开文件卡片 `MimFileCard`、胶囊悬浮输入工具坞。
3. 逐一升级业务页面：认证页 (`AuthScreen`)、主页面与会话列表 (`MainScreen`)、聊天主舞台 (`ChatScreen`) 及设置中心 (`SettingsScreen`)。
4. 执行 Gradle 生产级编译验证，输出最新 `app-debug.apk` 并保证零错误。

**Tech Stack:** 
- Kotlin 1.9+ / Jetpack Compose / Material 3
- Room SQLite / Kotlin Coroutines & Flow
- OkHttp WebSocket / Protobuf 二进制协议
- Gradle 8.6 / Android SDK 34

---

### Task 1: 主题系统与设计令牌建设 (Design Tokens & Color Architecture)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/theme/MimColors.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/theme/Theme.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/data/local/PreferencesHelper.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/MainActivity.kt`

- [ ] **Step 1: 创建 `MimColors.kt` 设计令牌**
  * 定义 `ColorPalette` 枚举 (`PINE`, `PETROL`, `OBSIDIAN`) 与 `ThemeMode` 枚举 (`LIGHT`, `DARK`, `SYSTEM`)。
  * 定义 `MimColors` 数据类，包含 `background`, `surface`, `surfaceElevated`, `border`, `textPrimary`, `textMuted`, `accent`, `bubbleSelf`, `bubbleSelfText`, `bubblePeer`, `bubblePeerText`, `inputBackground`, `isDark`。
  * 定义 3 套调色板对应的深浅色常量（松风墨翠、沧浪深黛、黑曜暖琥，与桌面端 CSS 变量精确对齐）。

- [ ] **Step 2: 重构 `Theme.kt` 并提供全局 `MimTheme`**
  * 使用 `staticCompositionLocalOf<MimColors>` 提供 `LocalMimColors`。
  * 提供 `MimTheme.colors` 便捷静态访问器。
  * `MimTheme` Composable 接收 `palette` 与 `themeMode`，计算后分发对应 `MimColors` 并适配 Material 3 基础容器。

- [ ] **Step 3: 扩展 `PreferencesHelper.kt` 状态持久化**
  * 增加 `themeMode` 与 `palette` 属性读写；
  * 提供 `MutableStateFlow` 触发界面主题切换热重组。

- [ ] **Step 4: 在 `MainActivity.kt` 接入动态主题**
  * 收集当前持久化的 `palette` 与 `themeMode`，作为 `MimTheme` 的根节点参数传入。

- [ ] **Step 5: 验证编译**
  * 运行 `./gradlew :app:compileDebugKotlin` 确保主题层无语法与编译错误。

- [ ] **Step 6: Commit**
  * `git add clients/android/app/src/main/java/com/mim/im/ui/theme/ clients/android/app/src/main/java/com/mim/im/data/local/PreferencesHelper.kt`
  * `git commit -m "feat(android): build bespoke design tokens and 3-palette theme engine"`

---

### Task 2: 高定原子组件库实装 (Atom Components)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/MimAvatar.kt`
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/MimFileCard.kt`

- [ ] **Step 1: 实现高对比度头像组件 `MimAvatar.kt`**
  * 支持自定义尺寸、文字徽标与图片 URL。
  * 内置实色算法：根据用户 ID 或名称哈希分配高饱和度实色背景，配合纯白粗体文字，杜绝水洗透明发灰。
  * 支持右上角/右下角在线状态点或群聊徽标。

- [ ] **Step 2: 实现 QQ 式两阶段文件卡片 `MimFileCard.kt`**
  * 解析文件后缀（PDF、DOC、XLS、ZIP、CODE 等）渲染对应图标与专色。
  * 未下载状态：半透明下载浮标。
  * 下载中状态：圆形进度指示环。
  * 已下载状态：展示“已下载”，点击直接调用 Android `FileProvider` 打开。

- [ ] **Step 3: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin` 确保原子组件无报错。

- [ ] **Step 4: Commit**
  * `git add clients/android/app/src/main/java/com/mim/im/ui/components/`
  * `git commit -m "feat(android): add MimAvatar and QQ-style MimFileCard atom components"`

---

### Task 3: 设置中心与多调色板切换实装 (SettingsScreen)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/screens/SettingsScreen.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`

- [ ] **Step 1: 编写 `SettingsScreen.kt`**
  * 个人信息区域：大号 `MimAvatar`、昵称、唯一账号 ID。
  * 外观与主题设置：
    * 主题模式单选切换（跟随系统、白天模式、暗黑模式）。
    * 调色板单选卡片（松风墨翠、沧浪深黛、黑曜暖琥），附带实色圆形预览色块。
  * 网络与服务配置：展示当前服务端 URL 与连接状态，点击弹出 `ServerConfigDialog`。
  * 账号操作：退出登录按钮（危险操作红字卡片）。

- [ ] **Step 2: 在 `MainScreen.kt` 整合第三 Tab**
  * 将原占位 Profile 视图替换为完整高定版 `SettingsScreen`。

- [ ] **Step 3: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin` 验证。

- [ ] **Step 4: Commit**
  * `git add clients/android/app/src/main/java/com/mim/im/ui/screens/SettingsScreen.kt clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`
  * `git commit -m "feat(android): implement SettingsScreen with 3-palette theme selector"`

---

### Task 4: 登录注册页高定重塑 (AuthScreen Redesign)

**Files:**
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/AuthScreen.kt`

- [ ] **Step 1: 重构品牌 Header 与背景底色**
  * 采用 `MimTheme.colors.background` 替代生硬纯白/纯黑。
  * 绘制精致 Mim 盾牌矢量 Logo 与副标排版。

- [ ] **Step 2: 重构输入框手感与呼吸间距**
  * 优化账号与密码输入卡片，左侧预留 36px 呼吸间距，彻底消除图标文字重叠。
  * 替换电光蓝为当前主题 `accent` 色，输入框使用微明度卡片背景与圆角。

- [ ] **Step 3: 重构演示账号快捷登录胶囊**
  * 4 位演示账号卡片采用当前主题卡片色，展示成员头像与免密登录提示，点击秒登。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git add clients/android/app/src/main/java/com/mim/im/ui/screens/AuthScreen.kt`
  * `git commit -m "refactor(android): overhaul AuthScreen with bespoke luxury aesthetic"`

---

### Task 5: 主页面与会话列表重塑 (MainScreen Redesign)

**Files:**
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`

- [ ] **Step 1: 沉浸式微明度顶栏**
  * 顶栏背景与 `MimTheme.colors.background` 自然融合；
  * 左侧展示当前用户字母头像与在线光环；
  * 右侧呈现 WebSocket 24B 二进制流状态指示灯（常亮翠绿）。

- [ ] **Step 2: 会话列表项高定排版**
  * 头像统一接入 `MimAvatar`；
  * 标题与时间行应用微距排版，时间锁定等宽；
  * 消息预览行智能添加前缀：`[文件]`、`[语音]`、`[图片]`、`[消息已撤回]`；
  * 未读数角标采用主题强调色圆润微胶囊。

- [ ] **Step 3: 底部导航栏极简重塑**
  * 改造 BottomNavigation，移除粗厚胶囊，替换为轻盈高定微扩散背景。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git add clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`
  * `git commit -m "refactor(android): enhance MainScreen conversation list and bottom dock"`

---

### Task 6: 聊天主舞台与气泡体系重塑 (ChatScreen Redesign)

**Files:**
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/ChatScreen.kt`

- [ ] **Step 1: 气泡层次与色阶重构**
  * 己方气泡：使用 `MimTheme.colors.bubbleSelf`，右上角 4dp，其余三角 16dp，白字高清晰；
  * 对方气泡：使用 `MimTheme.colors.bubblePeer`，左上角 4dp，带 1px 超微弱边缘分割线；
  * 撤回消息：居中微灰色胶囊提示（`你撤回了一条消息`），克制整洁。

- [ ] **Step 2: 接入 `MimFileCard` 与语音播放条**
  * 文件消息替换为 `MimFileCard`；
  * 语音条优化声波高低条与播放按钮。

- [ ] **Step 3: 悬浮输入坞 (Input Dock) 重构**
  * 胶囊卡片输入框，双层微投影手感；
  * 发送按钮激活态平滑变色。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git add clients/android/app/src/main/java/com/mim/im/ui/screens/ChatScreen.kt`
  * `git commit -m "refactor(android): redesign ChatScreen message bubbles and input dock"`

---

### Task 7: 全工程构建验证与 APK 生成 (Build Verification & Delivery)

**Files:**
- Output: `clients/android/app/build/outputs/apk/debug/app-debug.apk`

- [ ] **Step 1: 执行完整 Gradle 构建**
  * 执行 `cd clients/android && ./gradlew.bat assembleDebug`。
  * 确认整个工程零 Warning，零 Error，构建状态为 `BUILD SUCCESSFUL`。

- [ ] **Step 2: 验证 APK 产出**
  * 检查产物 `clients/android/app/build/outputs/apk/debug/app-debug.apk` 的文件大小与时间戳。

- [ ] **Step 3: 最终 Commit 与交付总结**
  * `git commit -am "feat(android): complete Phase 5 bespoke redesign and generate release-ready debug APK"`
