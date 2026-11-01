# Mim (M Instant Message) 架构设计与技术规范文档 (Architecture Spec)

---

## 1. 系统整体架构与组件拓扑

### 1.1 双模演化架构设计哲学
Mim 采用 **双模演化架构（Dual-Mode Architecture）**，一套业务代码同时支持极简单体与分布式微服务集群：
1. **轻量单体模式 (`mim-server`)**：
   * 单一二进制文件，集成 SQLite WAL 存储，支持 Redis 路由与本地内存降级模式，适合边缘计算、内网隔离与私有化极简交付（内存占用 < 30MB）。
2. **分布式微服务集群模式 (`mim-gateway` + `mim-logic` + `mim-transfer`)**：
   * **长连接网关层 (`mim-gateway`)**：纯无状态长连接集群，零数据库直连，维持海量 WebSocket 连接；处理 24 字节协议帧解包，通过 gRPC 负载均衡转发至 Logic 集群，并暴露内部 gRPC 推送端口供 Transfer / Logic 反向推送；
   * **业务逻辑层 (`mim-logic`)**：无状态 gRPC 服务集群，负责 JWT 鉴权、消息持久化落库、敏感词过滤、会话流转与分布式序列号分配，并将推送任务投递至 Kafka 管道；
   * **异步推送削峰层 (`mim-transfer`)**：消费 Kafka `mim_msg_push` 主题，查阅 Redis 在线路由表，通过 gRPC 将推送帧分发至目标网关节点，实现大群推送与高并发瞬时流量削峰；
   * **服务治理与注册发现 (`Etcd v3`)**：各微服务实例基于租约心跳（Lease + KeepAlive）自动注册至 `/mim/services/{logic,gateway}`，网关与 Transfer 动态感知节点变动并实现加权/轮询负载均衡；
   * **分布式协同底座 (`Redis 7`)**：维护跨节点多端在线路由表、执行原子 `INCR` 会话递增序列号、并通过 PubSub (`mim:channel:kickout`) 驱动跨节点同端设备互踢（FR-AUTH-03）；
   * **企业级关系数据库 (`MySQL 8`)**：支持事务隔离、连接池优化与分表索引，平滑替代 SQLite 单机写限制；
   * **生产级监控度量 (`Prometheus`)**：网关与服务端统一暴露 `/metrics` 采集端点，涵盖连接数、吞吐量、回执计数与延迟直方图。

### 1.2 系统拓扑图 (Distributed Cluster Mode)

