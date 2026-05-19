# 德州扑克模拟用户系统设计文档

## 1. 背景与目标

当前 game-server、bot-service、api-server 已打通基础联调。本设计目标是构建一个**高保真模拟用户系统**，用于：

- **功能验证**：20 个风格各异的模拟用户持续游戏，暴露系统 Bug
- **压力测试**：每日 1000 局，验证 game-server 在并发场景下的稳定性
- **数据分析**：记录完整操作日志，支持行为分析和异常检测
- **排行榜**：实时统计并展示各类排名数据

## 2. 总体架构

采用**方案 A：扩展现有 bot-service** 为 `simulation-service`。

```
┌────────────────────────────────────────────────────────────┐
│              api-server (Node.js + Prisma)                  │
│  ┌─────────┐ ┌────────────┐ ┌─────────────────────────┐   │
│  │ Auth    │ │ HandHistory│ │ Leaderboard API         │   │
│  │ /users  │ │ /hands     │ │ /sim/leaderboard        │   │
│  └─────────┘ └────────────┘ └─────────────────────────┘   │
└────────────────────────────────────────────────────────────┘
        ^                                ^
        | REST                           | REST
        |                                |
┌───────┴────────────────────────────────┴──────────────────┐
│              simulation-service (Go)                        │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐  │
│  │ User        │ │ Table       │ │ Result              │  │
│  │ Registrar   │ │ Scheduler   │ │ Collector           │  │
│  │ (20人注册)   │ │ (3桌并发)    │ │ (每局结果收集)       │  │
│  └─────────────┘ └─────────────┘ └─────────────────────┘  │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐  │
│  │ Multi-Style │ │ Action      │ │ Anomaly             │  │
│  │ AI Engine   │ │ Logger      │ │ Detector            │  │
│  │ (8种风格)    │ │ (写入DB)     │ │ (胜率/延迟/行为)      │  │
│  └─────────────┘ └─────────────┘ └─────────────────────┘  │
└────────────────────────────────────────────────────────────┘
        | WebSocket                           |
        v                                     v
┌─────────────────────────┐         ┌─────────────────────┐
│   game-server (Go)      │         │  PostgreSQL/SQLite  │
│  Table Manager (3桌x6人) │         │  sim_actions        │
└─────────────────────────┘         │  sim_leaderboard    │
                                    │  sim_anomalies      │
                                    └─────────────────────┘
```

## 3. 数据层设计

### 3.1 Prisma Schema 扩展

```prisma
// === 扩展 User 表（仅新增标记字段）===
model User {
  // ... 现有字段 ...
  isSimUser   Boolean  @default(false) @map("is_sim_user")
  simStyle    String?  @map("sim_style") // tight_aggressive, loose_passive, ...
  simPersonality Json? @map("sim_personality")
}

// === 新增：每局详细操作日志 ===
model SimAction {
  id          String   @id @default(uuid())
  sessionId   String   @map("session_id")
  userId      String   @map("user_id")
  tableId     String   @map("table_id")
  handNumber  Int      @map("hand_number") // 第几手牌
  phase       String   // preflop, flop, turn, river
  action      String   // fold, check, call, bet, raise, all_in
  amount      Int      @default(0)
  potBefore   Int      @map("pot_before")
  potAfter    Int      @map("pot_after")
  stackBefore Int      @map("stack_before")
  stackAfter  Int      @map("stack_after")
  holeCards   String?  @map("hole_cards")   // 脱敏存储，如 "AhKs"
  community   String?  @map("community")    // 公共牌
  timestamp   DateTime @default(now())

  @@index([userId, timestamp])
  @@index([sessionId, timestamp])
  @@index([tableId, timestamp])
  @@map("sim_actions")
}

// === 新增：排行榜快照（每局结束更新）===
model SimLeaderboard {
  id         String   @id @default(uuid())
  userId     String   @map("user_id")
  username   String
  metric     String   // hands_played, hands_won, gold_earned, bb_won, vpip, pfr
  rank       Int
  value      Float
  updatedAt  DateTime @updatedAt @map("updated_at")

  @@unique([metric, userId])
  @@index([metric, rank])
  @@map("sim_leaderboards")
}

// === 新增：异常事件记录 ===
model SimAnomaly {
  id          String   @id @default(uuid())
  type        String   // high_winrate, table_bias, server_lag, bot_stuck
  severity    String   // warning, critical
  description String
  data        Json     // 相关上下文数据
  detectedAt  DateTime @default(now()) @map("detected_at")

  @@index([type, detectedAt])
  @@map("sim_anomalies")
}
```

