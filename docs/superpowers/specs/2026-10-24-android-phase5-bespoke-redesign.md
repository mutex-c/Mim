# Mim Android 客户端 Phase 5 高定美学与核心体验重塑规格说明书 (Design Spec)

**创建日期**：2026-10-24  
**项目分支**：Phase 5 (Android 纯原生客户端实装与视觉升级)  
**作者/设计**：Antigravity & Mim Architecture Team  
**状态**：Approved (已通过用户方案评审)

---

## 1. 目标与背景

Mim 项目立足于工业级、高可用、跨三端（Web / Windows 原生 / Android 原生）即时通讯体系。在桌面端完成 Phase 4 高定 10 条铁律重构后，系统拥有了现代、克制且反“AI 塑料感”的独特轻奢美感（温润暖石白底色、微明度分层、松风墨翠/沧浪深黛/黑曜暖琥调色板、无边框与双环微阴影）。

目前 Android 原生工程（Kotlin + Jetpack Compose）虽已完成基本业务逻辑与 APK 构建，但样式仍停留在泛用的 Material 3 默认电光蓝与高对比纯黑白，缺少与桌面端的品牌感与设计一致性。

本规格书旨在指导 Phase 5 移动端的设计系统重构与核心体验落地：
1. **统一设计语言**：将桌面端 3 套高定调色板与微明度设计系统完整迁移至 Compose。
2. **重塑界面手感**：重构登录页、会话列表、聊天主舞台、输入栏与设置页，杜绝生硬线条与 AI 味。
3. **保持底层契约**：在保证 24 字节大端序网络二进制协议帧、Room 数据库、WebSocket 长连接 100% 稳定的前提下完成无缝改造。

---

## 2. 设计系统与调色板架构 (Design Tokens)

### 2.1 调色板定义
在 Android 端建立与桌面端完全镜像的三大高定调色板体系：

| 调色板 | 中文名称 | 英文标识 | 浅色基调 | 深色基调 | 主强调色 / 己方气泡 |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Pine** (默认) | 松风墨翠 | `pine` | 暖石白 `#EBEAE3` / 表面 `#F4F3EC` | 深炭墨 `#101211` / 表面 `#151816` | 墨绿 `#1E5243` / 深翠 `#235445` |
| **Petrol** | 沧浪深黛 | `petrol` | 冰岩冷灰 `#E7ECEE` / 表面 `#EEF2F4` | 深海黑黛 `#0F1417` / 表面 `#161C20` | 深黛 `#1E4555` / 冷黛 `#225368` |
| **Obsidian** | 黑曜暖琥 | `obsidian` | 象牙浅灰 `#ECEAE7` / 表面 `#F4F2EE` | 玄武黑曜 `#131211` / 表面 `#1A1816` | 暖琥 `#8C4C1B` / 曜金 `#9C5B28` |

### 2.2 Compose 颜色令牌实体 (`MimColors`)
通过 Kotlin 数据类定义 `MimColors`：
```kotlin
data class MimColors(
    val background: Color,       // 全局画布/舞台底色 (Stage)
    val surface: Color,          // 列表与常规卡片底色 (List/Card)
    val surfaceElevated: Color,  // 浮层/弹窗底色
    val border: Color,           // 1%-3% 微对比边框线 (严禁粗硬黑色)
    val textPrimary: Color,      // 主文本色 (浅色 #191B1A，深色 #E6EBE8)
    val textMuted: Color,        // 弱化辅助文本 (浅色 #646A66，深色 #8B958F)
    val accent: Color,           // 主题强调色
    val bubbleSelf: Color,       // 己方消息气泡背景
    val bubbleSelfText: Color,   // 己方气泡文本色
    val bubblePeer: Color,       // 对方消息气泡背景 (浅色 #FFFFFF，深色微明度卡片)
    val bubblePeerText: Color,   // 对方气泡文本色
    val inputBackground: Color,  // 输入框底色
    val isDark: Boolean
)
```