```text
                                  ┌────────────────────────┐
                                  │    三端客户端集群      │
                                  │ Web / Desktop / Android│
                                  └───────────┬────────────┘
                                              │ WSS (24B 定长二进制帧) / HTTPS
                                              ▼
                             ┌─────────────────────────────────┐
                             │       4层 / 7层 负载均衡        │
                             │         (Nginx / L4 LB)         │
                             └──────────┬───────────┬──────────┘
                                        │           │
                    ┌───────────────────┘           └───────────────────┐
                    ▼                                                   ▼
         ┌─────────────────────┐                             ┌─────────────────────┐
         │ mim-gateway (节点 1) │                             │ mim-gateway (节点 N) │
         │  • WebSocket 管理   │                             │  • WebSocket 管理   │
         │  • 24B 协议编解码   │                             │  • 24B 协议编解码   │
         │  • 内部 gRPC 8082   │                             │  • 内部 gRPC 8082   │
         │  • /metrics 监控    │                             │  • /metrics 监控    │
         └──────────┬──────────┘                             └──────────┬──────────┘
                    │                                                   │
                    ├─────────────────┐               ┌─────────────────┤
                    │ gRPC 请求       │ gRPC 请求     │ gRPC 请求       │ gRPC 请求
                    ▼                 ▼               ▼                 ▼
         ┌─────────────────────────────────────────────────────────────────────────┐
         │                         mim-logic 集群 (无状态 gRPC)                    │
         │  • 统一鉴权 (JWT)           • DFA 敏感词过滤       • 会话单调 SeqId 分配 │
         │  • 增量同步 (Sync)          • 3段式 ACK 状态机     • WebRTC 信令中继    │
         └───────────────┬───────────────────┬───────────────────┬─────────────────┘
                         │ 读写数据          │ 租约注册/发现     │ 投递推送任务
                         ▼                   ▼                   ▼
                 ┌───────────────┐   ┌───────────────┐   ┌───────────────┐
                 │ MySQL 8 / DB  │   │    Etcd v3    │   │   Kafka 3.x   │
                 │ (读写连接池)  │   │  (服务注册发现)│   │ (mim_msg_push)│
                 └───────┬───────┘   └───────┬───────┘   └───────┬───────┘
                         │                   │                   │
                         │                   │                   ▼ 批量消费削峰
                         │                   │           ┌───────────────┐
                         │                   │           │ mim-transfer  │
                         │                   │           │ (推送调度集群)│
                         │                   │           └───────┬───────┘
                         │                   │                   │ gRPC PushMessage
                         │                   │                   ▼
                         │           ┌───────┴───────┐   ┌───────────────────────┐
                         └──────────►│    Redis 7    │◄──┤ mim-gateway 节点推送  │
                                     │ • 在线路由表  │   └───────────────────────┘
                                     │ • 原子 INCR   │
                                     │ • 踢人 PubSub │
                                     └───────────────┘
```

---

## 2. 通信协议规范 (Packet Framing & gRPC 契约)

### 2.1 自定义 24 字节定长二进制帧结构

每个长连接数据包均按以下定长 24 字节帧头封包，多字节数值强制采用网络字节序（Big-Endian / 大端序）：

```text
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Magic (2B)           |  Version (1B) |   Flags (1B)  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|         CmdType (2B)          |         Reserved (2B)         | <-- 8字节边界对齐填充
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                                                               |
+                       Sequence ID (8B)                        + <-- 严格落在 Offset 8 边界
|                                                               |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      Payload Length (4B)                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      CRC32 Checksum (4B)                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                     Payload (Protobuf / JSON)                 |
|                              ...                              |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

* **Magic (魔数, 2 Bytes)**: 固定为 `0x4D49` (ASCII: 'M', 'I')，用于快速过滤非法连接与网络扫描。
* **Version (版本号, 1 Byte)**: 协议版本，当前为 `0x01`。
* **Flags (标志位, 1 Byte)**:
  * `0x00`: 标准 Protobuf 二进制负载
  * `0x04`: JSON 文本编码负载 (`FlagJSON`，兼顾前端/多端高可读性与无缝解析)
* **CmdType (命令字, 2 Bytes)**:
  * `0x0001`: 心跳 PING / `0x0002`: 心跳 PONG
  * `0x0006`: 同端挤占互踢通知 (`CmdKickOut`, FR-AUTH-03)
  * `0x0100`: 客户端消息发送 (`CmdChatMessageSend`)
  * `0x0101`: 服务端消息推送 (`CmdChatMessagePush`)
  * `0x0102`: 服务端接收确认 (`CmdMessageAck`, SENT - 状态1)
  * `0x0103`: 客户端送达回执 (`CmdDeliveryAck`, DELIVERED - 状态2)
  * `0x0300`: 增量消息拉取同步 (`CmdSyncReq`)
  * `0x0301`: 同步响应 (`CmdSyncResp`)
  * `0x0400`: 消息撤回通知 (`CmdRevokeNotify`)
  * `0x0401`: 客户端已读回执 (`CmdReadReceipt`, READ - 状态3)
  * `0x0500`: WebRTC 信令中继 (`CmdWebRTCSignal`)
* **Reserved (保留位, 2 Bytes)**: 固定 `0x0000`，确保后续 8 字节 `SeqId` 处于 64 位内存对齐边界。
* **Sequence ID (8 Bytes)**: 会话内单调严格递增序列号（Offset 8 起始）。
* **Payload Length (4 Bytes)**: 负载二进制字节长度。
* **CRC32 (4 Bytes)**: 针对 Payload 计算的标准 IEEE 802.3 校验和，防止网络乱序或数据篡改。

### 2.2 gRPC 微服务契约定义 (`protocol/mim.proto`)

```protobuf
syntax = "proto3";
package mim.protocol;
option go_package = "mim/protocol";

