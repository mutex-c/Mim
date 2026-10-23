# Mim (M Instant Message) 架构设计与技术规范规范文档 (Architecture Spec)

---

## 1. 系统整体架构与拓扑

### 1.1 总体架构图

```
                             ┌────────────────────────┐
                             │   三端客户端集群        │
                             │  Web / Windows / Android│
                             └───────────┬────────────┘
                                         │ HTTPS / WSS (Protobuf 二进制帧)
                                         ▼
                             ┌────────────────────────┐
                             │    mim-gateway 网关    │
                             │ (反向代理/鉴权/动态路由)│
                             └───────────┬────────────┘
               ┌─────────────────────────┼─────────────────────────┐
               │ gRPC                    │ gRPC                    │ gRPC
               ▼                         ▼                         ▼
      ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
      │  mim-auth 认证  │       │  mim-user 用户  │       │  mim-chat 对话  │
      │ (JWT/设备互踢)  │       │ (资料/关系链)   │       │(时序Seq/ACK/WS) │
      └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
               │                         │                         │
               ├─────────────────────────┼─────────────────────────┤
               │ gRPC                    │ gRPC                    │ gRPC
               ▼                         ▼                         ▼
      ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
      │ mim-group 群组  │       │  mim-file 文件  │       │mim-settings配置 │
      │ (RBAC/动态禁言) │       │(分块/秒传/防盗) │       │ (系统参数/AI)   │
      └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
               │                         │                         │
               └─────────────────────────┼─────────────────────────┘
                                         │ 异步事件驱动 (Kafka 削峰)
                                         ▼
                                ┌─────────────────┐
                                │  mim-log 日志   │ ◄─── OpenTelemetry (TraceId)
                                │ (审计/IP分析)   │
                                └─────────────────┘
 ─────────────────────────────────────────────────────────────────────────────
 基础设施层 (Middleware):
  • 注册发现: Etcd v3         • 缓存与在线状态: Redis 7 Cluster
  • 关系数据库: MySQL 8.0      • 消息队列: Kafka 3.x
  • 链路追踪: Jaeger          • 指标监控: Prometheus + Grafana
```

---

## 2. 通信协议规范 (Protobuf 二进制传输契约)

为保证网络吞吐量并防止数据膨胀，Mim 彻底废弃普通 JSON 字符串传输，全面采用 **自定义二进制帧头 + Protobuf 负载** 机制。

### 2.1 自定义二进制数据帧结构 (Packet Framing)

每个 WebSocket / TCP 数据包均按以下定长帧头封包：

```
+---------------+---------------+---------------+---------------+
| Magic (2B)    | Version (1B)  | Flags (1B)    | CmdType (2B)  |
+---------------+---------------+---------------+---------------+
|                       Sequence ID (8B)                        |
+---------------+---------------+---------------+---------------+
|                       Payload Length (4B)                     |
+---------------+---------------+---------------+---------------+
|                       CRC32 Checksum (4B)                     |
+---------------+---------------+---------------+---------------+
|                       Payload (Protobuf Data)                 |
|                               ...                             |
+---------------+---------------+---------------+---------------+
```

* **Magic (魔数, 2 Bytes)**: 固定为 `0x4D49` (ASCII: 'M', 'I')，用于快速过滤非法连接与探测。
* **Version (版本号, 1 Byte)**: 协议版本，初始为 `0x01`。
* **Flags (标志位, 1 Byte)**: `Bit 0`: 是否压缩 (Gzip/Snappy), `Bit 1`: 是否端到端加密 (E2EE), 其余保留。
* **CmdType (命令字, 2 Bytes)**:
  * `0x0001`: 心跳 PING / `0x0002`: 心跳 PONG
  * `0x0010`: 认证握手请求 / `0x0011`: 认证握手响应
  * `0x0100`: 单聊消息发送 / `0x0101`: 单聊消息推送 / `0x0102`: 消息投递 ACK
  * `0x0200`: 群聊消息发送 / `0x0201`: 群聊消息广播
  * `0x0300`: 增量消息拉取同步 (Sync Request) / `0x0301`: 同步响应 (Sync Response)
  * `0x0400`: 消息撤回通知 / `0x0401`: 消息已读回执
  * `0x0500`: WebRTC 信令交互 (Offer/Answer/Candidate/Hangup)
