# Mim (M Instant Message)

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue 3](https://img.shields.io/badge/Vue-3.4+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![Kotlin](https://img.shields.io/badge/Kotlin-1.9+-7F52FF?style=flat&logo=kotlin)](https://kotlinlang.org)
[![gRPC](https://img.shields.io/badge/gRPC-v1.62+-244c5a?style=flat&logo=google)](https://grpc.io)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Mim** 是一款高可靠、注重工程实效、支持双模演进的现代化跨平台即时通讯系统（Instant Messaging）。

项目支持 **轻量单体交付模式** 与 **分布式微服务集群模式** 双模自由切换：
- **轻量单体模式 (`mim-server`)**：单一可执行文件启动，支持嵌入式 SQLite 3 WAL，内存占用 < 30MB，零外部依赖极速启动，适合边缘端与中小团队私有化内网部署；
- **分布式微服务集群模式 (`mim-gateway` + `mim-logic` + `mim-transfer`)**：基于 gRPC + Etcd v3 动态服务注册发现、Redis 7 跨节点路由与原子 SeqID、Kafka 3.x 消息推送削峰、MySQL 8 企业存储及 Prometheus 生产级指标观测，支持支撑海量高并发长连接弹性横向扩容。

前端提供 **Web / Desktop (Tauri v2 + Rust)** 与 **Android (Kotlin + Jetpack Compose 原生)** 三端高质量一致性体验。

---

## 🌟 核心特性与技术亮点

### 1. 架构定位：双模自由演进 (Dual-Mode Architecture)
* **长连接接入网关 (`mim-gateway`)**：纯无状态长连接集群，零数据库直连，维持海量 WebSocket 连接；处理 24 字节协议编解码，通过 gRPC 转发至 Logic 集群，并暴露内部 gRPC 推送端口供反向精准下发。
* **业务逻辑核心 (`mim-logic`)**：无状态 gRPC 服务集群，负责鉴权、消息持久化落库、DFA 敏感词过滤、会话流转与分布式序列号分配，并将推送任务投递至 Kafka 管道。
* **异步推送削峰 (`mim-transfer`)**：消费 Kafka `mim_msg_push` 主题，查阅 Redis 在线路由表，通过 gRPC 将推送帧分发至目标网关节点，实现大群推送与瞬时流量削峰。
* **Etcd v3 动态服务治理 (`pkg/discovery`)**：实例基于租约心跳（Lease + KeepAlive）自动注册至 `/mim/services/{logic,gateway}`，各服务动态感知节点拓扑变动并实现轮询负载均衡。
* **双数据库引擎支持 (`pkg/database`)**：通过 GORM 抽象化适配，一键切换 SQLite 3 (WAL 高性能并发模式) 与 MySQL 8 企业级关系数据库，内置连接池调优（`MaxOpen=100`, `MaxIdle=20`）。

### 2. 消息可靠性与时序机制
* **Redis 7 原子分布式 SeqID**：会话序号统一采用 Redis `INCR mim:conv:{conv_id}:seq` 生成严格单调递增 SeqID，解决多实例并发冲突；支持无 Redis 环境下自动降级为 DB 守卫模式。
* **同类型端互踢（FR-AUTH-03）**：Redis 实时维护多端在线会话表（`mim:user:{uid}:devices`）。同类型端（如 Web 挤 Web）重复登录自动下发 `CmdKickOut` (0x0006) 并阻断旧连接；跨网关实例通过 Redis PubSub 广播剔除。
* **完整 3 段式客户端 ACK 状态机**：
  - **SENT (1)**：服务端入库持久化后即刻返回 `CmdMessageAck` (0x0102)；
  - **DELIVERED (2)**：接收端 WebSocket 收到推送后主动回传 `CmdDeliveryAck` (0x0103)，状态自动同步给发送方；
  - **READ (3)**：接收端进入会话或滚动触发 `CmdReadReceipt` (0x0401)，更新会话已读游标并通知对端。
* **原生 Goroutine 并发安全**：基于 `sync.RWMutex` 与原子状态位管理长连接 Hub，彻底消除向关闭 Channel 写入触发 Panic 的竞态隐患。

### 3. 通信协议：24 字节定长二进制帧头
为兼顾高吞吐与易调试性，Mim 长连接采用 **自定义 24 字节二进制数据包头**：
```text
| Magic (2B) | Version (1B) | Flags (1B) | CmdType (2B) | Reserved (2B) |
|                       Sequence ID (8B)                        |
|      Payload Length (4B)     |      CRC32 Checksum (4B)       |
|                Payload (Protobuf / JSON)                      |
```
* **功能支持**：心跳保活（PING/PONG）、同端挤占互踢（`CmdKickOut`）、单聊与群聊即时收发、消息投递回执（`CmdDeliveryAck`）、已读回执（`CmdReadReceipt`）、增量时序同步（`SyncRequest`）、点对点 WebRTC 音视频信令中继。

### 4. 生产级可观测性 (Prometheus)
* 全服务集成 Prometheus 监控端点 `/metrics`，实时输出标准监控指标：
  - `mim_gateway_online_connections`: 各终端在线 WebSocket 连接数
  - `mim_chat_messages_received_total`: 客户端发送上行消息计数
  - `mim_gateway_messages_pushed_total`: 网关推送下行消息计数
  - `mim_chat_delivery_ack_total`: 送达回执总数 (DELIVERED)
  - `mim_chat_read_receipt_total`: 已读回执总数 (READ)
  - `mim_auth_kicked_users_total`: 同端挤占互踢触发次数
  - `mim_chat_message_processing_duration_seconds`: 端到端消息流转延迟分布

### 5. 安全防护与鉴权防盗链
* **全链路 JWT 鉴权**：统一通过标准 JWT Bearer Header 与 URL Token 校验，彻底拔除调试越权后门。
* **会话隔离与防水平越权 (IDOR)**：单聊/群聊历史记录获取、消息撤回、消息删除与清空均内置会话参与人强校验。
* **附件防盗链**：文件下载基于 HMAC-SHA256 临时时效签名（`?sign=`），杜绝 ID 遍历拖库与路径遍历攻击。
* **敏感数据脱敏**：AI 接入密钥在 API 回显时自动脱敏，且在更新配置时支持掩码防护。

---

## 📁 目录结构

```text
Mim/
├── backend/                  # Go 后端工程根目录
│   ├── cmd/
│   │   ├── server/           # 单体/混合运行入口 (mim-server)
│   │   ├── gateway/          # 微服务：纯长连接接入网关 (mim-gateway)
│   │   ├── logic/            # 微服务：核心业务逻辑集群 (mim-logic)
│   │   ├── transfer/         # 微服务：Kafka 异步推流分发引擎 (mim-transfer)
│   │   ├── bench/            # 高并发压测套件
│   │   └── tools/            # 独立测试工具 (多端同步测试、WebRTC测试)
│   ├── pkg/                  # 基础设施组件
│   │   ├── database/         # 数据库连接池适配器 (SQLite 3 WAL / MySQL 8)
│   │   ├── discovery/        # Etcd 服务心跳注册与动态感知负载均衡
│   │   ├── kafkax/           # Kafka 高性能发布订阅引擎 (含内建零依赖回退队列)
│   │   ├── metrics/          # Prometheus 监控指标度量与 HTTP Handler
│   │   ├── redisx/           # Redis 7 连接池、分布式 SeqID 与在线路由表
│   │   └── jwtx/             # JWT 颁发与验签
│   ├── protocol/             # 24字节自定义二进制帧与 Protobuf/gRPC 契约
│   ├── rpc/                  # gRPC 服务端与客户端连接池复用引擎
│   ├── models/               # 数据模型定义 (GORM / SQLite / MySQL)
│   ├── service/              # 核心业务域
│   │   ├── auth/             # 鉴权与 JWT 生成
│   │   ├── chat/             # 消息会话、Hub 调度与安全通道
│   │   ├── transfer/         # 异步推送路由与网关反向下发服务
│   │   ├── user/             # 用户档案与好友关系链
│   │   ├── group/            # 群聊管理与成员权限
│   │   ├── file/             # 文件去重秒传与签名防盗链
│   │   └── ai/               # AI 智能摘要与对话辅助
│   ├── deploy/               # Docker Compose 生产集群编排文件
│   └── uploads/              # 文件与多媒体物理存储目录
├── clients/
│   ├── desktop-web/          # Web & 桌面客户端 (Vue 3 + Vite + Tailwind + Tauri v2)
│   └── android/              # Android 原生客户端 (Jetpack Compose + Kotlin 协程)
└── scripts/                  # 集群冒烟验证与自动化测试脚本
```

---

## 🚀 快速启动指南

### 1. 运行自动化单元测试 (全绿验证)
```bash
cd backend
go test -count=1 ./...
```

### 2. 方式 A：单体模式极速启动 (Monolith Mode)
适合本地调试、个人体验或内网轻量交付：
```bash
cd backend
# 编译单体服务
go build -o bin/server.exe ./cmd/server

# 启动 (默认使用嵌入式 SQLite WAL 存储，单二进制启动)
./bin/server.exe -port 8080 -db mim.db
```

### 3. 方式 B：分布式微服务集群部署 (Microservices Mode)

#### 方案 1：使用 Docker Compose 一键拉起完整集群
```bash
cd backend/deploy
docker-compose up -d
```
启动包含组件：
* `mysql`: MySQL 8 (端口 3306)
* `redis`: Redis 7 (端口 6379)
* `etcd`: Etcd v3 (端口 2379)
* `kafka`: KRaft 模式 Kafka (端口 9092)
* `mim-logic`: 业务逻辑微服务 (gRPC 端口 9091)
* `mim-gateway`: WebSocket 长连接接入网关 (端口 8081, 内部 gRPC 9092)
* `mim-transfer`: 异步推送流转引擎
* `prometheus`: 指标监控服务 (端口 9090)

#### 方案 2：本地多进程启动
```bash
cd backend
# 编译所有微服务组件
go build -o bin/ ./cmd/...

# 1. 启动业务逻辑服务 (连接 MySQL 8、Redis、Etcd、Kafka)
./bin/logic.exe -port 9091 -db-driver mysql -db "root:root@tcp(127.0.0.1:3306)/mim?charset=utf8mb4&parseTime=True&loc=Local" -redis-addr 127.0.0.1:6379 -etcd-addr 127.0.0.1:2379 -kafka-brokers 127.0.0.1:9092

# 2. 启动异步推送服务
./bin/transfer.exe -kafka-brokers 127.0.0.1:9092 -etcd-addr 127.0.0.1:2379 -redis-addr 127.0.0.1:6379

# 3. 启动接入网关 (可横向扩展多个网关节点)
./bin/gateway.exe -port 8081 -grpc-port 9092 -node-id gateway-1 -etcd-addr 127.0.0.1:2379 -redis-addr 127.0.0.1:6379
```

### 4. 前端启动 (Desktop & Web)

**环境要求**：Node.js 18+，npm。

```bash
cd clients/desktop-web

# 安装依赖
npm install

# 启动 Web 开发服务器 (反向代理自动转接后端)
npm run dev
```

* 构建跨平台桌面客户端 (Tauri v2)：
```bash
npm run tauri build
```

### 5. Android 客户端启动 (Android Studio)
* 使用 Android Studio 打开 `clients/android` 目录。
* JDK 17+ / Gradle 8.6 构建。
* 确保移动设备与后端在同一网络环境，即可直接收发消息、接收送达/已读回执与参与实时音视频通话。

---

## 🔒 预置测试账号 (开发模式)

启动服务后，系统会自动种子初始化 4 个典型角色账号供双向测试：

| 账号 | 初始密码 | 角色说明 |
| :--- | :--- | :--- |
| `alex` | `123456` | 预置测试用户 (带历史单聊与开发组群聊) |
| `bob` | `123456` | 预置测试用户 |
| `carol` | `123456` | 预置测试用户 |
| `david` | `123456` | 预置测试用户 |

---

## 📄 开源许可证

本项目基于 [MIT License](LICENSE) 开源。