// ---------------- 业务逻辑服务 (mim-logic) ----------------
service LogicService {
  // 发送消息 (网关 -> Logic): 持久化、生成SeqID、进入Kafka推送流水线
  rpc SendMessage(ChatMessage) returns (MessageAck);
  // 增量同步消息 (网关 -> Logic)
  rpc SyncMessages(SyncRequest) returns (SyncResponse);
  // 处理接收方送达ACK (DELIVERED)
  rpc ProcessDeliveryAck(DeliveryAck) returns (MessageAck);
  // 处理接收方已读回执 (READ)
  rpc ProcessReadReceipt(ReadReceipt) returns (MessageAck);
  // 边缘鉴权握手验证
  rpc VerifyToken(AuthHandshake) returns (AuthHandshakeResp);
}

// ---------------- 网关推送服务 (mim-gateway) ----------------
service GatewayService {
  // 推送消息给网关上的指定用户连接 (Transfer/Logic -> 网关)
  rpc PushMessage(GatewayPushReq) returns (GatewayPushResp);
  // 踢除网关上的指定平台设备连接 (跨节点互踢)
  rpc KickDevice(GatewayKickReq) returns (GatewayKickResp);
  // 广播数据帧至多用户
  rpc Broadcast(GatewayBroadcastReq) returns (GatewayBroadcastResp);
}

// ---------------- 状态与回执传输结构 ----------------
message DeliveryAck {
  string client_msg_id = 1;
  string conversation_id = 2;
  uint64 seq_id = 3;
  int64 delivered_time = 4;
}

message ReadReceipt {
  string conversation_id = 1;
  uint64 read_seq_id = 2;
  uint64 user_id = 3;
  int64 read_time = 4;
}

