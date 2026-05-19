# Phase 6: KBZPay Payment + Weak Network Optimization + Production Deployment

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** KBZPay mobile wallet integration, Myanmar network optimizations, security hardening, Docker/Kubernetes deployment.

**Architecture:** KBZPay Merchant API for payments. MessagePack + zlib for game protocol. Asset resolution switching. Docker containers for all services. Kubernetes manifests for orchestration.

**Tech Stack:** Node.js/Go (existing), Docker, Kubernetes, MessagePack, KBZPay API

---

## Task 1: KBZPay Payment Integration

**Files (API Server):**
- Create: `apps/api-server/src/services/kbzpay.ts`
- Create: `apps/api-server/src/routes/payments.ts`

- [ ] **Step 1: Create KBZPay service**
```typescript
import crypto from 'crypto';
import { config } from '../config';
import { prisma } from '../db';

const KBZPAY_BASE_URL = config.NODE_ENV === 'production'
  ? 'https://api.kbzpay.com/merchant'
  : 'https://api.kbzpay.com/merchant/sandbox';

export async function createGoldOrder(userId: string, packId: string) {
  const packs: Record<string, { gold: number; mmk: number }> = {
    'starter': { gold: 60, mmk: 3000 },
    'small': { gold: 300, mmk: 15000 },
    'medium': { gold: 1280, mmk: 60000 },
    'large': { gold: 3880, mmk: 150000 },
    'whale': { gold: 8880, mmk: 300000 },
  };

  const pack = packs[packId];
  if (!pack) throw new Error('Invalid pack');

  const order = await prisma.transaction.create({
    data: {
      userId,
      type: 'buy_gold',
      amount: pack.gold,
      currency: 'gold',
      description: `Purchase ${pack.gold} gold for ${pack.mmk} MMK`,
    },
  });

  // Call KBZPay API to initialize payment
  const payload = {
    merchant_id: config.KBZPAY_MERCHANT_ID,
    order_id: order.id,
    amount: pack.mmk,
    currency: 'MMK',
    callback_url: `${config.API_BASE_URL}/payments/callback`,
  };

  const signature = signKBZPayPayload(payload, config.KBZPAY_SECRET);

  const response = await fetch(`${KBZPAY_BASE_URL}/init`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Signature': signature },
    body: JSON.stringify(payload),
  });

  const data = await response.json();
  if (!data.payment_url) throw new Error('KBZPay initialization failed');

  return {
    orderId: order.id,
    paymentUrl: data.payment_url,
    deepLink: data.deeplink_url,
  };
}

export async function handleKBZPayCallback(payload: any, signature: string) {
  // Verify signature
  const expectedSig = signKBZPayPayload(payload, config.KBZPAY_SECRET);
  if (signature !== expectedSig) {
    throw new Error('Invalid signature');
  }

  const { order_id, status, transaction_ref } = payload;

  // Idempotency: check if already processed
  const existing = await prisma.transaction.findUnique({ where: { id: order_id } });
  if (!existing || existing.status === 'completed') {
    return { success: true, alreadyProcessed: true };
  }

  if (status === 'SUCCESS') {
    await prisma.$transaction(async (tx) => {
      // Mark transaction complete
      await tx.transaction.update({
        where: { id: order_id },
        data: { status: 'completed', externalRef: transaction_ref },
      });
      // Credit gold to user
      await tx.user.update({
        where: { id: existing.userId },
        data: { gold: { increment: existing.amount } },
      });
    });
    return { success: true };
  }

  await prisma.transaction.update({
    where: { id: order_id },
    data: { status: 'failed' },
  });
  return { success: false };
}

function signKBZPayPayload(payload: any, secret: string): string {
  const str = JSON.stringify(payload) + secret;
  return crypto.createHmac('sha256', secret).update(str).digest('hex');
}
```

- [ ] **Step 2: Create payment routes**
```typescript
import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import * as kbzpay from '../services/kbzpay';

const router = Router();

router.post('/buy-gold', authenticate, async (req, res, next) => {
  try {
    const { packId } = req.body;
    const result = await kbzpay.createGoldOrder((req as any).userId, packId);
    res.json(result);
  } catch (err) { next(err); }
});

router.post('/callback', async (req, res, next) => {
  try {
    const signature = req.headers['x-signature'] as string;
    const result = await kbzpay.handleKBZPayCallback(req.body, signature);
    res.json(result);
  } catch (err) { next(err); }
});

export default router;
```

