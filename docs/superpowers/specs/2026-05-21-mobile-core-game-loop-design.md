# 移动端核心游戏循环设计规范

> **目标:** 设计并实现 Flutter 移动端应用的核心游戏循环，支持横竖屏双模式、10人桌完整显示、中文为主多语言。

**架构:** 采用 BLoC 状态管理模式，每屏一个独立 BLoC，通过 Repository 层与后端 REST API 和 WebSocket 通信。

**技术栈:** Flutter, flutter_bloc, intl (i18n), WebSocket, shared_preferences

---

## 1. 导航架构

### 1.1 底部导航 (4 Tabs)

```
┌─────────────────────────────────────┐
│  ♠  大厅                    💰 2,450 │  ← 顶部栏: 图标 + 金币
├─────────────────────────────────────┤
│                                     │
│         [ 内容区域 ]                 │
│                                     │
├─────────────────────────────────────┤
│  🏠 大厅   💬 社交   🏆 排行   👤 我的  │  ← 底部导航
└─────────────────────────────────────┘
```

| Tab | 标签 | 功能 |
|-----|------|------|
| 1 | 大厅 | 桌子列表、筛选、加入游戏 |
| 2 | 社交 | 消息、好友、站内消息、聊天室 |
| 3 | 排行 | 排行榜 |
| 4 | 我的 | 个人中心、设置 |

### 1.2 顶部栏

- 每个页面左上角显示 ♠ 图标（品牌标识）
- 右上角显示金币余额 + 头像入口

---

## 2. BLoC 架构

### 2.1 架构分层

```
┌─────────────────────────────────────────┐
│           UI Layer (Screens)            │
│  LobbyScreen / TableScreen / ...        │
├─────────────────────────────────────────┤
│           BLoC Layer                    │
│  AuthBloc / LobbyBloc / TableBloc / ... │
├─────────────────────────────────────────┤
│           Repository Layer              │
│  AuthRepository / LobbyRepository       │
│  GameRepository (WebSocket)             │
├─────────────────────────────────────────┤
│           Data Source                   │
│  REST API (api-server:3000)             │
│  WebSocket (game-server:8080)           │
└─────────────────────────────────────────┘
```

### 2.2 TableBloc 状态机

```
TableInitial → TableConnecting → TableConnected → TableJoining
                    ↓                                          ↓
         TableDisconnected ←─ Any network error         TableJoined
                                                                ↓
                    TableWaiting → TableDealing → TableBetting
                         ↑                           ↓          ↓
                    TableDisconnected ←─ WebSocket close   TableShowdown
                                                                ↓
                                                           TableWaiting (next hand)
```

### 2.3 关键状态定义

| 状态 | 说明 |
|------|------|
| `TableConnecting` | 正在连接 WebSocket |
| `TableConnected` | WebSocket 已连接，等待加入桌子 |
| `TableJoined` | 已加入桌子，收到完整桌子状态 |
| `TableBetting` | 下注轮次，显示操作按钮 |
| `TableShowdown` | 摊牌阶段 |
| `TableDisconnected` | 连接断开，自动重连中 |
| `TableError` | 错误状态，显示重试按钮 |

---

## 3. 大厅界面 (Lobby Screen)

### 3.1 筛选标签

横向滚动标签栏：
- **现金桌** (默认选中)
- **SNG**
- **锦标赛**
- **训练赛**

### 3.2 桌子卡片

紧凑卡片设计，每屏显示 4-6 张：

```
┌──────────────────────────────┐
│ 经典六人桌        [● 在线]    │
│ 5/10 · 上限 5000    5/6 人   │
└──────────────────────────────┘
```

- 左侧：桌子名称 + 盲注级别 + 上限
- 右侧：人数/容量 + 在线状态指示点

### 3.3 桌子状态颜色

| 状态 | 颜色 | 说明 |
|------|------|------|
| 有空位 | #00b4d8 (青色) | 可加入 |
| 即将满员 | #90e0ef (浅青) | 剩1-2座 |
| 满员 | #778da9 (灰色) | 不可加入 |

---

## 4. 牌桌界面 (Table Screen) — 核心

### 4.1 设计原则

基于业界成熟扑克APP（参考德扑之星）设计：
- **纯色绿色绒布桌面**，无木质边框
- **方形头像框**（圆角矩形），非圆形
- **玩家均匀分布**在桌面四周
- **圆形操作按钮**，底部居中
- **底池信息**位于中上部，社区牌位于中间