message KickOutNotify {
  uint64 user_id = 1;
  string platform = 2;
  string new_device_id = 3;
  string reason = 4;
  int64 kick_time = 5;
}
```

---

## 3. 并发安全与消息可靠性机制

### 3.1 客户端连接池并发安全 (SafeSend 机制)
在原生 Go 长连接编程中，客户端网络闪断与心跳超时可能导致 `Client.Close()` 关闭底层 Channel。若服务端其他协程此时并发调用 `SendChan <- data`，将引发未捕获的 `panic: send on closed channel` 导致整个进程崩溃。

Mim 在 `Client` 结构体内引入 **双重状态保护 (Atomic Bool + sync.RWMutex)**：
```go
func (c *Client) Send(data []byte) bool {
    if c.closed.Load() {
        return false
    }
    c.sendMu.RLock()
    defer c.sendMu.RUnlock()

    if c.closed.Load() {
        return false
    }

    select {
    case c.SendChan <- data:
        return true
    default:
        // 通道满时的背压与非阻塞丢弃保护
        return false
    }
}
```
* `Close()` 执行时先获取排他写锁 `c.sendMu.Lock()` 并标记 `c.closed.Store(true)`，彻底消除并发关闭与写入的竞争窗口。

### 3.2 客户端幂等去重
* 客户端发送消息携带唯一 `client_msg_id (UUID)`。
* 服务端在插入消息前，首先通过本地唯一约束与查询校验 `client_msg_id`：若已存在，直接返回已生成的 ACK 与对应的 `seq_id`，防止弱网环境客户端自动重发导致消息重复入库。

### 3.3 分布式原子时序 (SeqId)
* 每一会话（单聊或群聊）拥有独立的递增时序。
* **Redis 原子 `INCR`**：键名为 `mim:conv:{conv_id}:seq`。当键不存在时，从数据库执行 `SELECT COALESCE(MAX(seq_id), 0) FROM mim_messages WHERE conversation_id = ?` 进行基线预热，再通过 Redis 事务原子递增。
* **双模降级能力**：在无 Redis 的单体环境下，自动降级为进程内读写锁 + DB 预热递增，确保单节点环境下的单调递增性与快速分配。

### 3.4 完整 3 段式 ACK 状态机
Mim 实现了端到端严格确认机制：
```text
 发送端 A                    服务端 (Logic / Gateway)                  接收端 B
    │                                  │                                  │
    ├─── CmdChatMessageSend (0x0100) ──►│                                  │
    │                                  │ 1. 落库 (status=1 SENT)          │
    │                                  │ 2. 分配 SeqID                    │
    │◄── CmdMessageAck (0x0102) ───────┤ 3. 发布 Kafka 推送任务           │
    │    (状态更新: 已发送)            ├─── CmdChatMessagePush (0x0101) ──►│
    │                                  │                                  │
    │                                  │                                  │ 收到推送帧
    │                                  │◄── CmdDeliveryAck (0x0103) ──────┤
    │                                  │ 4. 状态更新: status=2 DELIVERED  │
    │◄── 转发已送达通知 ───────────────┤                                  │
    │    (状态更新: 已送达)            │                                  │
    │                                  │                                  │ 进入会话/查看
    │                                  │◄── CmdReadReceipt (0x0401) ──────┤
    │                                  │ 5. 状态更新: status=3 READ       │
    │◄── 转发已读通知 ─────────────────┤ 6. 更新已读游标 max_read_seq     │
    │    (状态更新: 已读)              │                                  │