### 3.2 数据流

1. **游戏前**：simulation-service 调用 `POST /api/auth/register` 注册 20 个用户，获得 JWT token
2. **游戏中**：每手牌的每个 action 通过 WebSocket 接收状态快照，同时写入 `SimAction`
3. **游戏后**：收到 `hand_result` 时，更新 `User` 表金币/统计字段，更新 `SimLeaderboard`
4. **异常时**：触发规则时写入 `SimAnomaly`

## 4. 模拟用户设计

### 4.1 用户档案生成

每个模拟用户包含：

| 字段 | 说明 |
|------|------|
| `username` | 模拟用户名（如 sim_tom、sim_lily） |
| `nickname` | 显示昵称（如 "紧凶 Tom"、"疯鱼 Lily"） |
| `avatar` | 固定头像 URL 或随机分配 |
| `initialGold` | 初始金币（默认 10000） |
| `style` | 游戏风格标签 |
| `personality` | JSON 性格参数 |

### 4.2 8种游戏风格参数

```go
type Persona struct {
    Name       string
    VPIPTarget float64 // 15%-60%
    PFRTarget  float64 // 5%-50%
    Aggression float64 // 0.0-1.0
    BluffRate  float64 // 0.0-1.0
    TiltFactor float64 // 0.0-1.0
    Patience   float64 // 0.0-1.0
}
```

| 风格 | VPIP | PFR | 攻击性 | 诈唬率 | 特点 |
|------|------|-----|--------|--------|------|
| TAG 紧凶 | 18% | 15% | 0.75 | 0.25 | 只玩好牌，激进加注 |
| LAG 松凶 | 32% | 25% | 0.85 | 0.40 | 范围宽，持续施压 |
| NIT 极紧 | 10% | 8% | 0.60 | 0.10 | AA/KK 才入池 |
| LP 松被动 | 35% | 7% | 0.20 | 0.15 | 爱看翻牌，不爱加注 |
| MANIAC 疯鱼 | 55% | 45% | 0.95 | 0.55 | 几乎每手都加注 |
| ROCK 石头 | 12% | 10% | 0.50 | 0.05 | 极少参与，一参就 all-in |
| CALLING_STATION | 45% | 3% | 0.10 | 0.05 | 从不 fold，从不加注 |
| ADAPTIVE 自适应 | 25% | 18% | 0.70 | 0.30 | 根据对手调整策略 |

### 4.3 AI 决策引擎扩展

现有 `ai.Engine` 基于手牌强度做简单决策。新引擎需要：

1. **风格约束**：根据 VPIP/PFR 目标调整入池范围
2. **位置感知**：BTN/SB/BB/UTG 采用不同策略
3. **情绪模拟**：连输后 tiltFactor 升高，决策更激进
4. **对手建模**：记录对手 fold/call/raise 频率，ADAPTIVE 风格据此调整

## 5. 多桌并发调度

### 5.1 调度器设计

```go
type Scheduler struct {
    users      []*SimUser      // 20 个用户
    tables     []*TableSession // 3 张桌
    targetHands int            // 每日目标：1000 局
    handsPlayed int            // 已完成的局数
}
```

**分配策略**：
- 20 人分配到 3 张桌（6 + 7 + 7）
- 每完成一局，该局玩家可选择：留在原位、换桌、或等待下一轮换
- 每完成 20 局，全局重新洗牌分配座位（避免固定对手偏差）

### 5.2 每日 1000 局计算

- 3 桌并发，每桌平均 3 分钟一局（含延迟）
- 每小时约 60 局，17 小时可达 1000 局
- 配置 `SIM_DAILY_HANDS=1000`，达到后自动停止或循环

## 6. 排行榜设计

### 6.1 实时更新机制

每局结束时（收到 `hand_result`）：

