# Mim (M Instant Message) 详细分阶段实施计划 (Implementation Plan)

本文档制定了 **Mim 企业级即时通讯平台** 的严格交付步骤、产出物定义与阶段性验证准则。后续开发必须严格依循本计划执行。

---

## 阶段规划概览

```
  Phase 1: 基础设施、数据表模型与 Protobuf 协议契约
    │
    ▼
  Phase 2: Go-Zero 后端基础微服务集群 (Gateway, Auth, User, File, Settings)
    │
    ▼
  Phase 3: 实时即时通讯引擎 (Chat, Group, WebSocket, SeqId, ACK, Kafka, Log)
    │
    ▼
  Phase 4: 现代化 Web 端 & Windows (Tauri v2 + Rust) 客户端实装
    │
    ▼
  Phase 5: Android 纯原生客户端实装 (Kotlin + Jetpack Compose)
    │
    ▼
  Phase 6: WebRTC 音视频跨端互通、可观测性 (APM) 与高并发压测报告
```

---

## 详细实施阶段与任务分解

### Phase 1: 基础设施、数据表模型与 Protobuf 协议契约

* [x] **Task 1.1: 基础设施容器化编排搭建**
  * 编写 `backend/deploy/docker-compose.yaml`，包含 MySQL 8.0、Redis 7、Kafka 3.x + Zookeeper、Etcd v3、Jaeger、Prometheus。
  * 编写 `backend/deploy/env.example` 与 Prometheus 监控配置。
* [x] **Task 1.2: GORM 数据模型与迁移脚本落地**
  * 在 `backend/models` 下实现 9 张核心实体表（`mim_users`, `mim_user_configs`, `mim_friends`, `mim_friend_verifications`, `mim_conversations`, `mim_messages`, `mim_groups`, `mim_group_members`, `mim_files`）。
  * 编写 `backend/models/migrate.go` 自动迁移与 `models_test.go` 完整验证。
* [x] **Task 1.3: Protobuf 二进制协议定义与代码生成**
  * 编写 `backend/protocol/mim.proto`，严格定义帧格式、枚举与消息体（`ChatMessage`, `MessageAck`, `SyncRequest`, `WebRTCSignal` 等）。
  * 执行 `protoc` 生成 Go 语言二进制编解码代码 `mim.pb.go`。
  * 编写定长二进制数据帧头与 CRC32 校验引擎 `backend/protocol/frame.go` 并通过单元测试。

---

### Phase 2: Go-Zero 后端基础微服务集群

* [x] **Task 2.1: `mim-gateway` 网关核心**
  * 实现基于 Go-Zero 的 TraceMiddleware，自动生成并跨链路透传全局唯一 TraceID。
  * 编写 AuthMiddleware，实现正则白名单路由放行、JWT Bearer 凭据解析及上下文用户数据注入。
  * 封装统一响应组件 `backend/pkg/response` 与单元测试。
* [x] **Task 2.2: `mim-auth` 认证服务**
  * 实现账号密码注册与登录，密码采用 bcrypt 强哈希加盐存储。
  * 实现双 Token 机制（AccessToken 短效 + RefreshToken 长效）与无感续签。
  * 编写 `backend/service/auth` 核心服务与单元测试。
* [x] **Task 2.3: `mim-user` 用户与好友服务**
  * 实现个人资料查询、修改与防骚扰搜索控制更新。
  * 实现好友申请状态机流转（申请、同意、拒绝、忽略）以及双向好友关系自动维护。
  * 实现好友列表多维排序、备注名更新与单向/双向好友删除。
* [x] **Task 2.4: `mim-file` 安全文件微服务**
  * 实现客户端文件 SHA-256 计算与秒传（Instant Upload）去重。
  * 实现分片物理落盘存储，以及基于 HMAC-SHA256 的时效性防盗链访问 Token 签名与校验。
* [x] **Task 2.5: `mim-settings` 系统配置服务**
  * 实现系统全局参数动态热更引擎。
  * 基于 DFA（确定有限状态自动机）算法实现高性能敏感词实时过滤与屏蔽。

---

### Phase 3: 实时即时通讯引擎 (核心亮点)

* [x] **Task 3.1: WebSocket 长连接网关与双向通信**
  * 在 `mim-chat` 落地基于 Gorilla WebSocket 的长连接接入层。
  * 实现基于客户端 DeviceID 与 UserID 的多端并发安全连接池 (`Hub`)，支持 Web/Windows/Android 跨端多活。
* [x] **Task 3.2: 会话单调递增 SeqId 与消息可靠投递机制**
  * 实现会话粒度严格单调连续递增的 `SeqId` 引擎。
  * 客户端发送消息持久化后立刻回传携带 `SeqId` 的 `MessageAck`。
  * 严格保证幂等性：基于 `client_msg_id` 防止弱网重复入库。
  * 实现增量拉取接口 `SyncMessages`：通过 `last_seq_id` 差量补齐离线消息，杜绝丢消息与乱序。
