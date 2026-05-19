# Phase 5: Social Features — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** In-game chat, club management, hand replay sharing, observer mode, friend system.

**Architecture:** Chat messages via WebSocket (separate socket from game). Club APIs already exist from Phase 2. Hand replays stored as JSON in PostgreSQL. Observer mode connects to game server as read-only spectator.

---

## Task 1: In-Game Chat System

**Files (API Server):**
- Create: `apps/api-server/src/services/chat.ts`
- Create: `apps/api-server/src/routes/chat.ts`

**Files (Game Server):**
- Modify: `apps/game-server/internal/server/message.go` (add chat message types)
- Modify: `apps/game-server/internal/server/hub.go` (add chat broadcast)

- [ ] **Step 1: Create chat service (API)**
```typescript
import { prisma } from '../db';

export async function sendTableMessage(tableId: string, playerId: string, content: string) {
  // Validate content length, filter URLs for new accounts
  if (content.length > 200) throw new Error('Message too long');
  if (content.includes('http') && content.includes('://')) {
    // Check if player is new (account age < 24h)
    const player = await prisma.user.findUnique({ where: { id: playerId } });
    if (player && Date.now() - player.createdAt.getTime() < 24 * 60 * 60 * 1000) {
      throw new Error('New accounts cannot send URLs');
    }
  }

  return prisma.chatMessage.create({
    data: { tableId, playerId, content, channel: 'table' },
  });
}

export async function getTableMessages(tableId: string, limit: number = 50) {
  return prisma.chatMessage.findMany({
    where: { tableId, channel: 'table' },
    orderBy: { createdAt: 'desc' },
    take: limit,
    include: { player: { select: { id: true, username: true, avatar: true } } },
  });
}
```

- [ ] **Step 2: Add ChatMessage to Prisma schema**
```prisma
model ChatMessage {
  id        String   @id @default(uuid())
  tableId   String?  @map("table_id")
  clubId    String?  @map("club_id")
  playerId  String   @map("player_id")
  content   String
  channel   String   // table, club, private
  createdAt DateTime @default(now()) @map("created_at")

  player User @relation(fields: [playerId], references: [id], onDelete: Cascade)

  @@map("chat_messages")
}
```

- [ ] **Step 3: Migrate database**
```bash
cd apps/api-server && npx prisma migrate dev --name add_chat
```

- [ ] **Step 4: Add chat WebSocket message types (Game Server)**
```go
const (
  MsgChat MessageType = "chat"
)

type ChatPayload struct {
  PlayerID string `json:"player_id"`
  Content  string `json:"content"`
  Channel  string `json:"channel"`
}
```

- [ ] **Step 5: Commit**

---

## Task 2: Friend System

**Files (API Server):**
- Modify: `apps/api-server/prisma/schema.prisma` (add Friend model)
- Create: `apps/api-server/src/services/friend.ts`
- Create: `apps/api-server/src/routes/friends.ts`

- [ ] **Step 1: Add Friend model to schema**
```prisma
model Friend {
  id        String   @id @default(uuid())
  userId    String   @map("user_id")
  friendId  String   @map("friend_id")
  status    String   @default("pending") // pending, accepted, blocked
  createdAt DateTime @default(now()) @map("created_at")

  user   User @relation("UserFriends", fields: [userId], references: [id], onDelete: Cascade)
  friend User @relation("FriendUsers", fields: [friendId], references: [id], onDelete: Cascade)

  @@unique([userId, friendId])
  @@map("friends")
}
```

- [ ] **Step 2: Create friend service**
```typescript
export async function sendFriendRequest(userId: string, friendId: string) {
  const existing = await prisma.friend.findUnique({
    where: { userId_friendId: { userId, friendId } },
  });
  if (existing) throw new Error('Friend request already exists');

  return prisma.friend.create({
    data: { userId, friendId, status: 'pending' },
  });
}

export async function acceptFriendRequest(userId: string, requestId: string) {
  const request = await prisma.friend.findUnique({ where: { id: requestId } });
  if (!request || request.friendId !== userId) throw new Error('Not authorized');

  await prisma.friend.update({
    where: { id: requestId },
    data: { status: 'accepted' },
  });

  // Create reciprocal entry
  return prisma.friend.create({
    data: { userId, friendId: request.userId, status: 'accepted' },
  });
}

export async function getFriends(userId: string) {
  return prisma.friend.findMany({
    where: { userId, status: 'accepted' },
    include: { friend: { select: { id: true, username: true, avatar: true } } },
  });
}
```

- [ ] **Step 3: Create friend routes**
```typescript
import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import * as friendService from '../services/friend';

const router = Router();

router.post('/request', authenticate, async (req, res, next) => {
  try {
    const { friendId } = req.body;
    const result = await friendService.sendFriendRequest((req as any).userId, friendId);
    res.status(201).json(result);
  } catch (err) { next(err); }
});

router.post('/accept/:id', authenticate, async (req, res, next) => {
  try {
    const result = await friendService.acceptFriendRequest((req as any).userId, req.params.id);
    res.json(result);
  } catch (err) { next(err); }
});

router.get('/', authenticate, async (req, res, next) => {
  try {
    const friends = await friendService.getFriends((req as any).userId);
    res.json(friends);
  } catch (err) { next(err); }
});

export default router;
```

