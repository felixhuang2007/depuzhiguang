# 德扑之光 — Master Implementation Roadmap

This document maps the full spec into **6 phased implementation plans**. Each phase produces working, testable software and builds on previous phases.

---

## Phase 1: Go Poker Engine + Game Server Core ✅ PLANNED
**File:** `2026-05-19-phase1-poker-engine.md`

**Goal:** Core Texas Hold'em engine — cards, deck, hand evaluator, pot calculator, table state machine, WebSocket server foundation.

**Deliverables:**
- `Card` / `Deck` / `Suit` / `Rank` primitives
- 5-card hand evaluator (all categories + tie-breakers)
- Pot calculator (main pot + side pots)
- Table configuration and seat management
- Game state machine (preflop → showdown)
- Betting action validation (fold, check, call, bet, raise, all-in)
- Showdown logic with hand comparison
- WebSocket message protocol (MessagePack)
- Connection hub with disconnect handling
- HTTP server bootstrap with health check

**Est. Duration:** 3–4 days

---

## Phase 2: Node.js API Server + Database
**Status:** To be planned

**Goal:** RESTful API for users, authentication, clubs, leaderboards, hand history. PostgreSQL schema and migrations.

**Key Components:**
- User registration/login (JWT + refresh tokens)
- Player profiles and stats (VPIP, PFR, win rate)
- Club CRUD (create, join, manage, invite-only)
- Club member roles (Owner/Manager/Agent/Member)
- Custom table creation within clubs
- Leaderboard queries (window functions)
- Hand history storage (JSONB)
- Admin dashboard endpoints
- Database schema + migrations (golang-migrate or Atlas)

**Est. Duration:** 4–5 days

---

## Phase 3: Flutter Mobile App — Core UI
**Status:** To be planned

**Goal:** Cross-platform mobile app with poker table UI, WebSocket client, and game interaction.

**Key Components:**
- Project initialization (Flutter 3.x, null safety)
- Theme and design system (dark poker table aesthetic)
- WebSocket client with reconnect logic
- Network quality detection + lite mode UI
- Login / registration screens
- Lobby screen (table list, filters, stakes)
- Poker table screen:
  - Seat layout (2/6/9 players)
  - Card rendering (HD/SD assets)
  - Chip stack animation
  - Betting controls (fold/check/call/raise slider)
  - Timer countdown
  - Pot display
- Hand history viewer
- Profile screen (stats, settings)
- Multi-language support (intl + ARB files)

**Est. Duration:** 7–10 days

---

## Phase 4: Bot Service (Simulated Users)
**Status:** To be planned

**Goal:** 500+ concurrent AI players with realistic poker behavior.

**Key Components:**
- Bot lifecycle manager (spawn/pause/retire)
- Table allocator (fill empty seats intelligently)
- AI decision engine:
  - Hand strength vs. range
  - Pot odds calculator
  - Position-aware ranges
  - Bluff frequency by difficulty
  - Randomization / humanization
- 4 difficulty levels (Fish/Regular/Shark/Whale)
- Identity generator (names, avatars, fake stats)
- Humanization behaviors (timing variance, tilt, breaks)
- gRPC client to game server
- Scaling target: 500 bots, <500ms latency

**Est. Duration:** 5–7 days

---

## Phase 5: Social System + Club Features
**Status:** To be planned

**Goal:** Friends, clubs, chat, observer mode, hand replay sharing.

**Key Components:**
- Friend system (add by ID/QR, online status)
- Club creation and management
- Club chip economy (owner buys from platform, sells to members)
- In-game chat (table, private, club) with emoji reactions
- Translation layer (Chinese ↔ Burmese) for chat
- Observer mode (spectate without playing)
- Hand replay viewer (street-by-street)
- Share hand as image / replay link
- Club activity feed (notable hands)

**Est. Duration:** 5–6 days

---

## Phase 6: Payment + Polish + Production Prep
**Status:** To be planned

**Goal:** KBZPay integration, weak network optimization, security hardening, deployment.

**Key Components:**
- KBZPay Merchant API integration
- Gold purchase flow (pack selection → payment → credit)
- Webhook validation + idempotency
- Daily reconciliation job
- Weak network optimizations:
  - MessagePack serialization
  - Payload compression
  - Action queuing (SQLite)
  - Optimistic UI updates
  - Asset resolution switching
  - Reconnect resilience
- Security:
  - Rate limiting
  - Input validation
  - SQL injection prevention
  - Provably fair shuffle verification
  - Audit logging
- Deployment:
  - Docker images for all services
  - docker-compose for local dev
  - Kubernetes manifests for production
  - CI/CD pipeline (GitHub Actions)

**Est. Duration:** 5–7 days

---

## Total Estimated Timeline

| Phase | Duration | Cumulative |
|---|---|---|
| Phase 1: Poker Engine | 3–4 days | Week 1 |
| Phase 2: API + DB | 4–5 days | Week 2 |
| Phase 3: Flutter App | 7–10 days | Week 3–4 |
| Phase 4: Bot Service | 5–7 days | Week 5 |
| Phase 5: Social | 5–6 days | Week 6 |
| Phase 6: Payment + Prod | 5–7 days | Week 7 |

**Total: ~7 weeks** for MVP with 2 developers (1 backend, 1 Flutter).

**Parallelization possible:**
- Phase 1 + Phase 3 can start in parallel once message protocol is defined.
- Phase 4 can start once Phase 1 game server API is stable.
- Phase 5 depends on Phase 2 (clubs) and Phase 3 (chat UI).

---

*Phase 1 plan is detailed and ready for execution. Subsequent phases will be planned in detail when Phase 1 is complete.*
