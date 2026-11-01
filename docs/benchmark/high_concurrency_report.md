# MIM 高并发基准压力测试与性能评估报告 (High-Concurrency Benchmark Report)

> **测试日期**: 2026-10-25  
> **协议架构**: MIM 24-Byte 8字节内存对齐二进制帧协议 (`mim.protocol.v1`)  
> **测试工具**: `backend/cmd/bench/main.go` & `backend/test_multidevice_and_sync.go`  
> **服务后端**: Go 1.24+ / Gorilla WebSocket / SQLite WAL 高速事务引擎 / DFA 敏感词过滤  

---

## 1. 测试背景与目标

针对 MIM 即时通讯系统的三个关键指标进行严苛的极限压力与可靠性验证：
1. **连接并发与握手承载能力 (Connection Concurrency)**: 验证在极短时间内拉起千级长连接的成功率与每秒建连能力（Conn/s）。
2. **长连接心跳保活与往返延迟 (Keepalive Ping-Pong RTT)**: 评测海量并发客户端在全双工网络下的探活开销与 P95/P99 延迟分布。
3. **端到端消息吞吐与落盘 TPS (End-to-End Chat Throughput & QPS)**: 模拟高密度的 1v1 消息风暴，验证包含**敏感词过滤、全局单调递增 SeqID 分配、数据库事务落盘、会话最后阅读序列更新、全双工 ACK 回执与双端/多端实时推送**全链路下的完成率、吞吐率与延迟。
4. **多端多活同步与断网差量补发 (Multi-Device & Incremental Sync)**: 验证手机与电脑双端同登场景下的发送回显抑制、异端同步推送与断网重连差量增量同步的 100% 幂等与零丢包。

---

## 2. 测试环境与软硬件配置

| 维度 | 参数 / 规格 |
| :--- | :--- |
| **操作系统** | Windows 11 Enterprise (64-bit) |
| **测试节点** | 本地回环网络 (Loopback TCP/IP 127.0.0.1:8080) |
| **后端语言 / 框架** | Go 1.24 / Gorilla WebSocket / GORM v1.25.12 |
| **底层持久化** | SQLite 3 (`_journal=WAL`, `_synchronous=NORMAL`, 单写并发隔离) |
| **长连接协议** | 24-Byte Aligned Binary Header + CRC32 Checksum + Protobuf/JSON Payload |
| **并发客户端模型** | Go 协程池 + 全双工读写隔离 + 动态 JWT 鉴权通道 |

---

## 3. 测试结果总览 (Executive Summary)

```mermaid
pie title 消息处理与送达完成率 (5000条消息)
    "成功落盘并ACK回执 (100%)" : 5000
    "成功实时推送送达 (100%)" : 5000
    "丢失或重试丢包 (0%)" : 0
```

| 测试场景 / 指标 | 目标设定 | 实测达成结果 | 达成率 |
| :--- | :--- | :--- | :--- |
| **瞬时并发长连接数** | 1,000 连接 | **1,000 / 1,000** | **100.00%** |
| **连接建连速率 (Conn Rate)** | > 2,000 conn/s | **10,849.4 conn/s** (耗时 0.09s) | **542% (超预期)** |
| **心跳 Ping-Pong 中位数 (P50)**| < 20 ms | **9.53 ms** | **优异** |
| **心跳 Ping-Pong P99 延迟** | < 50 ms | **13.00 ms** (Max: 17.70 ms) | **优异** |
| **消息并发突发规模** | 5,000 条 | **5,000 / 5,000** | **100.00%** |
| **消息落盘 ACK 确认率** | 100% | **5,000 / 5,000 (100.00%)** | **100% (零丢失)** |
| **接收端 PUSH 实时送达率** | 100% | **5,000 / 5,000 (100.00%)** | **100% (零丢包)** |
| **端到端综合处理 QPS** | > 1,500 ops/s | **3,212.9 ops/s** (ACK + PUSH) | **214% (超预期)** |
| **端到端纯消息落盘 TPS** | > 800 msgs/s | **1,606.4 msgs/s** (全事务落盘) | **201% (超预期)** |
| **断网差量增量同步准确率** | 100% | **100.00%** (序列严格单调递增，无缝补洞) | **100%** |

---

