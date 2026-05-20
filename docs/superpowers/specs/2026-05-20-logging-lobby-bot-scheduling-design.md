# 日志系统 + 大厅系统 + 模拟用户动态调度 设计文档

> **日期**: 2026-05-20
> **范围**: api-server, game-server, bot-service

---

## 一、日志系统

### 1.1 目标
为 api-server (Node.js)、game-server (Go)、bot-service (Go) 统一构建文件日志能力，支持：
- 独立日志文件按服务隔离
- 单文件大小限制自动轮转（100MB）
- 自动备份归档（最多保留 10 个备份文件）
- 日志保存时限管理（最多保留 30 天）
- 旧日志自动压缩（gzip）

### 1.2 架构决策
采用**"服务写文件 + Docker 挂载 + 宿主机 logrotate 兜底"**的轻量方案。

| 服务 | 技术选型 | 日志格式 |
|------|---------|---------|
| api-server | `winston` + `winston-daily-rotate-file` | JSON |
| game-server | `log/slog` + `lumberjack` | JSON |
| bot-service | `log/slog` + `lumberjack` | JSON |

### 1.3 日志路径
容器内统一写入 `/app/logs/service.log`：
```
/app/logs/
  service.log          # 当前活跃日志
  service-2026-05-20.log.gz   # 轮转归档（按日期或大小触发）
```

### 1.4 Docker 挂载
```yaml
# docker-compose.yml 新增
api-server:
  volumes:
    - ./logs/api-server:/app/logs

game-server:
  volumes:
    - ./logs/game-server:/app/logs

simulation-service:
  volumes:
    - ./logs/bot-service:/app/logs
```

宿主机路径：
- `/root/depg/logs/api-server/`
- `/root/depg/logs/game-server/`
- `/root/depg/logs/bot-service/`

### 1.5 宿主机 logrotate 兜底配置
`/etc/logrotate.d/depg`：
```
/root/depg/logs/*/*.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 0644 root root
}
```

### 1.6 日志字段规范（JSON）
```json
{
  "time": "2026-05-20T14:30:00Z",
  "level": "info",
  "msg": "Game started on table sim-table-0",
  "service": "game-server",
  "table_id": "sim-table-0",
  "player_id": "uuid"
}
```

---

## 二、大厅系统 (Lobby)

### 2.1 目标
新增"大厅"概念，用户不再直接上桌，而是先进入大厅查看桌子列表和实时人数，再选择桌子加入。

### 2.2 架构
大厅功能直接扩展在 `game-server` 内，新增 `LobbyManager` 模块：
- **HTTP API**: 获取桌子列表（REST，前端轮询或首次加载）
- **WebSocket**: `/ws/lobby` 实时推送桌子状态变化

### 2.3 大厅 HTTP API
```
GET /lobby/tables
```
响应：
```json
{
  "tables": [
    {
      "id": "sim-table-0",
      "name": "Sim Table 0",
      "max_seats": 7,
      "occupied": 4,
      "small_blind": 5,
      "big_blind": 10,
      "status": "waiting"
    }
  ]
}
```

### 2.4 大厅 WebSocket 协议
连接：`ws://host:8080/ws/lobby`

**Server -> Client 推送：**
```json
{"type": "tables_update", "payload": {"tables": [...]}}
```

**Client -> Server 请求上桌：**
```json
{"type": "join_table", "payload": {"table_id": "sim-table-0", "player_id": "uuid"}}
```

服务器返回上桌 token 或错误，客户端拿到后断开大厅连接，再连接 `/ws?player_id=xxx` 上桌。

### 2.5 桌子状态广播
当任意桌子人数变化时，`TableManager` 通知 `LobbyManager`，`LobbyManager` 向所有大厅连接广播 `tables_update`。

### 2.6 向后兼容
保留现有 `/ws` 端点用于直接上桌（供 bot-service 和旧客户端使用），大厅 WebSocket 走 `/ws/lobby`。

---

## 三、模拟用户动态调度

### 3.1 目标
让模拟用户行为更真实、更灵活，支持与真实用户动态混桌。

### 3.2 核心规则

| 规则 | 说明 |
|------|------|
| **用户池** | 注册 20 个模拟用户，分为"活跃组"和"备用池" |
| **上桌比例** | 每次只派部分用户上桌，保留 5-8 个用户在备用池 |
| **随机局数** | 每个模拟用户上桌时随机设定 `max_hands`（5-20 局），达到后主动 leave |
| **金币补充** | 金币低于最低买入（500）时，通过 API 自动充值回 10000 |
| **真实用户触发** | 真实用户进入大厅等待 10 秒后，若某桌总人数 < 5，从备用池派模拟用户上桌 |
| **人数上限** | 单桌总人数（真实+模拟）>= 5 时不再派新模拟用户 |
| **动态替换** | 模拟用户 leave 后，若桌子总人数 < 5 且有真实用户，从备用池补新用户 |