```

### 3.5 同端挤占互踢机制 (FR-AUTH-03)
* 当用户在相同平台（例如同为 `web` 或同为 `android`）建立新连接时：
  * **单机内**：Hub 检出已有的同端连接，下发 `CmdKickOut (0x0006)` 携带 `reason: "Account logged in on another device"` 与 `new_device_id`，随后关闭旧连接。
  * **跨节点集群**：通过 Redis PubSub (`mim:channel:kickout`) 广播踢人事件，各网关节点的订阅协程命中本地 Client 并执行精准驱逐。

### 3.6 Kafka 消息削峰与推送管线
* 业务逻辑服务 `mim-logic` 在接收并持久化消息后，将推送任务打包投递至 Kafka `mim_msg_push` 主题。
* 推送服务 `mim-transfer` 采用 Consumer Group 并行批量消费：
  * 针对单聊：直接查路由表获取接收方网关并调用 gRPC `PushMessage`；
  * 针对大群聊：将群成员按网关节点分组批量分发，避免大群突发消息冲击数据库或导致长连接拥塞。

---

## 4. 全链路安全与防越权架构 (Security Architecture)

### 4.1 统一凭据鉴权 (JWT)
* 全面弃用开发期的 `?user_id=1` 调试后门，HTTP 路由与 WebSocket 握手强制校验 JWT。
* 支持通过 `Authorization: Bearer <token>` 标头或 `?token=<token>` 参数传递有效凭据。
* 密钥支持通过环境变量 `MIM_JWT_SECRET` 注入，杜绝源码硬编码泄露风险。

### 4.2 会话隔离与水平越权防护 (IDOR)
* 对所有涉及会话内容的敏感接口（消息记录拉取、消息撤回、消息物理删除、清空会话）执行前置鉴权：
  * **单聊 (`minID_maxID`)**：严格验证请求发起人 `userID` 是否为会话双方之一。
  * **群聊 (`group_id`)**：严格验证请求发起人是否为 `mim_group_members` 表中的有效成员。
* 删除消息与撤回消息额外校验是否为消息发送本人（群主/管理员拥有群内管理权限）。

### 4.3 静态文件防盗链与路径沙箱
* **防盗链签名**：文件上传后生成 HMAC-SHA256 签名 Token（绑定 `file_id` 与防篡改密钥）。
* **下载鉴权**：文件下载接口接受 `?sign=` 签名或有效用户 JWT，未授权遍历请求均被拦截（返回 401 Unauthorized）。
* **路径沙箱隔离**：对存储物理路径进行 `filepath.Clean` 规范化，并强制验证目标路径处于 `uploads/` 存储沙箱之内，杜绝 `../` 目录穿越风险。

### 4.4 AI 密钥与敏感信息脱敏
* 用户的私有大模型 API Key（OpenAI / DeepSeek 等）在存储后，API 接口读取时自动进行掩码处理（如 `sk-******abcd`）。
* 前端更新配置时若提交包含 `******` 的掩码串，服务端自动保留已有的真实密钥，防止秘钥泄露与误覆盖。

---

## 5. 存储模型与数据库设计 (MySQL 8 / SQLite 3)

系统基于 GORM 提供多数据库引擎抽象支持，由 `db-driver` 配置项控制 (`sqlite` 或 `mysql`)。

### 5.1 核心表结构定义 (MySQL 8 DDL 规范)

```sql
-- 1. 用户基础表 (mim_users)
CREATE TABLE `mim_users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `account` VARCHAR(64) NOT NULL UNIQUE,
  `nickname` VARCHAR(64) NOT NULL DEFAULT '',
  `avatar` VARCHAR(255) NOT NULL DEFAULT '',
  `signature` VARCHAR(255) NOT NULL DEFAULT '',
  `password_hash` VARCHAR(128) NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  INDEX `idx_users_account` (`account`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. 消息存储表 (mim_messages)
CREATE TABLE `mim_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `client_msg_id` VARCHAR(64) NOT NULL UNIQUE,
  `conversation_id` VARCHAR(64) NOT NULL,
  `conv_type` TINYINT NOT NULL,          -- 1: 单聊, 2: 群聊
  `sender_id` BIGINT UNSIGNED NOT NULL,
  `receiver_id` BIGINT UNSIGNED NOT NULL,
  `group_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `msg_type` TINYINT NOT NULL DEFAULT 0,  -- 0: 文本, 1: Markdown, 2: 图片, 3: 语音, 5: 文件
  `content` LONGTEXT NOT NULL,
  `seq_id` BIGINT UNSIGNED NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1,    -- 1: SENT, 2: DELIVERED, 3: READ, 4: REVOKED
  `quoted_msg_id` VARCHAR(64) DEFAULT NULL,
  `quoted_content` VARCHAR(255) DEFAULT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_conv_seq` (`conversation_id`, `seq_id`),
  INDEX `idx_conv_created` (`conversation_id`, `created_at`),
  INDEX `idx_sender` (`sender_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. 会话状态与已读游标表 (mim_conversations)
CREATE TABLE `mim_conversations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL,
  `conversation_id` VARCHAR(64) NOT NULL,
  `conv_type` TINYINT NOT NULL,
  `target_id` BIGINT UNSIGNED NOT NULL,
  `last_msg_id` VARCHAR(64) DEFAULT NULL,
  `last_content` VARCHAR(255) DEFAULT NULL,
  `last_msg_time` DATETIME(3) DEFAULT NULL,
  `last_seq_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `max_read_seq` BIGINT UNSIGNED NOT NULL DEFAULT 0, -- 已读游标
  `unread_count` INT NOT NULL DEFAULT 0,
  `is_pinned` TINYINT NOT NULL DEFAULT 0,
  `is_muted` TINYINT NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_conv` (`user_id`, `conversation_id`),
  INDEX `idx_user_updated` (`user_id`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### 5.2 数据库连接池调优
针对高并发即时通信场景，数据库连接池进行了优化配置：
* `SetMaxOpenConns`: 100（防止瞬时高并发将 MySQL 连接数占满）
* `SetMaxIdleConns`: 20（保持常驻长连接，降低握手开销）
* `SetConnMaxLifetime`: 1 小时（避免 MySQL `wait_timeout` 导致的断连故障）

---

## 6. 生产级可观测性规范 (Prometheus Metrics)

各微服务节点统一通过 `mim/pkg/metrics` 在 `/metrics` 路径对外输出标准 Prometheus 指标：

| 指标名称 | 类型 | 标签 (Labels) | 业务含义 |
| :--- | :--- | :--- | :--- |
| `mim_gateway_online_connections` | Gauge | `platform` (web, android, desktop) | 当前网关活跃 WebSocket 长连接总数 |
| `mim_chat_messages_received_total` | Counter | `conv_type`, `msg_type` | 客户端发送的上行消息计数 |
| `mim_gateway_messages_pushed_total` | Counter | `conv_type` | 网关向下游客户端推送的消息计数 |
| `mim_chat_delivery_ack_total` | Counter | `status` (delivered) | 客户端送达回执总数 (DELIVERED) |
| `mim_chat_read_receipt_total` | Counter | `conv_type` (single, group) | 客户端已读回执总数 (READ) |
| `mim_auth_kicked_users_total` | Counter | `platform` (web, android, desktop) | 同端设备互踢触发次数 |
| `mim_chat_message_processing_duration_seconds`| Histogram| `operation` (send_message, gateway_send_msg)| 消息处理端到端耗时分布（P50, P90, P99） |

---

## 7. 部署架构与启动指南

### 7.1 单体极简模式启动 (Monolith Mode)
适合快速验证、本地测试与边缘设备部署：
```bash
# 1. 编译单体服务
cd backend && go build -o bin/server.exe ./cmd/server

# 2. 启动服务 (默认使用 SQLite WAL 模式与本地降级路由)
./bin/server.exe -port 8080 -db mim.db
```

### 7.2 分布式微服务集群一键部署 (Docker Compose)
在项目根目录或 `backend/deploy/` 下执行：
```bash
cd backend/deploy
docker-compose up -d
```
一键拉起全套组件：
* `mysql`: MySQL 8 关系数据库 (端口 `3306`)
* `redis`: Redis 7 缓存与分布式时序 (端口 `6379`)
* `etcd`: Etcd v3 注册发现中心 (端口 `2379`)
* `kafka`: KRaft 模式消息削峰管道 (端口 `9092`)
* `mim-logic`: 业务逻辑微服务 (gRPC 端口 `9091`)
* `mim-gateway`: WebSocket 接入网关 (WS 端口 `8081`, gRPC 端口 `9092`)
* `mim-transfer`: 异步推送流转服务
* `prometheus`: 指标监控中心 (端口 `9090`)

### 7.3 分布式微服务手动多进程启动
```bash
# 1. 启动业务逻辑服务
./bin/logic.exe -port 9091 -db-driver mysql -db "root:root@tcp(127.0.0.1:3306)/mim?charset=utf8mb4&parseTime=True&loc=Local" -redis-addr 127.0.0.1:6379 -etcd-addr 127.0.0.1:2379 -kafka-brokers 127.0.0.1:9092

# 2. 启动异步推送服务
./bin/transfer.exe -kafka-brokers 127.0.0.1:9092 -etcd-addr 127.0.0.1:2379 -redis-addr 127.0.0.1:6379

# 3. 启动接入网关 (可横向扩展多个实例)
./bin/gateway.exe -port 8081 -grpc-port 9092 -node-id gateway-1 -etcd-addr 127.0.0.1:2379 -redis-addr 127.0.0.1:6379
```