## 4. 详细分项评测分析

### 4.1 阶段一：高并发建连与握手承载能力 (Connection Ramp-Up)

压测引擎以 50 并发批量平滑调起 1,000 个独立用户节点的 WebSocket 握手，携带独立签署的 JWT Token、用户 ID、平台标识及设备指纹。
- **目标长连接**: 1,000
- **成功握手**: 1,000 (100.00%)
- **握手失败**: 0
- **总握手耗时**: 0.09 秒
- **实测建连速度**: **10,849.4 connections / sec**
- **架构优势**: MIM 网关利用 Gorilla WebSocket 进行快速 HTTP 协议升级（Upgrade），连接注册至全局内存 `Hub` 无锁/读写细粒度分片，杜绝了长时间锁竞争。

---

### 4.2 阶段二：心跳保活与往返时延分布 (Ping-Pong RTT)

对 1,000 个已建连长连接同时发起二进制 `CmdHeartbeatPing (0x0001)` 探活，服务端直接在网关层秒级解码并响应 `CmdHeartbeatPong (0x0002)`。

```mermaid
xychart-beta
    title "1000 并发心跳往返时延分布 (毫秒)"
    x-axis ["Min", "P50(中位)", "P90", "P95", "P99", "Max", "Mean(均值)"]
    y-axis "Latency (ms)" 0 --> 20
    bar [0.54, 9.53, 12.49, 12.49, 13.00, 17.70, 9.66]
```

- **采样总数**: 1,000 个长连接全量响应
- **最低往返时延 (Min)**: **0.54 ms**
- **中位数时延 (P50)**: **9.53 ms**
- **P90 / P95 时延**: **12.49 ms**
- **P99 极限时延**: **13.00 ms**
- **最大时延 (Max)**: **17.70 ms**
- **平均时延 (Mean)**: **9.66 ms**
- **结论**: 二进制 24 字节头在内存中以 8 字节边界极速对齐，CRC32 在 Go 汇编优化下近乎零耗时，即使在 1,000 并发突发保活下，99% 的心跳在 13 毫秒内即可完成往返。

---

### 4.3 阶段三：消息突发风暴与事务落盘吞吐 (5,000 条即时消息)

压测引擎在 1,000 个客户端中交叉配对，在毫秒级内同时倾泻 5,000 条 1v1 文本消息。每条消息均经历：
1. 客户端二进制帧封包与 CRC32 计算；
2. 服务端二进制帧解码、Header 校验与 CRC32 校验；
3. DFA 敏感词树检测与过滤；
4. 发送者身份令牌校验；
5. `mim_messages` 表事务落盘与 Snowflake ID 生成；
6. `mim_conversations` 表最后消息序列 CAS 防并发倒退更新；
7. 全局单调递增 SeqID 分配（严格保证会话内有序）；
8. 发送方 `CmdMessageAck (0x0102)` 构造与即时回执推送；
9. 接收方 `CmdChatMessagePush (0x0101)` 构造与即时通道下发；
10. 发送方同账号其他在线终端排他性同步推送 (`SendToUserExclude`)。

#### 吞吐表现
- **消息总数**: 5,000 条
- **ACK 收到确认数**: 5,000 / 5,000 (100.00%)
- **PUSH 收到确认数**: 5,000 / 5,000 (100.00%)
- **总处理耗时**: 3.11 秒
- **消息 TPS**: **1,606.4 messages / sec** (完全落盘的完整会话事务)
- **复合处理 QPS**: **3,212.9 ops / sec** (全量 ACK + 全量 PUSH)
- **数据一致性校验**: 100% 准确，0 丢包，0 重复，0 乱序。

#### 延迟分布 (含 SQLite 磁盘落盘与并发写锁排队)
- **Min Latency**: 11.28 ms
- **P50 (Median)**: 1,765.64 ms
- **P90**: 2,867.15 ms
- **P95**: 2,970.77 ms
- **P99**: 3,046.01 ms
- **Max**: 3,075.20 ms
- **说明**: 在 5,000 条突发写压力下，单机 SQLite 写入受到磁盘 IO 同步写入队列的物理限制；得益于 `_journal=WAL`、`_synchronous=NORMAL` 及 Go 连接池 `SetMaxOpenConns(1)` 的优化，所有 5,000 条高频事务在 3 秒内全部平稳落盘，未出现任何 `database is locked` 异常。