```go
func updateLeaderboard(userId string, result HandResult) {
    metrics := []struct{ name string; delta float64 }{
        {"hands_played", 1},
        {"hands_won", result.WinAmount > 0 ? 1 : 0},
        {"gold_earned", float64(result.WinAmount)},
        {"bb_won", result.BBWon},
    }
    // 更新该用户各指标值
    // 20 人全量重排（数据量极小，无需增量算法）
}
```

### 6.2 支持的排行榜维度

| 维度 | 指标 | 说明 |
|------|------|------|
| 最勤劳 | `hands_played` | 参与局数最多 |
| 大赢家 | `hands_won` | 获胜局数最多 |
| 金库王 | `gold_earned` | 累计赢取金币最多 |
| BB猎人 | `bb_won` | 大盲单位净胜最多 |
| 松凶榜 | `vpip` | 入池率最高 |
| 激进榜 | `pfr` | 翻牌前加注率最高 |

### 6.3 查询 API

```
GET /api/sim/leaderboard?metric=hands_won&limit=20
GET /api/sim/users/:id/stats           // 个人综合统计
GET /api/sim/users/:id/history?limit=100 // 最近局数
GET /api/sim/sessions/:id/replay       // 单局完整回放
GET /api/sim/anomalies                 // 异常事件列表
```

## 7. 异常检测

### 7.1 检测规则

| 异常类型 | 规则 | 严重级别 |
|---------|------|---------|
| `high_winrate` | 胜率 > 40% 且局数 > 50 | warning |
| `bot_stuck` | 连续 20 手 fold | warning |
| `server_lag` | action 响应 > 10s | critical |
| `table_bias` | 同一桌 10 局内同一用户赢 > 7 次 | warning |
| `gold_drain` | 金币 < 初始值 × 10% | warning |
| `ws_disconnect` | WebSocket 异常断开 | critical |
| `action_timeout` | 轮到行动但 30s 无响应 | warning |

### 7.2 告警输出

- 写入 `SimAnomaly` 表
- 控制台 ERROR 级别日志
- 可选：推送到告警通道（Webhook）

## 8. 实现范围

### 8.1 需要修改/新增的文件

**bot-service → simulation-service：**
- `cmd/bot-service/main.go` → `cmd/simulation-service/main.go`（入口改造）
- `internal/ai/engine.go`（扩展为8种风格）
- `internal/manager/manager.go`（增加调度器）
- `internal/client/client.go`（增加 action 日志上报）
- `internal/registrar/registrar.go`（新增：用户注册）
- `internal/collector/collector.go`（新增：结果收集）
- `internal/leaderboard/leaderboard.go`（新增：排行榜更新）
- `internal/anomaly/anomaly.go`（新增：异常检测）

**api-server：**
- `prisma/schema.prisma`（新增3张表）
- `src/routes/sim.ts`（新增：排行榜/异常查询路由）

**infra：**
- `docker-compose.yml`（服务名 `bot-service` → `simulation-service`，环境变量扩展）

### 8.2 不在这个版本实现

- Flutter App 查看模拟用户（仅后端 API 支持）
- 模拟用户之间的聊天互动
- 复杂的对手建模（GTO solver 级别）
- 分布式多机部署

## 9. 测试策略

1. **单元测试**：AI 引擎各风格的决策分布符合 VPIP/PFR 目标
2. **集成测试**：20 个用户完成 10 局，验证数据完整性
3. **异常测试**：人为注入高胜率/超时场景，验证异常检测触发

## 10. 部署配置

```yaml
# docker-compose.yml 变更
simulation-service:
  build:
    context: ../apps/simulation-service
    dockerfile: Dockerfile
  environment:
    API_BASE_URL: http://api-server:3000
    GAME_SERVER_WS: ws://game-server:8080/ws
    SIM_USER_COUNT: 20
    SIM_TABLE_COUNT: 3
    SIM_TABLE_SIZE_MIN: 5
    SIM_TABLE_SIZE_MAX: 7
    SIM_DAILY_HANDS: 1000
    SIM_SEAT_ROTATION_INTERVAL: 20
    DB_URL: ${DATABASE_URL}
```

---

**设计确认**：本设计基于方案 A（扩展现有 bot-service），采用混合数据层（复用 User + 专用日志表），支持3桌并发、8种AI风格、实时排行榜和7类异常检测。