* [x] **Task 3.3: 消息撤回、引用与已读回执流转**
  * 实现 2 分钟时效内双向消息撤回，存储层置位并在会话列表动态替换预览内容，广播 `REVOKE` 二进制帧。
  * 引用消息持久化父消息元数据与摘要。
  * 实现 `MarkRead` 批量抹平已读序列号游标。
* [x] **Task 3.4: `mim-group` 企业级群聊微服务**
  * 实现群聊创建、群成员邀请、移交群主、踢人等完整生命周期与 RBAC 权限控制。
  * 实现全员禁言与基于动态倒计时定时器的指定成员禁言。
* [x] **Task 3.5: Kafka 异步削峰与 `mim-log` 审计**
  * 构建非阻塞异步日志生产与多 Worker 消费队列。
  * 实现网络 IP 归属地自动解析与结构化操作审计日志沉淀。

---

### Phase 4: 现代化 Web 端 & Windows (Tauri v2 + Rust) 客户端

* [x] **Task 4.1: 高质感前端设计系统与工程脚手架**
  * 在 `clients/desktop-web` 基于 Vue 3 + TypeScript + Vite + Tailwind CSS 搭建工程。
  * 严格落地 10 条高定铁律，支持白天（Light）、黑夜（Dark）、跟随系统（System）三套主题方案。
  * 彻底去除生硬线条，采用 1%-3% 微明度色阶分层与暗色模式 360 字重光学补偿。
  * 全平台界面坚决杜绝任何 Emoji，统一采用专业 Lucide SVG 矢量图标。
* [x] **Task 4.2: 二进制 WebSocket 通信层与客户端状态流**
  * 编写前端二进制 ArrayBuffer 编解码器（对接 `mim.proto` 与 22 字节定长帧头）。
  * 实现 CRC32 防篡改校验、自动心跳、断线重连与离线增量 Sync 同步。
* [x] **Task 4.3: 核心交互组件库实装**
  * 落地 Telegram / Linear 级三栏式布局（极简导航坞、会话列表、聊天舞台）。
  * 输入框双环机械咬合感微投影、`tabular-nums` 计时器/角标防抖、`text-wrap: pretty` 气泡排版防孤字。
  * 实现 2.5 秒反馈自熄的复制交互（Self-Extinguishing Feedback）。
  * 落地纯 CSS Grid 0fr -> 1fr 丝滑折叠引用条与多媒体/WebRTC 呼叫浮层。
* [x] **Task 4.4: 集成 Tauri v2 (Rust 核心) 打造 Windows 纯原生桌面端**
  * 配置 `src-tauri` 纯 Rust 运行时环境，严格杜绝 Electron。
  * 适配 Windows 系统窗口特性，安装包体控制在 15MB 左右，空闲内存仅约 30-40MB。

---

### Phase 5: Android 纯原生客户端 (Kotlin + Jetpack Compose)

* [ ] **Task 5.1: Android 纯原生工程基建与网络层**
  * 在 `clients/android` 创建纯原生工程（Kotlin + Jetpack Compose）。
  * 基于 OkHttp / Ktor 实现长连接 WebSocket 客户端与 Protobuf 序列化。
* [ ] **Task 5.2: 现代化原生 UI 与触觉反馈**
  * 采用 Material You 规范实现会话列表、联系人列表、聊天界面。
  * 120Hz 丝滑滑动手势、原生系统震动触觉反馈（长按、撤回、发送）。
* [ ] **Task 5.3: 移动端专属特性**
  * 前台保活 Service 维持后台连接，系统状态栏消息通知推送。
  * 本地 Room SQLite 数据库缓存消息，支持弱网秒开即看。

---

### Phase 6: WebRTC 音视频通信、全链路可观测性与压测报告

* [ ] **Task 6.1: 跨端 1v1 WebRTC 音视频实时通话**
  * 完善信令服务器（基于 WebSocket 转发 Offer/Answer/Candidate）。
  * Web / Windows / Android 跨端拉通音视频实时对讲，支持画中画（PIP）与静音切换。
* [ ] **Task 6.2: 全链路可观测性打通 (Jaeger + Prometheus + Grafana)**
  * 各微服务埋点 Prometheus 性能指标（连接数、QPS、P99 时延）。
  * OpenTelemetry 透传 TraceId，在 Jaeger 控制台完整还原一条消息穿越网关、微服务与 Kafka 的调用拓扑。
* [ ] **Task 6.3: 压力测试与基准性能报告输出**
  * 使用 Locust / JMeter 对 WebSocket 网关进行 10,000+ 并发长连接建立与消息吞吐压测。
  * 收集压测数据（吞吐量、CPU/内存占用、端到端时延曲线），输出完整的 Markdown 压测分析报告。
* [ ] **Task 6.4: 简历亮点提炼与项目开源级包装**
  * 输出一份高水准架构文档与简历项目描述模块（直接可复制到个人简历）。
