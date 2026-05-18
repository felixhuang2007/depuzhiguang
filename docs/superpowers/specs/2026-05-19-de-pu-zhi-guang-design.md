# 德扑之光 (De Pu Zhi Guang) — Product Design Specification

**Product:** 德扑之光 — Myanmar-focused Online Texas Hold'em Poker Platform  
**Date:** 2026-05-19  
**Status:** Approved for implementation  
**Target Market:** Myanmar (primary Chinese-speaking community, with Burmese localization)  
**Platforms:** iOS & Android native mobile apps  

---

## 1. Executive Summary

德扑之光 is a full-featured No-Limit Texas Hold'em poker platform designed for the Myanmar market. It supports four core game modes (Cash Games, Tournaments, Sit & Go, Spin & Go), a robust club-based social system, and a simulated-user (bot) module to ensure table liquidity during early-stage growth.

The platform uses a dual-currency model (Gold + BB) with dynamic exchange rates in tournament modes. Payment is integrated with KBZPay (Myanmar's dominant mobile wallet). The architecture is optimized for Myanmar's variable mobile network conditions.

**Key non-functional requirements:**
- 500+ concurrent simulated users (bots)
- Sub-200ms action latency end-to-end
- Graceful degradation on 3G / high-packet-loss networks
- Multi-language: Simplified Chinese (primary), English, Burmese
- Virtual currency first, architecture ready for real-money expansion

---

## 2. System Architecture

### 2.1 High-Level Diagram

```
┌─────────────┐     ┌─────────────┐     ┌─────────────────┐
│  Flutter    │◄───►│  API Gateway │◄───►│  Node.js API    │
│  Mobile App │     │  (Nginx/     │     │  Server         │
│  (iOS/Android)    │   Kong)      │     │  (Users, Clubs, │
└──────┬──────┘     └─────────────┘     │  Payments,      │
       │                                  │  Leaderboards)  │
       │ WebSocket                        └────────┬────────┘
       │                                           │
       ▼                                           ▼
┌─────────────┐                          ┌─────────────────┐
│  Go Game    │◄────────────────────────►│  PostgreSQL     │
│  Server     │   gRPC / HTTP            │  (Primary DB)   │
│  (Real-time │                          └─────────────────┘
│   poker     │◄────────────────────────►│  Redis          │
│   engine)   │   Pub/Sub                │  (Cache,        │
└──────┬──────┘                          │   Sessions,     │
       │                                 │   Leaderboards) │
       ▼                                 └─────────────────┘
┌─────────────┐
│  Bot Service│
│  (Go)       │
│  500+ sim   │
│  users      │
└─────────────┘
```

### 2.2 Component Responsibilities

| Component | Language | Responsibility |
|---|---|---|
| **Flutter App** | Dart | UI rendering, game animation, local state caching, WebSocket client, offline caching |
| **API Gateway** | Nginx/Kong | SSL termination, rate limiting, JWT validation, routing, WAF rules |
| **Node.js API** | TypeScript | RESTful APIs — user auth, profiles, clubs, payments (KBZPay), leaderboards, admin dashboard |
| **Go Game Server** | Go | WebSocket connections, real-time game state, hand evaluation, pot calculation, side-pot logic, timer management, blind level advancement |
| **Bot Service** | Go | Simulated user lifecycle, poker AI decision engine, table allocation, humanization behaviors |
| **PostgreSQL** | SQL | Persistent data — users, clubs, transactions, hand histories, tournament records, audit logs |
| **Redis** | — | Hot data — active sessions, real-time leaderboards, table state cache, rate limits, pub/sub |

### 2.3 Key Architectural Decisions

1. **Go Game Server owns all table state in-memory.** It never queries PostgreSQL during a hand. State is backed up to Redis every 5 seconds for crash recovery. This guarantees sub-100ms decision latency.
2. **Node.js API owns all persistent business logic.** It never touches active table state. Game server and API server communicate via gRPC only for account balance checks at sit-down / cash-out.
3. **Bot Service connects to the Game Server exactly like a real player** (via internal gRPC using the same protocol). Bots are first-class citizens — no special-case code in the game engine.
4. **Dual WebSocket connections** per client: one latency-critical game socket, one tolerant social socket. Prevents chat spikes from interfering with hand timing.

---

## 3. Gameplay & Game Modes

### 3.1 Core Rules: No-Limit Texas Hold'em

- **Deck:** 52 cards, shuffled with Fisher-Yates, seeded from CSPRNG
- **Positions:** Dealer, Small Blind, Big Blind, UTG, UTG+1, MP, MP+1, Hijack, Cutoff
- **Betting rounds:** Preflop → Flop → Turn → River
- **Actions:** Fold, Check, Call, Bet, Raise, All-in
- **Showdown:** Best 5-card hand wins. Side pots calculated automatically for all-in players.
- **Hand evaluation:** Fast 5-card evaluator using lookup tables. Rank hands in <1μs.

### 3.2 Game Mode 1: Cash Games

| Parameter | Options |
|---|---|
| Table size | 2 (heads-up), 6 (short), 9 (full ring) |
| Stakes | Bronze (1/2), Silver (5/10), Gold (25/50), Platinum (100/200), Diamond (500/1K) |
| Buy-in | 20–200 BB |
| Rake | 3–5% of pot, capped at 3 BB (configurable per table) |
| Straddle | Optional UTG straddle (2 BB) |
| Run it twice | Optional when all-in preflop (high stakes only) |

Players join/leave anytime. Stack credited to account instantly on leave. **BB→Gold rate is fixed per table level.**

### 3.3 Game Mode 2: Tournaments

| Parameter | Options |
|---|---|
| Format | Freezeout, Rebuy, Re-entry, Bounty |
| Speed | Slow (15-min levels), Normal (10-min), Turbo (5-min), Hyper (3-min) |
| Payout structure | Top 10–15% standard, customizable for private clubs |
| Late registration | Up to level 6 (configurable) |

**Tournament state machine:** Registration → Running (blind levels advance automatically) → Final Table → Completed.

**Dynamic BB→Gold exchange rates apply** — see Section 8.

### 3.4 Game Mode 3: Sit & Go (SNG)

Single-table or multi-table. Starts when all seats fill.

| Type | Players | Payout |
|---|---|---|
| STT | 6 or 9 | Top 2 (6-max) / Top 3 (9-max) |
| MTT SNG | 18, 45, 90, 180 | Standard tournament payout |
| Heads-up SNG | 2 | Winner takes all |

### 3.5 Game Mode 4: Spin & Go

3-player, hyper-turbo, winner-takes-all format with random prize multiplier.

| Buy-in | Multiplier range | Expected ROI |
|---|---|---|
| Low tier | 2x–10,000x | ~95% (5% rake built in) |
| High tier | 2x–25,000x | ~95% |

The multiplier is drawn from a weighted distribution before the first hand. Prize pool displayed prominently during play.

### 3.6 Game Flow State Machine

```
WaitingForPlayers → DealingHoleCards → PreflopBetting → DealingFlop
→ FlopBetting → DealingTurn → TurnBetting → DealingRiver
→ RiverBetting → Showdown/AllFold → HandComplete → WaitingForPlayers
```

**Timer rules:**
- Player action timer: 15 seconds standard, 10 seconds for Spin & Go
- Time bank: Players accumulate +10s per 10 hands played (max 60s)
- Disconnection grace: 60 seconds. Auto-fold if timer expires. Auto-check if no bet to call.

---

## 4. Social System

### 4.1 Friends

- **Add friend:** By username, player ID, or QR code scan
- **Friend status:** Online / In-Game / Offline / Last seen
- **Interaction:** Invite to table, invite to club, view profile stats, private message
- **Buddy list limit:** 500 friends (soft cap)

### 4.2 Clubs (Private Rooms)

The core monetization and retention engine. Players join clubs managed by agents/hosts.

| Feature | Description |
|---|---|
| **Club creation** | Any player can create a club (first free, subsequent costs Gold) |
| **Membership** | Public, Approval-required, or Invite-only |
| **Roles** | Owner → Manager → Agent → Member |
| **Custom tables** | Club owners create tables with custom stakes, rake, and game modes |
| **Club chip system** | Separate virtual currency within clubs. Owners buy bulk chips from platform; members buy from owners via external payment |
| **Hand history** | Club owners can review all hands played in their club |
| **Club leaderboard** | Weekly/monthly rankings by winnings, hands played, VPIP |

### 4.3 Chat

| Type | Features |
|---|---|
| **Table chat** | Text + preset emoji reactions. Optional translation (Chinese ↔ Burmese). Mute per player. |
| **Private chat** | 1-on-1 messaging between friends |
| **Club chat** | Group chat for all club members |
| **Moderation** | Keyword filtering, player mute, report function |

**Anti-spam:** New accounts cannot send URLs. Rate limit: 1 msg/3s at tables.

### 4.4 Observer Mode (Railbird)

- Any non-player can observe cash games or final tables
- Observer sees all community cards and betting action
- Observer **does not** see hole cards until showdown
- Observer count displayed to players
- Separate "spectator chat" channel

### 4.5 Hand Replay & Sharing

- **Replay viewer:** Step through the hand street by street
- **Share:** Generate shareable image (cards + board + action summary) or replay link
- **Save:** Players bookmark hands to their profile
- **Club feed:** Notable hands auto-posted to club activity feed

---

## 5. Simulated User (Bot) Module

### 5.1 Architecture

```
Bot Service (Go)
├── Bot Manager
│   ├── Table Allocator (assign bots to tables needing players)
│   ├── Lifecycle Controller (spawn / pause / retire bots)
│   └── Behavior Scheduler (human-like delays, breaks)
├── AI Engine
│   ├── Hand Evaluator (same as game server)
│   ├── Decision Model (action selection)
│   └── Risk Profile (aggression, bluff frequency)
└── Identity Generator
    ├── Name generator (Chinese, Myanmar, English names)
    ├── Avatar assignment
    └── Fake stats with realistic variance
```

The Bot Service connects to the **Go Game Server via gRPC** using the same protocol as real players. The game server cannot distinguish a bot from a human without checking a hidden flag.

### 5.2 Bot Difficulty Levels

| Level | % of Bots | Description | Use Case |
|---|---|---|---|
| **Fish** | 30% | Loose-passive. Calls too much, folds to aggression. | Makes new players feel skilled |
| **Regular** | 50% | TAG style. Solid fundamentals. Small mistakes. | Realistic practice opponent |
| **Shark** | 15% | Aggressive, position-aware, balanced ranges. | Challenge for improving players |
| **Whale** | 5% | Loose-aggressive, unpredictable. Big swings. | Entertainment value |

### 5.3 Decision Model

Each bot evaluates using a simplified poker AI:

1. **Hand strength:** Calculate equity vs. random range (preflop) or vs. perceived range (postflop)
2. **Pot odds:** Required equity = call amount / (pot + call amount)
3. **Position adjustment:** Tighter early, looser late. More bluffs in position.
4. **Stack depth:** Adjust implied odds for deep stacks
5. **Randomization:** Controlled noise so bots don't play identically

**Example preflop logic (Regular bot):**
- UTG: Raise top 12%, fold rest
- MP: Raise top 18%
- BTN: Raise top 35%, call limps top 25%
- BB: Defend top 25% vs. raises

### 5.4 Humanization

| Behavior | Implementation |
|---|---|
| **Timing** | Decision delays 2–8 seconds (Fish faster, Shark slower on tough decisions) |
| **Chat** | Bots never chat. Occasional preset emoji use only. |
| **Session length** | 30–120 minutes, then sit out for 10–30 minutes |
| **Table switching** | Occasionally leave and join new tables |
| **Tilt simulation** | After bad beat, play slightly worse for 10–15 hands |

### 5.5 Scaling Target

| Metric | Value |
|---|---|
| Concurrent bots | 500+ |
| Tables managed | 50–100 |
| Decision latency | <500ms |
| CPU per bot | ~0.1 core |
| Bot service instances | 2–3 (redundancy) |

### 5.6 Transparency

When real money is eventually introduced, bots must be clearly labeled or removed. Architecture supports a config flag `bots_enabled`. When disabled, bots finish current hands and leave. No bot identities are reused for real players.

---

## 6. Security, Anti-Cheating & Compliance

### 6.1 Foundation (Launch)

| Layer | Measures |
|---|---|
| **Auth** | JWT access tokens (15-min expiry) + refresh tokens. Passwords bcrypt-hashed. Optional 2FA for club owners. |
| **Transport** | TLS 1.3 for all connections. WSS for WebSocket. |
| **API Security** | Rate limiting (100 req/min per IP, 1000 req/min per user). Input validation. Parameterized queries. |
| **Game Integrity** | Provably fair shuffle: client seed + server seed. Server seed hash revealed after hand. |

### 6.2 Anti-Cheating (Phase 1)

| Detection | Method |
|---|---|
| **Multi-accounting** | Flag same IP, device fingerprint, or behavior pattern |
| **Collusion** | Detect soft-play and chip-dumping between accounts |
| **Bot detection** | Monitor decision time variance |
| **Geolocation** | Block play from prohibited jurisdictions (future) |

### 6.3 Expansion-Ready Architecture

- **Audit log:** Every hand, bet, admin action logged immutably (append-only PostgreSQL partition)
- **Event stream:** All game events published to message queue for real-time fraud analysis
- **Admin dashboard:** View flagged accounts, review hands, freeze/ban users
- **Dispute resolution:** Full hand replay with timestamps

### 6.4 Current-Phase Compliance (Virtual Currency)

- Terms of Service clearly state "play money entertainment, no real-world value"
- Age gate: 18+ registration
- Responsible gaming: Daily play time reminders, voluntary self-exclusion
- Wallet system separated from game engine for future real-money support

---

## 7. Multi-Language & Localization

### 7.1 Supported Languages

| Language | Priority | Use Case |
|---|---|---|
| **简体中文** | Primary | Main player base |
| **English** | Secondary | International players, expats |
| **မြန်မာဘာသာ** (Burmese) | Secondary | Local Myanmar players |

### 7.2 Implementation

**Frontend:** Flutter `intl` package + ARB files.  
**Backend:** Error messages and push notifications localized per user preference.

### 7.3 Game Terminology Standardization

| English | 简体中文 | မြန်မာဘာသာ |
|---|---|---|
| Big Blind | 大盲 | ဘလိုင့်ကြီး |
| Small Blind | 小盲 | ဘလိုင့်ငယ် |
| Flop | 翻牌 | ဖလော့ပ် |
| Turn | 转牌 | တန်း |
| River | 河牌 | ရီဗာ |
| All-in | 全下 | အော်လ်အင် |
| Royal Flush | 皇家同花顺 | ရော်ယယ်ဖလ်ရှ် |

### 7.4 Number & Currency Formatting

| Locale | Format |
|---|---|
| zh-CN | 1,234.56 bb / 1,234.56 金币 |
| en | 1,234.56 bb / 1,234.56 Gold |
| my-MM | ၁,၂၃၄.၅၆ (optional Burmese numerals) |

Default in-game display: **bb** for table action, **金币** for wallet/buy-in screens.

---

## 8. Currency System — 金币 (Gold) & BB (Big Blinds)

### 8.1 Currency Model

| Unit | Type | Description |
|---|---|---|
| **MMK** | Real money | Only for purchasing Gold. Never appears in-game. |
| **金币 (Gold)** | Platform currency | Purchased with MMK. Used for buy-ins and purchases. |
| **BB** | Game unit | In-table betting unit. Exchangeable to/from Gold at rates varying by game mode and level. |

**Core rule:** Game never shows USD or MMK. Players see only 金币 and BB.

### 8.2 Gold Purchase Pricing (MMK)

Exchange rates configured dynamically in admin panel using Myanmar black-market rates.

| Gold Pack | MMK Price |
|---|---|
| 60 金币 | 3,000 Ks |
| 300 金币 | 15,000 Ks |
| 1,280 金币 | 60,000 Ks |
| 3,880 金币 | 150,000 Ks |
| 8,880 金币 | 300,000 Ks |

**Admin config:** `MMK_PER_GOLD` ratio adjusted monthly.

### 8.3 Cash Games: Fixed Exchange Rate

Cash tables have **fixed** BB→Gold rates.

| Table Level | 1 BB = Gold | Big Blind (Gold) | Buy-in (Gold) |
|---|---|---|---|
| Bronze | 10 | 20 | 200–2,000 |
| Silver | 50 | 100 | 1,000–10,000 |
| Gold | 200 | 400 | 4,000–40,000 |
| Platinum | 1,000 | 2,000 | 20,000–200,000 |
| Diamond | 5,000 | 10,000 | 100K–1M |

### 8.4 Tournament / SNG / Spin & Go: Dynamic Exchange Rate

BB→Gold **increases each blind level**. Combines standard blind increases with exchange-rate multipliers.

**Example Tournament Structure:**

| Level | SB (BB) | BB (BB) | Ante | Exchange Rate | BB Gold Value |
|---|---|---|---|---|---|
| 1 | 10 | 20 | — | ×1 | 20 |
| 2 | 15 | 30 | — | ×1 | 30 |
| 3 | 20 | 40 | — | ×2 | 80 |
| 4 | 30 | 60 | 5 | ×2 | 120 |
| 5 | 50 | 100 | 10 | ×4 | 400 |
| 6 | 100 | 200 | 20 | ×4 | 800 |
| 7 | 200 | 400 | 40 | ×8 | 3,200 |
| 8 | 300 | 600 | 60 | ×8 | 4,800 |
| 9 | 500 | 1,000 | 100 | ×16 | 16,000 |
| 10 | 1,000 | 2,000 | 200 | ×16 | 32,000 |

**Mechanics:**
- Buy-in: Fixed Gold amount (e.g., 1,000 Gold) → receive 1,000 BB starting stack
- During play: UI displays stack in BB or current Gold equivalent
- Payouts: Fixed prize pool based on buy-ins (standard tournament payout)
- Dynamic rate adds visual and psychological stakes — every BB won in later levels feels more valuable

### 8.5 Economy Sinks

| Sink | Description |
|---|---|
| **Rake** | 3–5% of pot, capped at 3 BB. Destroyed. |
| **Tournament fee** | 10% of buy-in is platform fee. |
| **Club fees** | Club creation, custom table hosting. |
| **Cosmetics** | Avatars, card backs, emotes (Gold only). |

### 8.6 Free-to-Play Pipeline

| Source | Daily Reward |
|---|---|
| Daily login | 50–200 bb |
| Daily missions | 100–200 bb |
| Spin wheel | 50–1,000 bb (once/day) |
| Ad rewards | 100 bb (max 5/day) |
| Referral | 1,000 bb (on friend's first hand) |

**Daily cap:** ~2,000 bb/day. Sufficient for micro-stakes indefinitely.

---

## 9. Weak Network Optimization (Myanmar-Specific)

### 9.1 Network Quality Detection

The Flutter app continuously monitors RTT, packet loss, and connection type.

| Network Tier | RTT | Action |
|---|---|---|
| **Excellent** | <150ms | Full animations, real-time chat, observer mode |
| **Good** | 150–400ms | Standard mode. Minor animation simplification. |
| **Fair** | 400–800ms | **Lite mode:** Reduced animation frames, compressed graphics, delta updates only. Chat limited to emojis. |
| **Poor** | >800ms or >5% loss | **Emergency mode:** Pause non-critical sync. Queue actions locally. Auto-fold on timer expiry. |

### 9.2 Protocol Optimizations

| Technique | Implementation |
|---|---|
| **MessagePack** | Binary serialization, 30–40% smaller than JSON |
| **Payload compression** | zlib for messages >512 bytes |
| **Heartbeat tuning** | Normal: 30s. Poor network: 10s with 3-strike disconnect. |
| **Action queuing** | Local SQLite queue. Actions sent on reconnect. |
| **Predictive UI** | Optimistic updates: show bet immediately, confirm/revert on server ACK. |

### 9.3 Asset Strategy

| Asset | Strategy |
|---|---|
| Card decks | 2 resolution sets: HD (retina) and SD (low-bandwidth). Auto-switch by network tier. |
| Avatars | Aggressively cached. Low-res placeholders on poor network. |
| Animations | Pre-rendered sprites. Skip entirely in Lite mode. |
| Sounds | Optional download. Default off on cellular. |

### 9.4 Reconnect Resilience

```
Detect disconnect
├── Immediate: Show "Reconnecting..." overlay (non-blocking)
├── Retry 1: Immediate WebSocket reconnect
├── Retry 2–3: Exponential backoff (1s, 2s, 4s)
├── Retry 4+: Switch to backup endpoint
└── After 60s: Auto-fold. Preserve seat for 5 minutes.
```

**State recovery on reconnect:**
1. Client sends `RESUME {table_id, last_seq_id}`
2. Server sends delta from `last_seq_id`, or full snapshot if gap >10 messages
3. Client replays queued local actions
4. Target: back in game within 2 seconds

### 9.5 Offline-Tolerant Features

| Feature | Offline Behavior |
|---|---|
| Hand history | Fully cached locally. Viewable without network. |
| Club announcements | Cached. Show last-known with timestamp. |
| Player profile | Cached. Show stale data with "last updated" indicator. |
| Chat history | Cached last 50 messages per channel. |

---

## 10. Payment Integration — KBZPay Priority

### 10.1 Integration Method

**Primary:** KBZPay Merchant API (direct integration, ~1.5% fee, instant notification).  
**Fallback:** KBZPay QR Payment (user scans QR, manual reconciliation).  
**Future:** WavePay, AYAPay, bank transfer.

### 10.2 Gold Purchase Flow (KBZPay Merchant API)

```
Player selects Gold pack
    │
    ▼
Flutter → Node.js API: CREATE_ORDER {pack_id, channel: "kbzpay"}
    │
    ▼
Node.js creates pending order
    │
    ▼
Node.js → KBZPay API: initPayment {amount, order_id, callback_url}
    │
    ▼
KBZPay returns payment URL / deep-link
    │
    ▼
Flutter opens KBZPay app or in-app WebView
    │
    ▼
Player confirms payment in KBZPay
    │
    ▼
KBZPay → Node.js callback: PAYMENT_SUCCESS {order_id, transaction_ref}
    │
    ▼
Node.js validates signature, marks complete, credits Gold
    │
    ▼
Flutter receives WebSocket event: "Purchase successful!"
```

### 10.3 Security

- **Webhook validation:** All KBZPay callbacks verified with HMAC signature
- **Idempotency:** Duplicate callbacks ignored
- **Reconciliation:** Daily automated reconciliation of KBZPay ledger vs. platform orders
- **Dispute handling:** 48-hour chargeback window. Admin dashboard for review/refund.

---

## 11. Data Flow & Real-Time State Synchronization

### 11.1 Dual Connection Model

| Connection | Protocol | Purpose |
|---|---|---|
| **Game Socket** | WebSocket (port 8443) | Real-time table state: cards, bets, timers, player actions |
| **API Socket** | WebSocket (port 8444) | Social: chat, friend status, club notifications, lobby updates |

### 11.2 Delta-State Sync Protocol

```
Client ──► Server: JOIN_TABLE {table_id, auth_token}
Server ──► Client: STATE_SNAPSHOT {full_table_state}
Server ──► Client: DELTA {player_3: {action: "RAISE", amount: 200}}
Server ──broadcast──► All clients: DELTA {pot: 1200, current_turn: 4}
```

- **Snapshot:** On join, reconnection, or desync detection.
- **Delta:** 95% of messages. Only changed fields.
- **Sequence numbers:** Every message tagged with `seq_id`. Client detects gaps and requests re-sync.

### 11.3 Action Timeline (One Betting Round)

| Time | Event |
|---|---|
| T+0ms | Player taps "Raise 200" |
| T+20ms | Flutter sends ACTION {type: "RAISE", amount: 200, seq: 42} |
| T+50ms | API Gateway validates JWT |
| T+80ms | Go Game Server receives action |
| T+100ms | Server validates turn, amount, timer |
| T+120ms | Server updates table state in memory |
| T+150ms | Server broadcasts DELTA to all players |
| T+200ms | Flutter clients receive delta, render animation |
| T+300ms | Next player's turn timer starts |

**Target:** <200ms end-to-end for Myanmar (server in Singapore or Yangon).

### 11.4 Bot Service Integration

```
Bot Service ──gRPC──► Go Game Server
  └── Each bot opens a "virtual WebSocket"
  └── Bot decisions computed locally, sent via gRPC
  └── Game Server treats bot actions identically to human actions
```

Bot-to-server latency target: <100ms.

---

## 12. Testing Strategy

| Layer | Approach | Coverage |
|---|---|---|
| **Go Game Server** | Unit tests for hand evaluator, pot calculator, side-pot logic. Property-based tests for shuffle randomness. | 90% |
| **Go Game Server** | Integration tests: full 9-handed hands with bots, verify state transitions, payouts, edge cases | All game modes |
| **Bot Service** | Unit tests for decision model. Simulation: 10,000 hands per difficulty level, verify target stats. | — |
| **Node.js API** | Unit tests for business logic. Integration tests for auth, payments, club CRUD. | 80% |
| **Flutter** | Widget tests for UI flows. E2E tests: join → play hand → leave. | Critical paths |
| **Load Testing** | 100 concurrent tables (900 players) + 500 bots. Measure latency, memory, reconnects. | p99 <300ms |

**Critical game integrity test cases:**
- Royal flush vs. straight flush showdown
- 3-way all-in with different stack sizes (side pot correctness)
- Disconnection during betting, turn, and showdown
- Blind level advancement (verify BB exchange rate update)
- Bot collusion detection (two bots at same table should not soft-play)
- Weak network simulation: packet loss, 3G throttling, reconnect mid-hand

---

## 13. Technology Stack Summary

| Layer | Technology | Reasoning |
|---|---|---|
| **Mobile App** | Flutter | Single codebase for iOS/Android. Custom poker UI renders smoothly. Excellent animation support. |
| **Game Server** | Go | Goroutines handle thousands of concurrent tables. Sub-100ms latency. Memory-efficient. |
| **API Server** | Node.js + TypeScript | Fast REST API development. Rich ecosystem for auth, payments, admin. |
| **Database** | PostgreSQL | ACID compliance for financial transactions. Excellent JSON support for flexible schemas. |
| **Cache / PubSub** | Redis | Session store, real-time leaderboards, table state backup, message queue. |
| **Message Format** | MessagePack | Binary serialization, 30-40% smaller than JSON. Critical for Myanmar networks. |
| **Game Protocol** | WebSocket + delta sync | Real-time bidirectional. Delta updates minimize bandwidth. |
| **Bot→Game** | gRPC | High-performance internal RPC. Type-safe contracts. |
| **Payment** | KBZPay Merchant API | Myanmar's dominant mobile wallet. |
| **Deployment** | Docker + Kubernetes | Containerized services. Horizontal scaling for game servers and bot service. |

---

## Appendix A: Glossary

| Term | Definition |
|---|---|
| **BB** | Big Blind. The primary betting unit in Texas Hold'em. |
| **VPIP** | Voluntarily Put money In Pot. Percentage of hands a player plays. |
| **PFR** | Pre-Flop Raise. Percentage of hands a player raises before the flop. |
| **TAG** | Tight-Aggressive. A solid poker playing style. |
| **Rake** | Platform fee taken from each pot. |
| **C-bet** | Continuation bet. A bet on the flop by the pre-flop raiser. |