### 3.3 bot-service 调度器重构

新增 `DynamicScheduler` 替换现有 `Scheduler`：

```go
type DynamicScheduler struct {
    profiles      []SimProfile      // 20 个模拟用户
    activeBots    map[string]*Bot   // 当前上桌的
    standbyBots   map[string]*Bot   // 备用池
    tableAssignments map[string][]string // table_id -> []bot_id
    
    // 配置
    totalPoolSize     int  // 20
    minStandby        int  // 5（最少保留备用数）
    targetTableSize   int  // 5（触发停止增加的目标人数）
    realUserWaitSec   int  // 10（真实用户等待秒数）
    handsRangeMin     int  // 5
    handsRangeMax     int  // 20
    refillThreshold   int  // 500（金币补充阈值）
    refillAmount      int  // 10000（补充后金额）
}
```

### 3.4 调度循环
```
每 5 秒执行一次：
1. 轮询 game-server /lobby/tables 获取所有桌子实时人数
2. 检查每张桌子：
   a. 若总人数 < targetTableSize 且桌子有真实用户：
      - 从 standbyBots 中选一个，设定 random max_hands
      - 派上桌（通过 Manager.AssignToTable）
   b. 若总人数 >= targetTableSize：
      - 不再派新模拟用户
3. 检查每个 activeBot：
   a. 若 handsPlayed >= max_hands：
      - 发送 leave_table 消息，移回 standbyBots
   b. 若 stack < refillThreshold：
      - 调用 API /api/users/refill 充值到 refillAmount
4. 检查是否有真实用户进入大厅 > realUserWaitSec：
   - 是，则优先给该用户所在桌子补模拟用户到 targetTableSize
```

### 3.5 真实用户检测机制
`game-server` 的 `LobbyManager` 在 WebSocket 连接时区分 `isRealUser`（通过 JWT token 或请求参数）。

`bot-service` 通过轮询 `/lobby/tables` 的响应中的 `real_user_count` 字段判断。

### 3.6 模拟用户 leave 流程
1. `bot-service` 的 `GameClient` 收到 `MsgHandResult` 时增加 `handsPlayed` 计数
2. 若达到 `max_hands`，发送 `leave_table` WebSocket 消息
3. `game-server` `HandleLeave` 处理：释放座位、广播 `player_left`、通知 `LobbyManager`
4. `bot-service` 将该 bot 移回 `standbyBots`

### 3.7 金币补充 API
api-server 新增：
```
POST /api/sim/refill
Body: { "user_id": "uuid" }
```
逻辑：检查用户是模拟用户且金币 < 500，则设置为 10000。

---

## 四、数据流图

```
真实用户 -> 前端大厅 -> /ws/lobby (game-server)
                               |
                               v
真实用户选择桌子 -> /ws (game-server) 上桌
                               ^
                               |
bot-service -> 轮询 /lobby/tables -> 动态决策 -> 派模拟用户上桌
     |
     v
模拟用户达到 max_hands / 输完 -> leave_table -> 回归备用池
```

---

## 五、部署变更

### docker-compose.yml 新增
```yaml
volumes:
  api_data:
  logs_api:
    driver: local
  logs_game:
    driver: local
  logs_bot:
    driver: local

api-server:
  volumes:
    - api_data:/app/prisma
    - logs_api:/app/logs

game-server:
  volumes:
    - logs_game:/app/logs

simulation-service:
  volumes:
    - logs_bot:/app/logs
```

### 宿主机目录
```bash
mkdir -p /root/depg/logs/{api-server,game-server,bot-service}
```

---

## 六、验收标准

### 日志系统
- [ ] 三个服务均有独立的 `/app/logs/service.log` 文件
- [ ] 单日志文件超过 100MB 后自动轮转
- [ ] 超过 30 天的旧日志自动清理
- [ ] 宿主机 `logrotate` 配置生效

### 大厅系统
- [ ] `GET /lobby/tables` 返回正确的桌子列表和人数
- [ ] `/ws/lobby` 连接后实时收到桌子人数变化
- [ ] 真实用户可通过大厅选择桌子上桌

### 模拟用户调度
- [ ] 20 个模拟用户中，始终保留至少 5 个在备用池
- [ ] 每个模拟用户上桌时随机设定 5-20 局上限
- [ ] 达到局数上限后自动 leave 并回归备用池
- [ ] 金币低于 500 时自动充值到 10000
- [ ] 真实用户进入大厅 10 秒后，若桌子 < 5 人，自动补模拟用户
- [ ] 桌子总人数 >= 5 时不再增加模拟用户