### 4.2 竖屏布局 (Portrait 375×812)

```
┌─────────────────────────────┐
│ 03:51          💰 2,450  71% │  ← 状态栏
├─────────────────────────────┤
│                             │
│      [柒少]  [庄家D]        │  ← 顶部玩家
│         239.5BB             │
│                             │
│  [空座]  底池:5.3BB  [静牌] │  ← 底池信息在上
│                             │
│   [薄注]  [A♥][K♠][Q♣] [超哥]│  ← 中间玩家 + 社区牌
│                             │
│   [脆皮]              [南山]│  ← 底部侧玩家
│                             │
│        [我] K♥3♦            │  ← Hero 底部中央
│        119.8BB              │
│                             │
│    [✕弃牌] [+] [2BB跟分]    │  ← 圆形操作按钮
│       50%     67%           │  ← 快捷下注比例
├─────────────────────────────┤
│ ⬜ ⏱              💬 4      │  ← 底部工具栏
└─────────────────────────────┘
```

#### 玩家分布（10人桌）

玩家均匀分布在椭圆形桌面边缘，使用参数化定位：
- 椭圆中心：50%, 50%
- 竖屏椭圆：a=35%, b=38%（纵向椭圆）
- 从底部中央（Hero/BTN）开始，逆时针每 36° 一个座位

| 位置 | 角度 | 对应角色 |
|------|------|----------|
| 底部中央 | 90° | Hero (BTN) |
| 右下 | 54° | SB |
| 右中 | 18° | BB |
| 右上 | 342° | UTG |
| 上右 | 306° | UTG+1 |
| 顶部中央 | 270° | MP (Dealer) |
| 上左 | 234° | MP+1 |
| 左上 | 198° | MP+2 |
| 左中 | 162° | HJ |
| 左下 | 126° | CO |

#### 玩家卡片样式

```
    昵称
  ┌────┐
  │ 🧔 │  ← 32×32px 方形头像，圆角4px
  └────┘
  239.5BB  ← 筹码数（金色）
```

- 头像外框：深色半透明背景 + 细边框
- 当前行动玩家：头像框 **金色边框** 高亮 + 底部倒计时圆点
- 弃牌玩家：显示 **"弃牌"** 灰色标签，头像变暗
- 庄家：**黄色 "D"** 圆点标记

#### 社区牌

- 尺寸：28×40px
- 真实扑克牌布局：左上角点数 + 中央花色 + 右下角旋转点数
- 白色渐变背景 + 阴影
- 牌背：蓝色渐变 + 虚线边框 + ♠ 标志

#### 底池信息

- 位置：**桌面中上部**（top: 20%），靠近顶部玩家
- 样式：深色圆角胶囊 + 金色文字
- 显示："底池: 5.3BB"
- 玩家筹码堆：底池上方横向排列

#### Hero 区域（底部中央）

- 昵称在上方（青色高亮）
- 方形头像 + **青色边框**（表示当前行动）
- 头像上方显示倒计时（"11S"）
- 手牌显示在头像右侧：30×44px 真实扑克牌
- 手牌下方显示牌型（"高牌"）

#### 操作按钮（底部）

三个圆形大按钮：

| 按钮 | 颜色 | 图标 | 文字 |
|------|------|------|------|
| 弃牌 | 红色渐变 | ✕ | 弃牌 |
| 加分/加注 | 蓝色渐变 | + | 精准加分 |
| 跟注 | 绿色渐变 | 2BB | 跟分 |

快捷下注比例标签（加注按钮上下）：
- 50% 底池
- 67% 底池

### 4.3 横屏布局 (Landscape 812×375)

- 桌面改为**横向椭圆**：a=38%, b=26%
- 玩家分布：4顶部 + 2上侧 + 2下侧 + 2底部（含Hero）
- 头像可稍大（36×36px）
- 操作按钮保持同样圆形风格，横向排列
- 社区牌可稍大（30×44px）

### 4.4 空座显示

```
┌────┐
│空座│  ← 灰色虚线边框，半透明背景
└────┘
```

### 4.5 状态标签

| 标签 | 样式 | 触发条件 |
|------|------|----------|
| 弃牌 | 灰色胶囊 | 玩家弃牌后 |
| Straddle | 红色胶囊 | 玩家下 Straddle |
| ALL IN | 红色粗体 | 玩家 All-in |
| 暂离 | 虚线边框 | 玩家离席 |

---