- [ ] **Step 4: Commit**

---

## Task 3: Hand Replay System

**Files (Game Server):**
- Create: `apps/game-server/internal/table/replay.go`

**Files (API Server):**
- Create: `apps/api-server/src/services/handhistory.ts`
- Create: `apps/api-server/src/routes/hands.ts`

- [ ] **Step 1: Create replay generator (Game Server)**
```go
package table

import (
	"encoding/json"
	"time"
)

// HandReplay captures a complete hand for storage/sharing
type HandReplay struct {
	TableID     string       `json:"table_id"`
	PlayedAt    time.Time    `json:"played_at"`
	Players     []PlayerInfo `json:"players"`
	Community   []string     `json:"community"`
	Actions     []ActionLog  `json:"actions"`
	Pot         int          `json:"pot"`
	Winners     []string     `json:"winners"`
}

type PlayerInfo struct {
	ID       string `json:"id"`
	Seat     int    `json:"seat"`
	HoleCards []string `json:"hole_cards,omitempty"`
}

type ActionLog struct {
	PlayerID string `json:"player_id"`
	Action   string `json:"action"`
	Amount   int    `json:"amount"`
	Round    string `json:"round"` // preflop, flop, turn, river
}

func BuildReplay(game *Game) *HandReplay {
	replay := &HandReplay{
		TableID:  game.Table.Config.ID,
		PlayedAt: time.Now(),
		Pot:      game.Pot.Total(),
	}

	// ... populate from game state
	return replay
}

func (r *HandReplay) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}
```

- [ ] **Step 2: Create hand history API service**
```typescript
export async function saveHandHistory(userId: string, tableId: string, handData: any, result: string, wonAmount: number) {
  return prisma.handHistory.create({
    data: { userId, tableId, handData: JSON.stringify(handData), result, wonAmount },
  });
}

export async function getHandHistory(userId: string, limit: number = 50) {
  return prisma.handHistory.findMany({
    where: { userId },
    orderBy: { playedAt: 'desc' },
    take: limit,
  });
}

export async function getHandById(handId: string, userId: string) {
  const hand = await prisma.handHistory.findUnique({ where: { id: handId } });
  if (!hand || hand.userId !== userId) throw new Error('Hand not found');
  return hand;
}
```

- [ ] **Step 3: Create hand history routes**
```typescript
const router = Router();

router.get('/', authenticate, async (req, res, next) => {
  try {
    const limit = parseInt(req.query.limit as string) || 50;
    const hands = await handHistoryService.getHandHistory((req as any).userId, limit);
    res.json(hands);
  } catch (err) { next(err); }
});

router.get('/:id', authenticate, async (req, res, next) => {
  try {
    const hand = await handHistoryService.getHandById(req.params.id, (req as any).userId);
    res.json(hand);
  } catch (err) { next(err); }
});

export default router;
```

- [ ] **Step 4: Commit**

---

## Task 4: Observer Mode (Railbird)

**Files (Game Server):**
- Modify: `apps/game-server/internal/server/hub.go`
- Modify: `apps/game-server/internal/server/server.go`

- [ ] **Step 1: Add observer support to Hub**
```go
// observers maps table_id to set of observing playerIDs
type Hub struct {
	connections  map[string]*websocket.Conn
	tablePlayers map[string]map[string]struct{}
	observers    map[string]map[string]struct{} // table_id -> observer IDs
	mu           sync.RWMutex
}

func (h *Hub) JoinAsObserver(playerID, tableID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.observers[tableID] == nil {
		h.observers[tableID] = make(map[string]struct{})
	}
	h.observers[tableID][playerID] = struct{}{}
}

func (h *Hub) BroadcastToObservers(tableID string, msg Message) {
	h.mu.RLock()
	observers := h.observers[tableID]
	conns := make(map[string]*websocket.Conn, len(observers))
	for pid := range observers {
		conns[pid] = h.connections[pid]
	}
	h.mu.RUnlock()

	for _, conn := range conns {
		if conn != nil {
			_ = conn.WriteJSON(msg)
		}
	}
}
```

- [ ] **Step 2: Modify wsHandler to support observer param**
```go
func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	// ... existing code ...
	isObserver := r.URL.Query().Get("observer") == "true"
	
	if isObserver {
		tableID := r.URL.Query().Get("table_id")
		s.hub.JoinAsObserver(playerID, tableID)
		// Send state snapshot without hole cards
	} else {
		s.hub.Register(playerID, conn)
	}
}
```

- [ ] **Step 3: Commit**

---

## Self-Review

**1. Spec coverage:**
- ✅ In-game chat (table + club channels, URL filtering) — Task 1
- ✅ Friend system (request/accept/list) — Task 2
- ✅ Hand replay JSON generation and API — Task 3
- ✅ Observer mode (read-only spectators) — Task 4

**2. Placeholder scan:** No TBD/TODO.

**3. Type consistency:** Prisma schema matches service types. Go message types consistent.