* **Sequence ID (8 Bytes)**: 客户端自增序列号，用于全链路请求响应匹配。
* **Payload Length (4 Bytes)**: 负载二进制字节长度（大端序）。
* **CRC32 (4 Bytes)**: 负载校验和，防止网络乱序或数据篡改。
* **Payload**: Google Protobuf 序列化后的真实二进制数据。

### 2.2 核心 Protobuf 协议定义 (`mim.proto`)

```protobuf
syntax = "proto3";
package mim.protocol;
option go_package = "mim/protocol";

// 消息类型枚举
enum MessageType {
  TEXT = 0;
  MARKDOWN = 1;
  IMAGE = 2;
  VOICE = 3;
  VIDEO = 4;
  FILE = 5;
  STICKER = 6;
  REVOKE = 7;
  WEBRTC_SIGNAL = 8;
  SYSTEM_NOTIFICATION = 9;
}

// 会话类型
enum ConversationType {
  SINGLE = 0;
  GROUP = 1;
}

// 核心消息结构
message ChatMessage {
  int64 msg_id = 1;               // 服务端全局唯一 Snowflake ID
  string client_msg_id = 2;       // 客户端本地生成的 UUID，用于幂等去重
  string conversation_id = 3;     // 会话ID (单聊: min_max格式, 群聊: 群ID)
  ConversationType conv_type = 4;
  int64 sender_id = 5;
  int64 receiver_id = 6;          // 单聊为对方ID，群聊为0
  MessageType msg_type = 7;
  string content = 8;             // 文本内容或媒体元数据(JSON/URL)
  
  // 引用与回复
  QuoteInfo quote = 9;
  
  // 消息时序Seq
  int64 seq_id = 10;              // 会话内连续递增序号
  int64 send_time = 11;           // 毫秒时间戳
  
  // 安全与扩展
  bool is_encrypted = 12;         // 是否端到端加密
}

message QuoteInfo {
  int64 quoted_msg_id = 1;
  int64 quoted_seq_id = 2;
  int64 quoted_user_id = 3;
  string quoted_user_name = 4;
  string preview_content = 5;
}

// 客户端投递确认 ACK
message MessageAck {
  string client_msg_id = 1;
  int64 msg_id = 2;
  int64 seq_id = 3;
  int32 code = 4;                 // 0: 成功, 其他: 错误码
  string error_msg = 5;
  int64 server_time = 6;
}

// 增量同步请求 (断网重连/换设备离线拉取)
message SyncRequest {
  string conversation_id = 1;
  int64 last_seq_id = 2;          // 客户端本地已同步的最大 SeqId
  int32 limit = 3;                // 单次拉取数量 (默认 50)
}

message SyncResponse {
  string conversation_id = 1;
  repeated ChatMessage messages = 2;
  bool has_more = 3;
  int64 latest_seq_id = 4;
}

// WebRTC 信令
message WebRTCSignal {
  string call_id = 1;
  int64 from_user_id = 2;
  int64 to_user_id = 3;
  string signal_type = 4;         // OFFER, ANSWER, CANDIDATE, REJECT, HANGUP
  string data = 5;                // SDP 或 Candidate JSON 字符串
}
```

---

## 3. 金融级消息可靠性设计 (Message Reliability Engine)

为彻底解决传统教学项目中“发消息偶发丢失”、“弱网重试导致重复消息”、“断网重连后消息顺序错乱”的死穴，Mim 实现一套工业级消息投递保证机制。

```
 客户端 A (发送端)            mim-chat 服务 (集群)              客户端 B (接收端)
     │                                │                                │
 1.  ├──── 发送 (带 client_msg_id) ───>│                                │
     │                                │ 2. Redis 幂等校验 (防重发)      │
     │                                │ 3. 生成会话自增 SeqId           │
     │                                │ 4. 异步写入 Kafka / MySQL       │
 5.  │<─── 回传 MessageAck (带SeqId) ──┤                                │
     │                                │ 6. 路由查找 B 的在线节点(Redis) │
     │                                ├──── 实时推送 ChatMessage ─────>│ 7.
     │                                │                                │ 8. 校验 Seq 连续性
     │                                │<─── 回传 ReceiveACK ───────────┤ 9.
     │                                │ 10. 标记消息为 DELIVERED       │
```