## 5. 数据流与 WebSocket 协议

### 5.1 连接生命周期

1. `TableBloc` 收到 `TableConnect(wsURL)` → 打开 WebSocket
2. 连接成功 → `TableConnected` → 发送 `join_table {table_id, token}`
3. 服务器返回完整 `table_state` → `TableJoined(state)`
4. 游戏中 → 服务器推送 `game_event` → `TableBloc` 映射为状态

### 5.2 服务器 → 客户端事件映射

| 服务器事件 | BLoC Action | UI 状态 |
|------------|-------------|---------|
| `table_state` | `TableGameStateUpdated` | 更新桌子状态 |
| `your_turn {timeout}` | `TablePlayerTurnStarted` | 高亮玩家 + 显示操作栏 |
| `player_action` | `TablePlayerActionReceived` | 筹码动画 |
| `community_cards` | `TableCommunityCardsDealt` | 翻牌动画 |
| `showdown` | `TableShowdownStarted` | 摊牌 + 高亮赢家 |
| `pot_won` | `TablePotWon` | 筹码飞向赢家 |
| `player_left` | `TablePlayerLeft` | 移除座位 |
| `chat_message` | `TableChatReceived` | 追加聊天 |

### 5.3 客户端 → 服务器操作

- `action {type: fold|check|call|raise, amount?}` — 玩家决策
- `chat {text}` — 牌桌聊天
- `leave_table` — 离开桌子

### 5.4 方向切换状态保持

- `TableBloc` 位于 `TableScreen` 上方的 Widget 树中
- 方向改变时仅 UI 重建，BLoC 状态保持不变
- 不需要重新连接

---

## 6. 错误处理策略

### 6.1 连接错误

| 场景 | 行为 |
|------|------|
| 连接超时 (10s) | `TableError("连接超时")` → 显示重试按钮 |
| 游戏中断线 | `TableDisconnected` → 自动重连（退避：1s, 2s, 4s, max 30s） |
| 重连成功 | 重新加入桌子，服务器同步完整状态 |

### 6.2 游戏错误

| 场景 | 行为 |
|------|------|
| 操作被拒绝 | `TableActionRejected` → Toast 提示，重新启用操作栏 |
| 操作超时 | 服务器自动弃牌 → `TablePlayerFolded` (auto) |

### 6.3 边缘情况

| 场景 | 处理 |
|------|------|
| App 进入后台 | 保持 WebSocket（30s 心跳），被杀后恢复时重连 |
| 来电中断 | 同后台处理 |
| 筹码不足跟注 | 操作按钮禁用，提示 "筹码不足" |

---

## 7. 测试策略

### 7.1 单元测试 — BLoC 状态转换

```dart
blocTest('join table flows through correct states',
  build: () => TableBloc(mockRepo),
  act: (bloc) => bloc.add(TableConnect('ws://test')),
  expect: () => [
    isA<TableConnecting>(),
    isA<TableConnected>(),
    isA<TableJoined>(),
  ],
);
```

### 7.2 Widget 测试 — 响应式布局

- 竖屏渲染网格布局
- 横屏渲染椭圆布局
- 方向切换保持状态

### 7.3 集成测试 — 完整流程

登录 → 大厅 → 加入桌子 → 玩一手 → 离开

---

## 8. 配色方案

| 用途 | 色值 |
|------|------|
| 背景 | #0d1b2a |
| 桌面 | #1a5c3a (radial gradient to #1e7a48) |
| 表面/卡片 | #1b263b |
| 边框 | #415a77 |
|  muted 文字 | #778da9 |
| 强调色 | #00b4d8 |
| 金色（筹码） | #ffd700 |
| 红色（弃牌） | #e74c3c |
| 蓝色（加注） | #3498db |
| 绿色（跟注） | #2ecc71 |

---

## 9. 文件结构

```
lib/
├── l10n/
│   ├── app_zh.arb          # 中文主语言
│   └── app_en.arb          # 英文 fallback
├── blocs/
│   ├── auth_bloc.dart
│   ├── lobby_bloc.dart
│   └── table_bloc.dart
├── repositories/
│   ├── auth_repository.dart
│   ├── lobby_repository.dart
│   └── game_repository.dart
├── screens/
│   ├── lobby_screen.dart
│   ├── table_screen.dart
│   └── ...
├── widgets/
│   ├── player_card.dart
│   ├── poker_card.dart
│   ├── action_buttons.dart
│   └── ...
└── main.dart
```
