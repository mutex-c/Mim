# Android Phase 5 移动端核心功能对齐实施计划 (Implementation Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐 Mim Android 移动原生端相比桌面端缺失的核心功能，包括语音录制播放、消息引用回复、群聊全生命周期与三级 RBAC 权限（建群、禁言、任免管理、踢人）、好友申请审批流以及会话置顶/免打扰高级管理。

**Architecture:** 
1. 在 `data/model` 和 `api/ApiService.kt` 补齐对应后端的全套 API 契约（群操作、好友审批、会话状态操作）。
2. 在 `utils/` 实现基于 Android `MediaRecorder` 与 `MediaPlayer` 的音频录制与播放辅助器。
3. 在 `ui/components/` 封装高复用交互抽屉与弹窗：`NewGroupDialog`、`GroupSettingsSheet`、`FriendRequestsSheet`、`AddFriendDialog`、`ConversationActionSheet`。
4. 深度重构 `ChatScreen.kt` 与 `MainScreen.kt`，接入语音条、引用条、禁言感知与置顶排序。
5. 最终通过 Gradle 生产级编译，输出最新 `app-debug.apk`。

**Tech Stack:** 
- Kotlin 1.9+ / Jetpack Compose / Material 3
- Android MediaRecorder / MediaPlayer (AAC/M4A 录制与流式播放)
- Retrofit 2 / OkHttp WebSocket / Room SQLite
- Gradle 8.6 / Android SDK 34

---

### Task 1: 网络契约与 API 接口扩展 (ApiService & Models)

**Files:**
- Modify: `clients/android/app/src/main/java/com/mim/im/data/model/Models.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/api/ApiService.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/repository/ChatRepository.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/repository/AuthRepository.kt`

- [ ] **Step 1: 补充数据模型 `Models.kt`**
  * 增加 `FriendRequestDto`（id, fromUserId, toUserId, fromUserNickname, fromUserAvatar, fromUserAccount, message, status, createdAt）。
  * 增加 `HandleFriendRequestReq`（requestId, status: 1=agree, 2=reject）。
  * 增加 `GroupUpdateReq`、`GroupInviteReq`、`GroupMuteReq`、`GroupKickReq`、`GroupRoleReq`、`GroupTransferReq`。
  * 增加会话操作请求：`PinConversationReq`、`MuteConversationReq`、`DeleteConversationReq`、`ClearMessagesReq`。

- [ ] **Step 2: 在 `ApiService.kt` 声明对应端点**
  * `getFriendRequests`、`handleFriendRequest`、`deleteFriend`。
  * `updateGroup`、`inviteToGroup`、`muteGroupMember`、`kickGroupMember`、`disbandGroup`、`setGroupRole`、`transferGroupOwner`、`quitGroup`。
  * `pinConversation`、`muteConversation`、`deleteConversation`、`clearMessages`。