### 3.1 核心机制与算法
1. **客户端消息幂等性 (Idempotency)**：
   * 客户端每次生成唯一的 `client_msg_id (UUID v4)`。
   * 服务端在 Redis 中维护一个 `SetNX(client_msg_id, 1, ex=600s)` 的滑动去重窗口。
   * 若发生弱网重试，服务端直接返回先前的 ACK 确认，绝不重复写入 MySQL 数据库。
2. **会话单调递增 SeqId**：
   * 使用 Redis `INCR mim:seq:<conversation_id>` 原子递增计数器。
   * 每个会话内的所有消息严格按 SeqId 排序。客户端本地存储维护该会话的最新 `last_seq_id`。
   * 客户端在 UI 渲染时按照 `seq_id` 排序，杜绝本地乱序。
3. **推拉结合的增量离线同步 (Sync Protocol)**：
   * **在线状态**：服务端通过 WebSocket 直接向活跃客户端推送最新消息。
   * **弱网/断线重连**：客户端连接建立后，向服务端发起 `SyncRequest(last_seq_id)`。
   * 服务端执行高效范围查询：`SELECT * FROM mim_messages WHERE conversation_id = ? AND seq_id > ? ORDER BY seq_id ASC LIMIT ?`，成批填补离线消息缺口。
4. **混合读写扩散策略**：
   * **单聊 (1v1)**：采用**写扩散**，收发双方各一份消息投递索引，方便独立执行单向删除与已读计数。
   * **群聊 (Group)**：采用**读扩散**，群消息单份落库，群成员仅维护各自的 `read_seq_id` 游标。避免千人大群因写扩散瞬间产生数千条冗余记录拖垮数据库。

---

## 4. 数据库物理模型设计 (MySQL 8.0 DDL)

采用前缀 `mim_` 规范命名，字符集 `utf8mb4`，时区 `UTC`。

```sql
-- 1. 用户基础表
CREATE TABLE `mim_users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `account` VARCHAR(32) NOT NULL COMMENT '用户唯一账号',
  `nickname` VARCHAR(64) NOT NULL COMMENT '用户昵称',
  `password_hash` VARCHAR(128) NOT NULL COMMENT '密码bcrypt哈希',
  `avatar` VARCHAR(255) DEFAULT '' COMMENT '头像URL',
  `signature` VARCHAR(255) DEFAULT '' COMMENT '个性签名',
  `email` VARCHAR(128) DEFAULT '' COMMENT '绑定邮箱',
  `phone` VARCHAR(32) DEFAULT '' COMMENT '绑定手机',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1正常 2禁言 3封禁',
  `role` TINYINT NOT NULL DEFAULT 1 COMMENT '1普通用户 2管理员',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_account` (`account`),
  KEY `idx_phone` (`phone`),
  KEY `idx_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 2. 用户个性化配置表 (防骚扰/加好友验证)