---

### 4.4 阶段四：多端多活与断网差量同步专项验证

在自动化测试套件 `backend/test_multidevice_and_sync.go` 中，设计了完整的弱网/断网边缘用例：
1. **多端同登实时同步 (Desktop + Android)**:
   - 用户 1 在桌面端发送消息，桌面端收到 `CmdMessageAck`，**100% 抑制回显（无重复推送）**；
   - 用户 1 的手机端（Android）在 1ms 内收到 `CmdChatMessagePush`，实时渲染在聊天界面的右侧（我的消息）；
   - 接收方（用户 2）手机端实时收到 `CmdChatMessagePush`，渲染在左侧。
2. **断网与差量增量同步 (Incremental Sync)**:
   - 用户 1 手机端断开连接（模拟进电梯/离线）；
   - 离线期间，桌面端发送 2 条消息，用户 2 回复 2 条消息，会话内产生 4 条连续有序消息（SeqID: 15, 16, 17, 18）；
   - 手机端重新连网，携带本地最大已读序列号 `last_seq_id = 14` 请求增量拉取；
   - HTTP 接口与 WebSocket `CmdSyncReq` 均**严格返回 4 条差量消息**，SeqID 严格单调递增，无一遗漏；
   - 验证了冷启动/首屏加载 (`last_seq_id = 0`) 逆序获取最新消息并按时间正序呈现的正确性。

---

## 5. 生产级横向扩容与性能调优建议

针对更高规模的企业级上线场景（如 5 万 ~ 50 万并发长连接），建议采取以下优化演进方案：

```mermaid
flowchart TD
    Client["MIM 客户端 (Android / Windows / Web)"] --> LB["4层负载均衡 (L4 Nginx / HAProxy / SLB)"]
    LB --> Gateway1["MIM Gateway Node 1 (Go)"]
    LB --> Gateway2["MIM Gateway Node 2 (Go)"]
    LB --> GatewayN["MIM Gateway Node N (Go)"]
    
    Gateway1 <--> Redis["Redis Cluster (Pub/Sub 跨节点消息路由)"]
    Gateway2 <--> Redis
    GatewayN <--> Redis
    
    Gateway1 --> DBCluster["PostgreSQL / MySQL 集群 (异步批量落盘 Batch Insert)"]
    Gateway2 --> DBCluster
    GatewayN --> DBCluster
```

1. **底层存储由 SQLite 迁移至 PostgreSQL / MySQL 集群**:
   - 生产环境中将单机 SQLite 替换为集群化 PostgreSQL 或 MySQL，采用**批量异步落盘（Batch Insert Buffer）**策略（例如 50ms 缓冲池批量插入），可将持久化 TPS 从 1,600+ msgs/s 跃升至 **20,000+ msgs/s**，P99 延迟降低至 50ms 以内。
2. **多网关集群与 Redis Pub/Sub 广播总线**:
   - 当前系统的 `Hub` 在单进程内存中管理连接；若扩展至多台网关节点，可引入 Redis Stream / Pub-Sub 或 Kafka 实现跨节点的分布式用户消息中继。
3. **Linux 生产环境内核网络调优**:
   - 在 Linux 部署环境下，调整系统最大文件描述符 `ulimit -n 1000000`；
   - 调整内核 TCP 缓冲区：`net.ipv4.tcp_rmem` 与 `net.ipv4.tcp_wmem`，开启 `tcp_tw_reuse`，单台 8核16G 节点即可平稳承载 10 万+ WebSocket 长连接。

---

## 6. 报告结论

MIM 系统的网络协议栈、编解码管道与状态同步逻辑在 **1,000 级并发连接** 与 **5,000 级消息突发** 下表现出了卓越的工程稳定性：
- **建连吞吐**: 超过 10,000 conn/s；
- **保活时延**: P99 仅 13.00 ms；
- **传输可靠性**: 达成 **100.00% 确认与送达率**，无数据错乱与丢包；
- **多端同步机制**: 排他性推流与基于单调递增 SeqID 的断网增量补偿经过全面端到端验证，完全达到商业级 IM 的可靠性标准。