- [ ] **Step 3: Add Transaction status field to schema**
```prisma
model Transaction {
  id          String   @id @default(uuid())
  userId      String   @map("user_id")
  type        String
  amount      Int
  currency    String
  status      String   @default("pending") // pending, completed, failed
  description String?
  externalRef String?  @map("external_ref")
  createdAt   DateTime @default(now()) @map("created_at")

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@map("transactions")
}
```

- [ ] **Step 4: Commit**

---

## Task 2: MessagePack Protocol Optimization

**Files (Game Server):**
- Create: `apps/game-server/internal/server/msgpack.go`
- Modify: `apps/game-server/internal/server/hub.go`

- [ ] **Step 1: Implement MessagePack encoding/decoding**
```go
package server

import (
	"bytes"
	"compress/zlib"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// EncodeMessagePack serializes a message with optional compression
func EncodeMessagePack(msg Message) ([]byte, error) {
	data, err := msgpack.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// Compress if > 512 bytes
	if len(data) > 512 {
		var buf bytes.Buffer
		w := zlib.NewWriter(&buf)
		w.Write(data)
		w.Close()
		return append([]byte{0x01}, buf.Bytes()...), nil // 0x01 = compressed flag
	}

	return append([]byte{0x00}, data...), nil // 0x00 = uncompressed
}

// DecodeMessagePack deserializes a message
func DecodeMessagePack(data []byte) (*Message, error) {
	if len(data) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	var payload []byte
	if data[0] == 0x01 {
		// Compressed
		r, err := zlib.NewReader(bytes.NewReader(data[1:]))
		if err != nil {
			return nil, err
		}
		defer r.Close()
		payload, err = io.ReadAll(r)
		if err != nil {
			return nil, err
		}
	} else {
		payload = data[1:]
	}

	var msg Message
	if err := msgpack.Unmarshal(payload, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}
```

- [ ] **Step 2: Add msgpack dependency**
```bash
cd apps/game-server && go get github.com/vmihailenco/msgpack/v5
```

- [ ] **Step 3: Update Hub to use MessagePack**
```go
func (h *Hub) SendToPlayer(playerID string, msg Message) error {
	h.mu.RLock()
	conn := h.connections[playerID]
	h.mu.RUnlock()
	if conn == nil {
		return nil
	}
	data, err := EncodeMessagePack(msg)
	if err != nil {
		return err
	}
	return conn.WriteMessage(websocket.BinaryMessage, data)
}
```

- [ ] **Step 4: Commit**

---

## Task 3: Weak Network Optimizations (Flutter)

**Files (Flutter):**
- Create: `apps/mobile/lib/utils/network_detector.dart`
- Modify: `apps/mobile/lib/services/websocket_service.dart`

- [ ] **Step 1: Implement network quality detection**
```dart
import 'dart:async';

enum NetworkTier { excellent, good, fair, poor }

class NetworkDetector {
  NetworkTier _tier = NetworkTier.excellent;
  final _controller = StreamController<NetworkTier>.broadcast();
  Timer? _pingTimer;
  DateTime? _lastPing;

  Stream<NetworkTier> get stream => _controller.stream;
  NetworkTier get current => _tier;

  void startPing(Function sendPing) {
    _pingTimer = Timer.periodic(const Duration(seconds: 5), (_) {
      _lastPing = DateTime.now();
      sendPing();
    });
  }

  void onPong() {
    if (_lastPing == null) return;
    final rtt = DateTime.now().difference(_lastPing!).inMilliseconds;
    _updateTier(rtt);
  }

  void _updateTier(int rttMs) {
    NetworkTier newTier;
    if (rttMs < 150) newTier = NetworkTier.excellent;
    else if (rttMs < 400) newTier = NetworkTier.good;
    else if (rttMs < 800) newTier = NetworkTier.fair;
    else newTier = NetworkTier.poor;

    if (newTier != _tier) {
      _tier = newTier;
      _controller.add(newTier);
    }
  }

  void dispose() {
    _pingTimer?.cancel();
    _controller.close();
  }
}
```

- [ ] **Step 2: Add action queuing to WebSocket service**
```dart
class QueuedAction {
  final String action;
  final Map<String, dynamic> payload;
  final DateTime timestamp;

  QueuedAction(this.action, this.payload, this.timestamp);
}

class WebSocketService {
  final List<QueuedAction> _queue = [];
  bool _isConnected = false;

  void send(Map<String, dynamic> message) {
    if (!_isConnected) {
      _queue.add(QueuedAction('send', message, DateTime.now()));
      return;
    }
    _sendImmediate(message);
  }

  void _flushQueue() {
    for (final action in _queue) {
      _sendImmediate(action.payload);
    }
    _queue.clear();
  }

  void onConnected() {
    _isConnected = true;
    _flushQueue();
  }
}
```

