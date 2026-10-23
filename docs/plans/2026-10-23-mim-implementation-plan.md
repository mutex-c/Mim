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

* [ ] **Task 1.1: 基础设施容器化编排搭建**
  * 编写 `backend/deploy/docker-compose.yaml`，包含 MySQL 8.0、Redis 7、Kafka 3.x + Zookeeper、Etcd v3、Jaeger、Prometheus。
  * 验证所有容器健康启动并网络互通。
* [ ] **Task 1.2: GORM 数据模型与迁移脚本落地**
  * 在 `backend/models` 下实现 9 张核心实体表（`mim_users`, `mim_user_configs`, `mim_friends`, `mim_friend_verifications`, `mim_conversations`, `mim_messages`, `mim_groups`, `mim_group_members`, `mim_files`）。
  * 编写自动迁移与预置初始管理员数据脚本。
* [ ] **Task 1.3: Protobuf 二进制协议定义与代码生成**
  * 编写 `backend/protocol/mim.proto`，严格定义帧格式、枚举与消息体（`ChatMessage`, `MessageAck`, `SyncRequest`, `WebRTCSignal` 等）。
  * 配置并执行 `protoc`，生成 Go 语言与前端 TypeScript 对应的数据绑定代码。

---

### Phase 2: Go-Zero 后端基础微服务集群

* [ ] **Task 2.1: `mim-gateway` 网关核心**
  * 基于 Go-Zero API 框架搭建网关，配置基于 Etcd 的动态微服务路由与反向代理。
  * 集成 JWT 校验中间件、请求白名单正则过滤、全链路 TraceId 注入与透传。
* [ ] **Task 2.2: `mim-auth` 认证服务**
  * 实现账号密码注册登录、密码加盐存储（bcrypt）。
  * 颁发双 Token（AccessToken + RefreshToken）与无感续期。
  * Redis 记录设备会话状态，实现“同端互踢、三端共存”下线广播。
* [ ] **Task 2.3: `mim-user` 用户与好友服务**
  * 实现个人资料查询与修改，防骚扰配置更新。
  * 好友申请流转（申请、同意、拒绝、忽略）、好友列表分页查询与备注名修改。
  * 实现 RPC 接口供其他微服务查询用户简要信息与在线状态。
* [ ] **Task 2.4: `mim-file` 安全文件微服务**
  * 实现分块上传、SHA-256 秒传判重逻辑。
  * 实现图片本地安全裁剪与基于有时效 Token 的防盗链文件下载预览。
* [ ] **Task 2.5: `mim-settings` 系统配置服务**
  * 系统全局参数动态读取与热更。

---

### Phase 3: 实时即时通讯引擎 (核心亮点)

* [ ] **Task 3.1: WebSocket 长连接网关与双向通信**
  * 在 `mim-chat` 搭建基于 Gorilla WebSocket 的长连接接入层。
  * 实现自定义二进制帧解析（Magic 校验、CRC32 校验、Protobuf 解码）。
  * 建立客户端连接池（Goroutine 安全分片 Map），集成心跳保活与超时断开。
* [ ] **Task 3.2: 会话单调递增 SeqId 与消息可靠投递机制**
  * 基于 Redis `INCR` 实现每个会话独立的连续单调递增 `seq_id`。
  * 客户端发送消息，服务端落地并立刻回传带 `seq_id` 的 `MessageAck`。
  * 实现离线增量拉取接口 `SyncMessages`：按 `last_seq_id` 差量拉取最新数据。
* [ ] **Task 3.3: 消息撤回、引用与已读回执流转**
  * 消息 2 分钟时效内撤回，双向广播 `REVOKE` 指令并在存储层打标记。
  * 引用消息持久化父消息元数据，前端反向定位支持。
* [ ] **Task 3.4: `mim-group` 企业级群聊微服务**
  * 建群、加群审核、邀请入群、踢人、移交群主、解散群逻辑。
  * 全员禁言与指定成员动态倒计时禁言（Redis TTL 动态刷新）。
  * 群消息广播引擎（基于读扩散模式避免写雪崩）。
* [ ] **Task 3.5: Kafka 异步削峰与 `mim-log` 审计**
  * gRPC 拦截器与网关拦截器将请求行为打入 Kafka。
  * `mim-log` 消费 Kafka 队列，异步解析真实 IP 归属地并记录结构化操作日志。

---

### Phase 4: 现代化 Web 端 & Windows (Tauri v2 + Rust) 客户端

* [ ] **Task 4.1: 高质感前端设计系统与工程脚手架**
  * 在 `clients/desktop-web` 基于 Vue 3 / React + TypeScript + Tailwind CSS 搭建工程。
  * 定制深色/浅色高质感轻奢主题（毛玻璃、精细阴影、平滑转场动画）。
* [ ] **Task 4.2: 二进制 WebSocket 通信层与客户端状态流**
  * 编写前端二进制 ArrayBuffer 编解码器（对接 `mim.proto`）。
  * 实现前端自动心跳保活、指数退避重连、断线自动发起 `SyncRequest` 增量同步。
* [ ] **Task 4.3: 核心交互组件库实装**
  * 仿 Telegram / Discord 级会话列表、吸顶时间戳、精美气泡组件。
  * 多媒体消息组件：多图预览、HTML5 音频录音波形条、视频播放器、文件下载卡片。
  * 消息引用联动高亮滚动、撤回消息重编、右键原生风味上下文菜单。
* [ ] **Task 4.4: 集成 Tauri v2 (Rust 核心) 打造 Windows 纯原生桌面端**
  * 配置 `src-tauri` Rust 运行时环境（纯 Rust 核心，绝对不引入 Electron）。
  * 编写 Rust 原生插件：系统托盘右键菜单、新消息托盘呼吸灯闪烁、全局唤起快捷键（`Ctrl+Alt+W`）、开机自启配置。
  * 验证打包生成纯原生 Windows 安装包（体积控制在 15MB 以内，空闲内存占用 < 40MB）。

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