### 2.3 状态注入与持久化
* 通过 `CompositionLocalProvider(LocalMimColors provides colors)` 注入全局树；
* 在 `PreferencesHelper` 中增加 `themeMode`（`system` / `light` / `dark`）与 `palette`（`pine` / `petrol` / `obsidian`），启动时由 `MimTheme` 自动监听并动态重组。

---

## 3. 具体页面重构与交互规范

### 3.1 登录与注册页 (`AuthScreen`)
1. **视觉层次**：
   * 顶部居中 Mim 矢量盾牌标识，搭配“企业级端到端即时通讯系统”副标；
   * 移除厚重边框，采用微明度悬浮卡片；
   * 模式切换采用平滑 Segmented Tab（账号登录 / 新用户注册）。
2. **输入框与交互**：
   * 12dp 大圆角卡片，左侧图标与占位文本预留 36px 独立空间，解决文字重叠问题；
   * 眼睛图标一键切换显隐密码。
3. **快捷通道与局域网支持**：
   * 下方排列 4 位预设演示账号（Alex Chen、Sarah Jenkins、Michael Zhang、David Miller），点击免密秒登；
   * 顶部保留局域网服务端配置入口，一键填入宿主机 IP 并即时测试连通性。

### 3.2 主页面与会话列表 (`MainScreen`)
1. **沉浸式顶栏**：
   * 背景色与页面自然交融，不再使用高反差纯白；
   * 左侧展示当前用户字母徽标头像与在线光圈；
   * 中间展示应用标题与会话数；
   * 右侧实时呈现 WebSocket 24B 二进制流连通状态（🟢 绿色常亮表示在线）。
2. **会话列表项 (ConversationItem)**：
   * 头像：高对比度实色渐变圆角头像，群聊附带群标识；
   * 时间与排版：消息时间应用 `tabular-nums` 防抖；
   * 多态消息预览：智能识别 `[文件] xxx`、`[语音 3s]`、`[图片]`、`[消息已撤回]`，防止裸露 URL；
   * 未读角标：高定强调色平滑小胶囊。
3. **悬浮式底部导航栏 (BottomNav)**：
   * 三大标签：消息（Message）、联系人（Contacts）、设置中心（Settings）；
   * 选中项带有柔和微扩散动效，摒弃生硬宽胶囊。

### 3.3 聊天主舞台 (`ChatScreen`)
1. **高定气泡层次**：
   * **己方气泡**：实色背景（松风墨翠 `#1E5243`），右上角 4dp，其余三角 16dp，搭配纯白文本；
   * **对方气泡**：温润白底卡片（深色模式下为 `#232926`），左上角 4dp，外加 1px 超微弱边缘分割；
   * **撤回消息**：居中轻灰色胶囊提示（`你撤回了一条消息`），克制整洁。
2. **QQ 式文件卡片**：
   * 左侧渲染文件类型图标（PDF、Word、Code、Archive）；
   * 右侧展示文件名、大小以及下载状态（未下载 $\rightarrow$ 下载中加载环 $\rightarrow$ 已下载直接打开）。
3. **悬浮输入坞 (Input Dock)**：
   * 双层微投影输入卡片，附带附件选择抽屉（相册、文件）与发送激活态动效。

### 3.4 设置中心 (`SettingsScreen`)
* 个人资料展示与退出登录；
* **主题与调色板选择器**：
  * 主题模式：白天 (Light) / 暗黑 (Dark) / 跟随系统 (System)；
  * 高定调色板：松风墨翠 / 沧浪深黛 / 黑曜暖琥单选卡片，带实时圆形色块预览。

---

## 4. 架构与依赖规范

* **保持底层协议稳定**：
  * `Frame.kt`：24 字节大端序网络二进制协议；
  * `WsManager.kt`：长连接生命周期；
  * `MimDatabase.kt` / `Daos.kt`：本地 Room 数据库。
* **编译与交付验证**：
  * 构建工具：Gradle Wrapper (`gradlew.bat`)；
  * 产出目标：`clients/android/app/build/outputs/apk/debug/app-debug.apk`；
  * 质量红线：Kotlin 严格类型检查通过，零编译错误，零破坏性变更。