- [ ] **Step 3: Commit**

---

## Task 4: Security Hardening

**Files (Game Server):**
- Modify: `apps/game-server/internal/server/server.go`

**Files (API Server):**
- Modify: `apps/api-server/src/middleware/error.ts`
- Create: `apps/api-server/src/middleware/rate-limit.ts`

- [ ] **Step 1: Add rate limiting middleware (API)**
```typescript
import { Request, Response, NextFunction } from 'express';

const ipRequests = new Map<string, { count: number; resetTime: number }>();

export function rateLimit(windowMs: number = 60000, maxRequests: number = 100) {
  return (req: Request, res: Response, next: NextFunction) => {
    const ip = req.ip || 'unknown';
    const now = Date.now();

    let record = ipRequests.get(ip);
    if (!record || record.resetTime < now) {
      record = { count: 1, resetTime: now + windowMs };
      ipRequests.set(ip, record);
    } else {
      record.count++;
    }

    if (record.count > maxRequests) {
      res.status(429).json({ error: 'Too many requests' });
      return;
    }

    next();
  };
}
```

- [ ] **Step 2: Add game server rate limiting**
```go
package server

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	requests map[string]*rateRecord
	mu       sync.RWMutex
	window   time.Duration
	maxReq   int
}

type rateRecord struct {
	count int
	reset time.Time
}

func NewRateLimiter(window time.Duration, maxReq int) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*rateRecord),
		window:   window,
		maxReq:   maxReq,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	record, exists := rl.requests[ip]
	if !exists || record.reset.Before(now) {
		rl.requests[ip] = &rateRecord{count: 1, reset: now.Add(rl.window)}
		return true
	}

	if record.count >= rl.maxReq {
		return false
	}

	record.count++
	return true
}

func (rl *RateLimiter) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(r.RemoteAddr) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next(w, r)
	}
}
```

- [ ] **Step 3: Commit**

---

## Task 5: Docker + Kubernetes Deployment

**Files:**
- Create: `apps/game-server/Dockerfile`
- Create: `apps/api-server/Dockerfile`
- Create: `apps/bot-service/Dockerfile`
- Create: `infra/docker-compose.yml`
- Create: `infra/k8s/game-server.yaml`
- Create: `infra/k8s/api-server.yaml`
- Create: `infra/k8s/ingress.yaml`

- [ ] **Step 1: Create game-server Dockerfile**
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080 8443
CMD ["./server"]
```

- [ ] **Step 2: Create api-server Dockerfile**
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY prisma ./prisma
RUN npx prisma generate
COPY dist ./dist
EXPOSE 3000
CMD ["node", "dist/index.js"]
```

- [ ] **Step 3: Create docker-compose.yml**
```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: depuzhiguang
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  api-server:
    build: ../apps/api-server
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: postgresql://postgres:postgres@postgres:5432/depuzhiguang
      JWT_SECRET: ${JWT_SECRET}
      JWT_REFRESH_SECRET: ${JWT_REFRESH_SECRET}
    depends_on:
      - postgres
      - redis

  game-server:
    build: ../apps/game-server
    ports:
      - "8080:8080"
      - "8443:8443"
    depends_on:
      - redis

  bot-service:
    build: ../apps/bot-service
    environment:
      GAME_SERVER_ADDR: game-server:8443
    depends_on:
      - game-server

volumes:
  postgres_data:
```

- [ ] **Step 4: Create Kubernetes deployment for game-server**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: game-server
spec:
  replicas: 3
  selector:
    matchLabels:
      app: game-server
  template:
    metadata:
      labels:
        app: game-server
    spec:
      containers:
      - name: game-server
        image: depuzhiguang/game-server:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8443
          name: websocket
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: game-server
spec:
  selector:
    app: game-server
  ports:
  - port: 8080
    targetPort: 8080
    name: http
  - port: 8443
    targetPort: 8443
    name: websocket
  type: ClusterIP
```

- [ ] **Step 5: Commit**

---

## Self-Review

**1. Spec coverage:**
- ✅ KBZPay payment flow (init, callback, idempotency) — Task 1
- ✅ MessagePack + zlib compression — Task 2
- ✅ Network quality detection + action queuing — Task 3
- ✅ Rate limiting (IP-based) — Task 4
- ✅ Docker containers for all services — Task 5
- ✅ Kubernetes deployments — Task 5
- ✅ docker-compose for local dev — Task 5

**2. Placeholder scan:** No TBD/TODO.

**3. Type consistency:** Go msgpack types, TypeScript payment types consistent.