- [ ] **Step 3: 在 Repository 层封装安全调用方法**
  * 在 `AuthRepository.kt` 提供 `getFriendRequests()`、`handleFriendRequest()`、`sendFriendRequest()`、`deleteFriend()`。
  * 在 `ChatRepository.kt` 提供 `pinConversation()`、`muteConversation()`、`clearMessages()`、`deleteConversation()` 以及群组生命周期方法。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git commit -m "feat(android): add comprehensive API contracts and repository methods for parity"`

---

### Task 2: 真实麦克风录音与流式播放实装 (Voice Notes)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/utils/AudioRecorderHelper.kt`
- Create: `clients/android/app/src/main/java/com/mim/im/utils/AudioPlayerHelper.kt`
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/VoiceRecordButton.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/ChatScreen.kt`

- [ ] **Step 1: 编写 `AudioRecorderHelper.kt`**
  * 封装 `MediaRecorder`，以 `MPEG_4` 容器录制 `AAC` 高清音频（44.1kHz）；
  * 提供 `startRecording(outputFile)`、`stopRecording(): Long`（返回时长毫秒）与 `cancelRecording()`；
  * 提供振幅获取 `getMaxAmplitude()` 用于声波反馈。

- [ ] **Step 2: 编写 `AudioPlayerHelper.kt`**
  * 封装 `MediaPlayer`，支持播放远程 URL 或本地文件；
  * 维护播放中状态与当前进度，自动监听播放完毕事件与资源重置。

- [ ] **Step 3: 编写录音按键组件 `VoiceRecordButton.kt`**
  * 支持点击切换模式（文本输入 / 按住说话）；
  * “按住说话”大胶囊配合手势检测（长按开始录制、上滑显示“松开取消”、松手提交发送）；
  * 录音时浮现精致计时条（如 `00:03s`）与红色呼吸环。

- [ ] **Step 4: 在 `ChatScreen.kt` 实装语音气泡与声波播放条**
  * 消息气泡展示：左侧播放/暂停圆形按钮、中间模拟声波线条、右侧秒数；
  * 点击播放即调用 `AudioPlayerHelper` 流式拉取播放，再次点击停止。

- [ ] **Step 5: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 6: Commit**
  * `git commit -m "feat(android): implement real audio recorder, streaming player and voice bubble"`

---

### Task 3: 消息引用回复体系实装 (Quote / Reply System)

**Files:**
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/ChatScreen.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/repository/ChatRepository.kt`

- [ ] **Step 1: 扩展文本消息发送以支持引用**
  * 允许在 `sendTextMessage` 中附加可选 `quotedMsgId: String?` 与 `quotedSummary: String?`。

- [ ] **Step 2: 在 `ChatScreen.kt` 实现引用横幅 (Quote Banner)**
  * 长按任意消息选择“引用 / 回复”时，记录 `replyingToMessage`；
  * 在输入框上方展示引用横幅（展示原作者名称、原内容缩略预览，带关闭按钮）。

- [ ] **Step 3: 在消息气泡内嵌引用预览框**
  * 若消息携带引用元数据，在气泡正文上方渲染圆角引用小卡片；
  * 点击引用条自动触发 LazyColumn 平滑滚动并高亮该条原消息。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git commit -m "feat(android): add quote-reply banner and in-bubble quote reference rendering"`

---

### Task 4: 好友申请与“新的朋友”审批流实装 (Friend Requests)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/AddFriendDialog.kt`
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/FriendRequestsSheet.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`

- [ ] **Step 1: 编写添加好友弹窗 `AddFriendDialog.kt`**
  * 展示目标用户的头像与账号；
  * 提供留言输入框（预填“我是 [当前用户]”），点击发送调用 `authRepo.sendFriendRequest()`。

- [ ] **Step 2: 编写待审批列表抽屉 `FriendRequestsSheet.kt`**
  * 列表渲染所有待审核的好友申请；
  * 展示申请人头像、昵称、账号与留言；
  * 提供“同意”（主题翠绿）与“拒绝”（轻灰）操作按钮；
  * 点击调用 `authRepo.handleFriendRequest()`，同意后触发会话与联系人列表自动刷新。

- [ ] **Step 3: 在 `MainScreen.kt` 通讯录顶部接入“新的朋友”**
  * 常驻“新的朋友”卡片，展示红色待处理未读角标；
  * 点击呼出 `FriendRequestsSheet`；
  * 在搜索结果中非好友项增加“+ 加好友”按钮并调出 `AddFriendDialog`。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git commit -m "feat(android): implement friend application and approval workflow with pending badge"`

---

### Task 5: 发起建群穿梭框实装 (Create Group Modal)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/NewGroupDialog.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`

- [ ] **Step 1: 编写 `NewGroupDialog.kt`**
  * 好友多选穿梭框：支持搜索过滤、复选框勾选；
  * 顶部水平横滑展示已选中好友头像小胶囊（点击可移除）；
  * 群名称输入框（自动依据勾选成员智能填充默认名）；
  * 点击“立即创建”调用 `chatRepo.createGroup()`，返回新群会话 ID。

