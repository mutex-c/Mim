# Mim Android 客户端 Phase 5 功能对齐规格说明书 (Feature Parity Design Spec)

**创建日期**：2026-10-24  
**项目阶段**：Phase 5 (Android 功能全量对齐第一批：群管理 + 语音/引用 + 好友审批 + 会话高级管理)  
**作者/设计**：Antigravity & Mim Architecture Team  
**状态**：Approved (已通过用户方案评审)

---

## 1. 目标与背景

Mim Android 移动原生端现已完成视觉系统与高定调色板的全面重构，整体界面质感已与桌面端高度统一。但在核心交互和社交关系管理功能上，移动端仍缺少桌面端具备的关键能力：
1. **聊天核心**：缺少真实麦克风录音与流式语音播放、缺少长按引用回复及其气泡嵌套展示。
2. **群聊管理**：缺少多选穿梭发起建群、缺少三级 RBAC 权限（群主 👑、管理员 🛡️、普通成员）、缺少全员与单人禁言、缺少成员任免、移出与转让群主。
3. **好友体系**：缺少“新的朋友”待审批列表、缺少主动加好友申请留言与双向同意/拒绝流转。
4. **会话管理**：缺少长按会话的置顶、免打扰、清空记录与删除会话。

本规格书详细定义上述功能的协议契约、UI 架构、状态流转与交付验证方案。

---

## 2. 语音录制/播放与消息引用回复 (Voice & Quote)

### 2.1 真实语音消息流转
1. **录音引擎**：
   * 基于 Android 原生 `android.media.MediaRecorder`，以 `MPEG_4` 容器封装 `AAC` 编码（`audio/mp4` 或 `audio/aac`），采样率 44.1kHz，比特率 64kbps；
   * 临时文件存放于 `context.cacheDir/voice_temp_{timestamp}.m4a`。
2. **交互手感**：
   * 输入栏左侧增加麦克风切换按钮；
   * 点击切换为“按住说话”大胶囊，长按时展示动态计时（如 `00:04s`）与红色呼吸光圈，手指上滑取消发送；
   * 松手自动触发 `chatRepo.sendFileMessage(..., file, "audio/mp4")`，由现有文件服务落地，消息类型为 `MsgType.VOICE`（3）。
3. **播放引擎与气泡展示**：
   * 语音气泡展示：左侧播放/暂停按钮、中间高低起伏的动态声波线、右侧显示时长秒数（如 `6"`）；
   * 基于 `android.media.MediaPlayer`，支持流式拉取 URL 播放，并具备重置与资源安全释放生命周期。

### 2.2 消息引用回复体系
1. **长按交互**：
   * 长按文本/文件/图片/语音消息，在操作菜单中选择 **“回复 / 引用”**；
2. **输入框引用横幅 (Quote Banner)**：
   * 输入框正上方展示优雅的卡片折叠条，显示被引用人昵称与内容摘要，右侧带关闭按钮；
3. **协议与气泡**：
   * 发送的消息在内容结构中携带引用元数据；
   * 消息气泡顶部内嵌浅色缩略条，点击平滑滚动并高亮定位原消息。

---

## 3. 群聊全生命周期与 RBAC 权限体系

### 3.1 发起建群穿梭框 (`NewGroupDialog`)
* **入口**：主页面会话列表或通讯录右上角 “发起群聊”；
* **流程**：多选好友穿梭框 $\rightarrow$ 动态群名生成 $\rightarrow$ 调用 `POST /api/group/create` $\rightarrow$ 自动在本地 Room 建立群会话并跳转群聊舞台。

### 3.2 群管理与成员抽屉 (`GroupSettingsSheet`)
* **入口**：群聊主舞台右上角设置图标；
* **数据契约**：调用 `GET /api/group/info` 与 `GET /api/group/members` 获取实时群详情与成员列表；
* **三级 RBAC 矩阵**：
  * **群主 (Owner, Role 1, 👑)**：
    * 修改群名称、群公告（`POST /api/group/update`）；
    * 委任 / 撤销管理员（`POST /api/group/set-role`）；
    * 转让群主（`POST /api/group/transfer-owner`）；
    * 全员禁言开关 / 单人限时禁言（`POST /api/group/mute-member`）；
    * 移出任意成员（`POST /api/group/kick-member`）；
    * 解散群聊（`POST /api/group/disband`）。
  * **管理员 (Admin, Role 2, 🛡️)**：
    * 修改群公告；
    * 移出普通成员；
    * 禁言普通成员（10分钟 / 1小时 / 1天 / 解禁）。
  * **普通成员 (Member, Role 3)**：
    * 仅查看成员列表与群公告；
    * 主动退出群聊（`POST /api/group/quit`）。
* **禁言感知**：
  * 若群处于全员禁言或当前用户被单独禁言，禁用发送输入框并显示黄色警告条。

---

## 4. 好友申请审批流与会话高级管理

### 4.1 好友申请与审批流
1. **主动加好友**：
   * 搜索陌生用户，点击“加好友”，弹出申请对话框填写验证信息；
   * 调用 `POST /api/user/add-friend`。
2. **“新的朋友”待审批列表**：
   * 通讯录顶部常驻待处理卡片与红色未读数徽标；
   * 点击展开申请列表（`GET /api/user/friend-requests`）；
   * 支持点击“同意”与“拒绝”（`POST /api/user/handle-friend-request`）；
   * 同意后自动入库并生成与对方的私聊会话。

### 4.2 会话长按高级管理
* 长按会话弹出底部抽屉操作项：
  * **置顶 / 取消置顶**：`POST /api/chat/pin-conversation`；
  * **免打扰 / 取消免打扰**：`POST /api/chat/mute-conversation`；
  * **清空记录**：`POST /api/chat/clear-messages`；
  * **从列表中移除**：`POST /api/chat/delete-conversation`；
  * **删除好友**：`POST /api/user/delete-friend`。

---

## 5. 架构与依赖规范

* **保持底层协议与架构稳定**：
  * 继续复用 `ApiService.kt`、`ChatRepository.kt` 与 `AuthRepository.kt`；
  * 新增 API 路由方法均严格映射 Go 后端已有路由；
* **测试与编译验证**：
  * 执行 Gradle 编译校验 `./gradlew assembleDebug`；
  * 确保零编译错误，生成可运行的更新版 APK。