CREATE TABLE `mim_user_configs` (
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '关联用户ID',
  `search_by_account` TINYINT NOT NULL DEFAULT 1 COMMENT '允许账号搜索: 0禁止 1允许',
  `search_by_nickname` TINYINT NOT NULL DEFAULT 1 COMMENT '允许昵称搜索: 0禁止 1允许',
  `friend_verify_mode` TINYINT NOT NULL DEFAULT 2 COMMENT '1直接添加 2需要验证 3问答验证 4拒绝添加',
  `friend_verify_question` VARCHAR(255) DEFAULT '' COMMENT '验证问题',
  `friend_verify_answer` VARCHAR(255) DEFAULT '' COMMENT '验证答案',
  `recall_time_limit` INT NOT NULL DEFAULT 120 COMMENT '撤回时效限制(秒)',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户偏好配置表';

-- 3. 好友关系链表
CREATE TABLE `mim_friends` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `friend_id` BIGINT UNSIGNED NOT NULL COMMENT '好友ID',
  `remark` VARCHAR(64) DEFAULT '' COMMENT '好友备注',
  `is_pinned` TINYINT NOT NULL DEFAULT 0 COMMENT '是否置顶 0否 1是',
  `is_black` TINYINT NOT NULL DEFAULT 0 COMMENT '是否拉黑 0否 1是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`, `friend_id`),
  KEY `idx_friend_id` (`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

-- 4. 好友申请验证流转表
CREATE TABLE `mim_friend_verifications` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `sender_id` BIGINT UNSIGNED NOT NULL COMMENT '申请人ID',
  `receiver_id` BIGINT UNSIGNED NOT NULL COMMENT '被申请人ID',
  `verify_msg` VARCHAR(255) DEFAULT '' COMMENT '申请留言',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1待处理 2同意 3拒绝 4忽略',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_receiver_status` (`receiver_id`, `status`),
  KEY `idx_sender_receiver` (`sender_id`, `receiver_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友申请流转表';

-- 5. 会话列表状态表
CREATE TABLE `mim_conversations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `conversation_id` VARCHAR(64) NOT NULL COMMENT '会话唯一标识',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '所属用户',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT '单聊目标用户ID或群聊ID',
  `conv_type` TINYINT NOT NULL DEFAULT 0 COMMENT '0单聊 1群聊',
  `last_msg_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '最新一条消息ID',
  `last_msg_preview` VARCHAR(255) DEFAULT '' COMMENT '最新消息预览',
  `last_msg_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `last_read_seq` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '已读最大SeqId',
  `unread_count` INT NOT NULL DEFAULT 0 COMMENT '未读数',
  `is_pinned` TINYINT NOT NULL DEFAULT 0 COMMENT '是否置顶',
  `is_muted` TINYINT NOT NULL DEFAULT 0 COMMENT '免打扰 0否 1是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_conv` (`user_id`, `conversation_id`),
  KEY `idx_user_pinned_time` (`user_id`, `is_pinned`, `last_msg_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话元数据表';

-- 6. 核心消息表 (单聊与群聊统一读写模型)
CREATE TABLE `mim_messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Snowflake 消息ID',
  `client_msg_id` VARCHAR(64) NOT NULL COMMENT '客户端防重UUID',
  `conversation_id` VARCHAR(64) NOT NULL COMMENT '会话ID',
  `conv_type` TINYINT NOT NULL DEFAULT 0 COMMENT '0单聊 1群聊',
  `sender_id` BIGINT UNSIGNED NOT NULL COMMENT '发送方ID',
  `receiver_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接收方ID(单聊有效)',
  `seq_id` BIGINT UNSIGNED NOT NULL COMMENT '会话内连续自增序号',
  `msg_type` TINYINT NOT NULL DEFAULT 0 COMMENT '消息类型 0文本 1MD 2图片 3语音 4视频 5文件 7撤回 8WebRTC',
  `content` TEXT COMMENT '消息文本或媒体结构体JSON',
  `quoted_msg_id` BIGINT UNSIGNED DEFAULT 0 COMMENT '引用的父消息ID',
  `quoted_content` VARCHAR(255) DEFAULT '' COMMENT '引用消息摘要',
  `is_revoked` TINYINT NOT NULL DEFAULT 0 COMMENT '是否撤回 0否 1是',
  `is_encrypted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否端到端加密',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_client_msg_id` (`client_msg_id`),
  UNIQUE KEY `uk_conv_seq` (`conversation_id`, `seq_id`),
  KEY `idx_conv_created` (`conversation_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息存储表';

-- 7. 群组基础表
CREATE TABLE `mim_groups` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '群ID',
  `title` VARCHAR(64) NOT NULL COMMENT '群名称',
  `avatar` VARCHAR(255) DEFAULT '' COMMENT '群头像',
  `notice` TEXT COMMENT '群公告',
  `owner_id` BIGINT UNSIGNED NOT NULL COMMENT '群主ID',
  `member_count` INT NOT NULL DEFAULT 1 COMMENT '当前成员总数',
  `max_members` INT NOT NULL DEFAULT 500 COMMENT '最大成员上限',
  `is_searchable` TINYINT NOT NULL DEFAULT 1 COMMENT '是否允许被搜索',
  `join_mode` TINYINT NOT NULL DEFAULT 2 COMMENT '1自由加入 2需审批 3禁止加入',
  `is_all_muted` TINYINT NOT NULL DEFAULT 0 COMMENT '是否全员禁言 0否 1是',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` DATETIME DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_owner` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群组表';

-- 8. 群成员与权限关系表
CREATE TABLE `mim_group_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL COMMENT '群ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `member_nickname` VARCHAR(64) DEFAULT '' COMMENT '群内昵称',
  `role` TINYINT NOT NULL DEFAULT 3 COMMENT '1群主 2管理员 3普通成员',
  `muted_until` DATETIME DEFAULT NULL COMMENT '禁言到期时间(NULL表示未禁言)',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_group` (`user_id`, `group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员表';

-- 9. 文件与资源元数据表 (秒传与鉴权)
CREATE TABLE `mim_files` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `file_hash` CHAR(64) NOT NULL COMMENT 'SHA-256 哈希值(用于秒传)',
  `file_name` VARCHAR(255) NOT NULL COMMENT '原始文件名',
  `file_size` BIGINT UNSIGNED NOT NULL COMMENT '文件字节大小',
  `file_type` VARCHAR(64) NOT NULL COMMENT 'MIME类型',
  `storage_path` VARCHAR(255) NOT NULL COMMENT '存储物理路径或Key',
  `uploader_id` BIGINT UNSIGNED NOT NULL COMMENT '上传者ID',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_file_hash` (`file_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件元数据表';
```

---

## 5. 三端客户端工程体系设计

```
mim/
├── backend/                  # Go-Zero 微服务后端工程
│   ├── app/
│   │   ├── gateway/          # API 网关
│   │   ├── auth/             # 认证服务 (RPC & API)
│   │   ├── user/             # 用户服务 (RPC & API)
│   │   ├── chat/             # 对话服务 (RPC & API & WebSocket)
│   │   ├── group/            # 群组服务 (RPC & API)
│   │   ├── file/             # 文件服务 (API)
│   │   ├── settings/         # 系统配置服务
│   │   └── log/              # 日志服务 (Kafka Consumer)
│   ├── protocol/             # Protobuf 编解码生成代码
│   └── deploy/               # Docker-Compose, Prometheus, Jaeger 配置
│
├── clients/
│   ├── desktop-web/          # Web 网页端 & Windows 桌面端 (统一工程)
│   │   ├── src/              # 前端业务核心 (Vue 3 / React + Tailwind + Pinia)
│   │   │   ├── components/   # 高质感组件 (聊天流/富文本/录音条/音视频通话)
│   │   │   ├── protocol/     # Protobuf 二进制帧序列化工具
│   │   │   └── stores/       # 客户端状态流
│   │   ├── src-tauri/        # Tauri v2 (Rust 核心)
│   │   │   ├── Cargo.toml    # Rust 依赖 (无任何 Electron 依赖)
│   │   │   └── src/          # 托盘、系统通知、全局快捷键绑定、本地 SQLite
│   │   └── package.json
│   │
│   └── android/              # Android 纯原生工程 (Kotlin + Compose)
│       ├── app/
│       │   ├── src/main/java/com/mim/im/
│       │   │   ├── ui/       # Jetpack Compose UI (120Hz 丝滑流畅)
│       │   │   ├── net/      # OkHttp / Ktor WebSocket + Protobuf 协议处理
│       │   │   ├── service/  # 前台保活 Service 与通知推送
│       │   │   └── rtc/      # WebRTC Native SDK 硬件加速音视频
│       │   └── build.gradle.kts
│       └── settings.gradle.kts
```

---

## 6. 可观测性与运维指标 (APM & Benchmarks)

1. **分布式链路追踪 (OpenTelemetry + Jaeger)**：
   * 网关接收入站请求时，若无 `traceparent` 头则自动生成全局唯一 `TraceId`。
   * 通过 gRPC Client/Server Interceptor 将 `TraceId` 透传至 `auth -> user -> chat -> group` 全链路。
   * Kafka 消息携带 `TraceId`，异步日志消费时无缝串联入库。
2. **Prometheus 指标埋点**：
   * `mim_ws_online_connections`: 当前活跃 WebSocket 连接数（按客户端平台打标签：`web/windows/android`）。
   * `mim_messages_delivered_total`: 消息投递总数计数器（按消息类型区分）。
   * `mim_message_delivery_latency_seconds`: 消息从发送端到接收端端到端时延直方图（重点监控 P95 < 50ms, P99 < 100ms）。