- [ ] **Step 2: 在 `MainScreen.kt` 挂载入口**
  * 顶栏右上角增加“发起群聊”图标（`Icons.Default.GroupAdd`）；
  * 建群成功后直接调用 `onOpenChat(...)` 自动进入新群聊天舞台。

- [ ] **Step 3: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 4: Commit**
  * `git commit -m "feat(android): implement multi-select contact shuttle dialog for group creation"`

---

### Task 6: 群管理抽屉与三级 RBAC 权限体系实装 (Group Settings Sheet & RBAC)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/GroupSettingsSheet.kt`
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/MuteDurationDialog.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/ChatScreen.kt`

- [ ] **Step 1: 编写 `MuteDurationDialog.kt`**
  * 提供禁言时长选项：10分钟、1小时、1天、解除禁言。

- [ ] **Step 2: 编写 `GroupSettingsSheet.kt`**
  * 群信息展示：群名称（群主/管理员可修改）、群公告（群主/管理员可修改）；
  * 成员网格与身份徽标：
    * 👑 群主 (Owner, Role 1)；
    * 🛡️ 管理员 (Admin, Role 2)；
    * 普通成员 (Member, Role 3)；
  * 邀请成员入口（“+ 邀请”）；
  * 权限操作菜单：
    * 群主专属：委任/撤销管理员、转让群主、全员禁言开关、解散群聊；
    * 管理员专属：移出普通成员、指定成员限时禁言；
    * 普通成员专属：退出群聊。

- [ ] **Step 3: 在 `ChatScreen.kt` 接入群设置与禁言拦截**
  * 群聊顶部右上角增加群设置按钮；
  * 监听群禁言状态：若当前处于禁言状态，输入框置灰并展示“全员禁言中，仅管理员可发言”提示条。

- [ ] **Step 4: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 5: Commit**
  * `git commit -m "feat(android): implement comprehensive 3-level RBAC group settings and mute controls"`

---

### Task 7: 会话长按高级操作（置顶/免打扰/清空/删除）实装 (Conversation Actions)

**Files:**
- Create: `clients/android/app/src/main/java/com/mim/im/ui/components/ConversationActionSheet.kt`
- Modify: `clients/android/app/src/main/java/com/mim/im/ui/screens/MainScreen.kt`

- [ ] **Step 1: 编写 `ConversationActionSheet.kt`**
  * 抽屉展示：置顶/取消置顶、免打扰/恢复提醒、清空聊天记录、从列表中移除、删除好友（私聊专有）。

- [ ] **Step 2: 在 `MainScreen.kt` 接入长按会话事件**
  * `ConversationRowItem` 支持长按呼出 `ConversationActionSheet`；
  * 置顶会话排序优先级最高，项右上角渲染微倾斜置顶图钉徽标；
  * 免打扰会话展示静音铃铛图标，隐藏高亮红点。

- [ ] **Step 3: 编译检查**
  * 运行 `./gradlew :app:compileDebugKotlin`。

- [ ] **Step 4: Commit**
  * `git commit -m "feat(android): add conversation long-press action sheet with pin, mute, and clear"`

---

### Task 8: 全工程构建验证与 APK 最终产出 (Build & Verification)

**Files:**
- Output: `clients/android/app/build/outputs/apk/debug/app-debug.apk`

- [ ] **Step 1: 执行完整 Gradle 构建**
  * 运行 `cd clients/android && ./gradlew.bat assembleDebug`；
  * 确认状态为 `BUILD SUCCESSFUL`，零 Warning / Error。

- [ ] **Step 2: 验证 APK 产出**
  * 检查产物 `app-debug.apk` 的文件大小与生成时间。

- [ ] **Step 3: 最终 Commit 与交付汇报**
  * `git commit -am "feat(android): complete Phase 5 feature parity with desktop and generate debug APK"`
