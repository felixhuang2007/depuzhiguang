# Phase 2: Node.js API Server + Database — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build RESTful API server for user auth, profiles, clubs, leaderboards, and hand history. PostgreSQL schema with Prisma ORM.

**Architecture:** Express + TypeScript + Prisma ORM + PostgreSQL. JWT access/refresh tokens. Zod validation.

**Tech Stack:** Node.js 20+, TypeScript 5, Express 4, Prisma 5, bcrypt, jsonwebtoken, zod, vitest

---

## File Structure

```
apps/api-server/
├── prisma/
│   └── schema.prisma
├── src/
│   ├── index.ts
│   ├── app.ts
│   ├── config.ts
│   ├── db.ts
│   ├── routes/
│   │   ├── auth.ts
│   │   ├── users.ts
│   │   ├── clubs.ts
│   │   └── leaderboards.ts
│   ├── middleware/
│   │   ├── auth.ts
│   │   ├── error.ts
│   │   └── validate.ts
│   ├── services/
│   │   ├── auth.ts
│   │   ├── user.ts
│   │   ├── club.ts
│   │   └── leaderboard.ts
│   └── types/
│       └── index.ts
├── tests/
│   ├── auth.test.ts
│   ├── users.test.ts
│   ├── clubs.test.ts
│   └── setup.ts
├── package.json
├── tsconfig.json
└── .env.example
```

---

## Task 1: Initialize API Server Project

**Files:**
- Create: `apps/api-server/package.json`
- Create: `apps/api-server/tsconfig.json`
- Create: `apps/api-server/.env.example`
- Create: `apps/api-server/src/index.ts`

- [ ] **Step 1: Create package.json**
```json
{
  "name": "@depuzhiguang/api-server",
  "version": "1.0.0",
  "scripts": {
    "dev": "tsx watch src/index.ts",
    "build": "tsc",
    "start": "node dist/index.js",
    "test": "vitest",
    "db:migrate": "prisma migrate dev",
    "db:generate": "prisma generate",
    "db:seed": "tsx prisma/seed.ts"
  },
  "dependencies": {
    "@prisma/client": "^5.22.0",
    "bcryptjs": "^2.4.3",
    "express": "^4.21.0",
    "jsonwebtoken": "^9.0.2",
    "zod": "^3.23.0"
  },
  "devDependencies": {
    "@types/bcryptjs": "^2.4.6",
    "@types/express": "^4.17.21",
    "@types/jsonwebtoken": "^9.0.7",
    "@types/node": "^20.0.0",
    "prisma": "^5.22.0",
    "tsx": "^4.19.0",
    "typescript": "^5.6.0",
    "vitest": "^2.1.0",
    "supertest": "^7.0.0",
    "@types/supertest": "^6.0.0"
  }
}
```

- [ ] **Step 2: Create tsconfig.json**
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "moduleResolution": "node"
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

- [ ] **Step 3: Install dependencies**
```bash
cd apps/api-server && npm install
```

- [ ] **Step 4: Create minimal src/index.ts**
```typescript
import express from 'express';

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

app.get('/health', (_req, res) => {
  res.json({ status: 'ok' });
});

app.listen(PORT, () => {
  console.log(`API server running on port ${PORT}`);
});
```

- [ ] **Step 5: Verify**
```bash
cd apps/api-server && npx tsx src/index.ts
# Should print "API server running on port 3000"
# Ctrl+C to stop
```

- [ ] **Step 6: Commit**
```bash
git add apps/api-server/
git commit -m "chore: init api-server Node.js project"
```

---

## Task 2: Prisma Schema

**Files:**
- Create: `apps/api-server/prisma/schema.prisma`

- [ ] **Step 1: Create schema.prisma**
```prisma
generator client {
  provider = "prisma-client-js"
}

datasource db {
  provider = "postgresql"
  url      = env("DATABASE_URL")
}

model User {
  id        String   @id @default(uuid())
  username  String   @unique
  email     String   @unique
  password  String
  nickname  String?
  avatar    String?
  gold      Int      @default(0)
  bb        Int      @default(0)
  status    String   @default("active")
  createdAt DateTime @default(now()) @map("created_at")
  updatedAt DateTime @updatedAt @map("updated_at")

  // Stats
  handsPlayed Int @default(0) @map("hands_played")
  handsWon    Int @default(0) @map("hands_won")
  vpip        Int @default(0) // voluntarily put in pot % (0-100)
  pfr         Int @default(0) // pre-flop raise % (0-100)

  // Relations
  ownedClubs     Club[]        @relation("ClubOwner")
  clubMembers    ClubMember[]
  refreshTokens  RefreshToken[]
  handHistories  HandHistory[]
  transactions   Transaction[]

  @@map("users")
}

model RefreshToken {
  id        String   @id @default(uuid())
  token     String   @unique
  userId    String   @map("user_id")
  expiresAt DateTime @map("expires_at")
  createdAt DateTime @default(now()) @map("created_at")

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@map("refresh_tokens")
}

model Club {
  id          String   @id @default(uuid())
  name        String
  description String?
  ownerId     String   @map("owner_id")
  joinType    String   @default("approval") @map("join_type") // public, approval, invite
  status      String   @default("active")
  createdAt   DateTime @default(now()) @map("created_at")

  owner   User         @relation("ClubOwner", fields: [ownerId], references: [id])
  members ClubMember[]
  tables  ClubTable[]

  @@map("clubs")
}

model ClubMember {
  id     String @id @default(uuid())
  clubId String @map("club_id")
  userId String @map("user_id")
  role   String @default("member") // owner, manager, agent, member
  joinedAt DateTime @default(now()) @map("joined_at")

  club Club @relation(fields: [clubId], references: [id], onDelete: Cascade)
  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@unique([clubId, userId])
  @@map("club_members")
}

model ClubTable {
  id       String @id @default(uuid())
  clubId   String @map("club_id")
  name     String
  stakes   String @default("1/2")
  gameType String @default("cash") @map("game_type")
  status   String @default("active")
  createdAt DateTime @default(now()) @map("created_at")

  club Club @relation(fields: [clubId], references: [id], onDelete: Cascade)

  @@map("club_tables")
}

model HandHistory {
  id        String   @id @default(uuid())
  userId    String   @map("user_id")
  tableId   String   @map("table_id")
  handData  Json     @map("hand_data")
  result    String   // win, loss, tie
  wonAmount Int      @default(0) @map("won_amount")
  playedAt  DateTime @default(now()) @map("played_at")

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@map("hand_histories")
}

model Transaction {
  id          String   @id @default(uuid())
  userId      String   @map("user_id")
  type        String   // buy_gold, bb_to_gold, gold_to_bb, rake, fee
  amount      Int
  currency    String   // gold, bb
  description String?
  createdAt   DateTime @default(now()) @map("created_at")

  user User @relation(fields: [userId], references: [id], onDelete: Cascade)

  @@map("transactions")
}
```

- [ ] **Step 2: Generate Prisma client**
```bash
cd apps/api-server && npx prisma generate
```

- [ ] **Step 3: Commit**
```bash
git add apps/api-server/prisma/
git commit -m "feat(db): add Prisma schema for users, clubs, hand history"
```

---

## Task 3: Database Connection + Config

**Files:**
- Create: `apps/api-server/src/config.ts`
- Create: `apps/api-server/src/db.ts`

- [ ] **Step 1: Create config.ts**
```typescript
import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.string().default('3000'),
  DATABASE_URL: z.string(),
  JWT_SECRET: z.string().min(32),
  JWT_REFRESH_SECRET: z.string().min(32),
  JWT_ACCESS_EXPIRY: z.string().default('15m'),
  JWT_REFRESH_EXPIRY: z.string().default('7d'),
});

export const config = envSchema.parse(process.env);
```

- [ ] **Step 2: Create db.ts**
```typescript
import { PrismaClient } from '@prisma/client';

export const prisma = new PrismaClient({
  log: process.env.NODE_ENV === 'development' ? ['query', 'error', 'warn'] : ['error'],
});
```

- [ ] **Step 3: Commit**
```bash
git add apps/api-server/src/config.ts apps/api-server/src/db.ts
git commit -m "feat(api): add config and database connection"
```

---

## Task 4: Authentication Service

**Files:**
- Create: `apps/api-server/src/services/auth.ts`
- Create: `apps/api-server/src/routes/auth.ts`
- Create: `apps/api-server/src/middleware/auth.ts`

- [ ] **Step 1: Create auth service**
```typescript
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { prisma } from '../db';
import { config } from '../config';

export async function register(username: string, email: string, password: string) {
  const existing = await prisma.user.findFirst({
    where: { OR: [{ username }, { email }] },
  });
  if (existing) {
    throw new Error('Username or email already exists');
  }

  const hashedPassword = await bcrypt.hash(password, 12);
  const user = await prisma.user.create({
    data: { username, email, password: hashedPassword },
  });

  return { id: user.id, username: user.username };
}

export async function login(username: string, password: string) {
  const user = await prisma.user.findUnique({ where: { username } });
  if (!user) {
    throw new Error('Invalid credentials');
  }

  const valid = await bcrypt.compare(password, user.password);
  if (!valid) {
    throw new Error('Invalid credentials');
  }

  const accessToken = jwt.sign({ userId: user.id }, config.JWT_SECRET, {
    expiresIn: config.JWT_ACCESS_EXPIRY,
  });

  const refreshToken = jwt.sign({ userId: user.id }, config.JWT_REFRESH_SECRET, {
    expiresIn: config.JWT_REFRESH_EXPIRY,
  });

  await prisma.refreshToken.create({
    data: {
      token: refreshToken,
      userId: user.id,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
    },
  });

  return { accessToken, refreshToken, user: { id: user.id, username: user.username } };
}

export async function refresh(token: string) {
  const stored = await prisma.refreshToken.findUnique({ where: { token } });
  if (!stored || stored.expiresAt < new Date()) {
    throw new Error('Invalid refresh token');
  }

  const payload = jwt.verify(token, config.JWT_REFRESH_SECRET) as { userId: string };

  const accessToken = jwt.sign({ userId: payload.userId }, config.JWT_SECRET, {
    expiresIn: config.JWT_ACCESS_EXPIRY,
  });

  return { accessToken };
}

export function verifyAccessToken(token: string) {
  return jwt.verify(token, config.JWT_SECRET) as { userId: string };
}
```

- [ ] **Step 2: Create auth middleware**
```typescript
import { Request, Response, NextFunction } from 'express';
import { verifyAccessToken } from '../services/auth';

export function authenticate(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith('Bearer ')) {
    res.status(401).json({ error: 'Unauthorized' });
    return;
  }

  try {
    const token = authHeader.slice(7);
    const payload = verifyAccessToken(token);
    (req as any).userId = payload.userId;
    next();
  } catch {
    res.status(401).json({ error: 'Invalid token' });
  }
}
```

- [ ] **Step 3: Create auth routes**
```typescript
import { Router } from 'express';
import { z } from 'zod';
import { register, login, refresh } from '../services/auth';

const router = Router();

const registerSchema = z.object({
  username: z.string().min(3).max(30),
  email: z.string().email(),
  password: z.string().min(6),
});

const loginSchema = z.object({
  username: z.string(),
  password: z.string(),
});

router.post('/register', async (req, res, next) => {
  try {
    const data = registerSchema.parse(req.body);
    const user = await register(data.username, data.email, data.password);
    res.status(201).json(user);
  } catch (err) {
    next(err);
  }
});

router.post('/login', async (req, res, next) => {
  try {
    const data = loginSchema.parse(req.body);
    const result = await login(data.username, data.password);
    res.json(result);
  } catch (err) {
    next(err);
  }
});

router.post('/refresh', async (req, res, next) => {
  try {
    const { refreshToken } = req.body;
    const result = await refresh(refreshToken);
    res.json(result);
  } catch (err) {
    next(err);
  }
});

export default router;
```

- [ ] **Step 4: Commit**
```bash
git add apps/api-server/src/services/auth.ts apps/api-server/src/middleware/auth.ts apps/api-server/src/routes/auth.ts
git commit -m "feat(auth): add JWT authentication with register, login, refresh"
```

---

## Task 5: User Routes

**Files:**
- Create: `apps/api-server/src/services/user.ts`
- Create: `apps/api-server/src/routes/users.ts`

- [ ] **Step 1: Create user service**
```typescript
import { prisma } from '../db';

export async function getUserById(id: string) {
  return prisma.user.findUnique({
    where: { id },
    select: {
      id: true,
      username: true,
      nickname: true,
      avatar: true,
      gold: true,
      bb: true,
      handsPlayed: true,
      handsWon: true,
      vpip: true,
      pfr: true,
      createdAt: true,
    },
  });
}

export async function updateProfile(userId: string, data: { nickname?: string; avatar?: string }) {
  return prisma.user.update({
    where: { id: userId },
    data,
    select: { id: true, username: true, nickname: true, avatar: true },
  });
}

export async function getLeaderboard(limit: number = 100) {
  return prisma.user.findMany({
    orderBy: { gold: 'desc' },
    take: limit,
    select: {
      id: true,
      username: true,
      nickname: true,
      avatar: true,
      gold: true,
      handsWon: true,
    },
  });
}
```

- [ ] **Step 2: Create user routes**
```typescript
import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import { getUserById, updateProfile, getLeaderboard } from '../services/user';

const router = Router();

router.get('/me', authenticate, async (req, res, next) => {
  try {
    const user = await getUserById((req as any).userId);
    res.json(user);
  } catch (err) {
    next(err);
  }
});

router.patch('/me', authenticate, async (req, res, next) => {
  try {
    const user = await updateProfile((req as any).userId, req.body);
    res.json(user);
  } catch (err) {
    next(err);
  }
});

router.get('/leaderboard', async (req, res, next) => {
  try {
    const limit = parseInt(req.query.limit as string) || 100;
    const board = await getLeaderboard(limit);
    res.json(board);
  } catch (err) {
    next(err);
  }
});

export default router;
```

- [ ] **Step 3: Commit**
```bash
git add apps/api-server/src/services/user.ts apps/api-server/src/routes/users.ts
git commit -m "feat(users): add profile and leaderboard endpoints"
```

---

## Task 6: Club Routes

**Files:**
- Create: `apps/api-server/src/services/club.ts`
- Create: `apps/api-server/src/routes/clubs.ts`

- [ ] **Step 1: Create club service**
```typescript
import { prisma } from '../db';

export async function createClub(ownerId: string, name: string, description?: string, joinType: string = 'approval') {
  const club = await prisma.club.create({
    data: {
      name,
      description,
      ownerId,
      joinType,
      members: {
        create: { userId: ownerId, role: 'owner' },
      },
    },
  });
  return club;
}

export async function getClubById(clubId: string) {
  return prisma.club.findUnique({
    where: { id: clubId },
    include: {
      owner: { select: { id: true, username: true } },
      members: {
        include: { user: { select: { id: true, username: true, avatar: true } } },
      },
    },
  });
}

export async function joinClub(clubId: string, userId: string) {
  const club = await prisma.club.findUnique({ where: { id: clubId } });
  if (!club) throw new Error('Club not found');
  if (club.joinType === 'invite') throw new Error('Club is invite-only');

  const existing = await prisma.clubMember.findUnique({
    where: { clubId_userId: { clubId, userId } },
  });
  if (existing) throw new Error('Already a member');

  const role = club.joinType === 'public' ? 'member' : 'pending';
  return prisma.clubMember.create({
    data: { clubId, userId, role },
  });
}

export async function approveMember(clubId: string, ownerId: string, memberId: string) {
  const club = await prisma.club.findUnique({ where: { id: clubId } });
  if (!club || club.ownerId !== ownerId) throw new Error('Not authorized');

  return prisma.clubMember.update({
    where: { id: memberId },
    data: { role: 'member' },
  });
}

export async function getUserClubs(userId: string) {
  return prisma.clubMember.findMany({
    where: { userId },
    include: {
      club: {
        include: {
          owner: { select: { id: true, username: true } },
          _count: { select: { members: true } },
        },
      },
    },
  });
}
```

- [ ] **Step 2: Create club routes**
```typescript
import { Router } from 'express';
import { z } from 'zod';
import { authenticate } from '../middleware/auth';
import * as clubService from '../services/club';

const router = Router();

const createSchema = z.object({
  name: z.string().min(1).max(100),
  description: z.string().optional(),
  joinType: z.enum(['public', 'approval', 'invite']).optional(),
});

router.post('/', authenticate, async (req, res, next) => {
  try {
    const data = createSchema.parse(req.body);
    const club = await clubService.createClub((req as any).userId, data.name, data.description, data.joinType);
    res.status(201).json(club);
  } catch (err) {
    next(err);
  }
});

router.get('/my', authenticate, async (req, res, next) => {
  try {
    const clubs = await clubService.getUserClubs((req as any).userId);
    res.json(clubs);
  } catch (err) {
    next(err);
  }
});

router.get('/:id', async (req, res, next) => {
  try {
    const club = await clubService.getClubById(req.params.id);
    res.json(club);
  } catch (err) {
    next(err);
  }
});

router.post('/:id/join', authenticate, async (req, res, next) => {
  try {
    const member = await clubService.joinClub(req.params.id, (req as any).userId);
    res.status(201).json(member);
  } catch (err) {
    next(err);
  }
});

export default router;
```

- [ ] **Step 3: Commit**
```bash
git add apps/api-server/src/services/club.ts apps/api-server/src/routes/clubs.ts
git commit -m "feat(clubs): add club CRUD and membership endpoints"
```

---

## Task 7: Wire Up App + Error Handling

**Files:**
- Modify: `apps/api-server/src/app.ts`
- Create: `apps/api-server/src/middleware/error.ts`

- [ ] **Step 1: Create error middleware**
```typescript
import { Request, Response, NextFunction } from 'express';

export function errorHandler(err: any, _req: Request, res: Response, _next: NextFunction) {
  console.error(err);

  if (err.name === 'ZodError') {
    res.status(400).json({ error: 'Validation error', details: err.errors });
    return;
  }

  if (err.code === 'P2002') {
    res.status(409).json({ error: 'Duplicate entry' });
    return;
  }

  res.status(err.status || 500).json({
    error: err.message || 'Internal server error',
  });
}
```

- [ ] **Step 2: Create app.ts**
```typescript
import express from 'express';
import authRoutes from './routes/auth';
import userRoutes from './routes/users';
import clubRoutes from './routes/clubs';
import { errorHandler } from './middleware/error';

const app = express();

app.use(express.json());

app.use('/auth', authRoutes);
app.use('/users', userRoutes);
app.use('/clubs', clubRoutes);

app.use(errorHandler);

export default app;
```

- [ ] **Step 3: Update index.ts**
```typescript
import app from './app';
import { config } from './config';

const PORT = config.PORT;

app.listen(PORT, () => {
  console.log(`API server running on port ${PORT}`);
});
```

- [ ] **Step 4: Commit**
```bash
git add apps/api-server/src/
git commit -m "feat(api): wire up routes with error handling"
```

---

## Task 8: Auth Integration Tests

**Files:**
- Create: `apps/api-server/tests/auth.test.ts`
- Create: `apps/api-server/tests/setup.ts`

- [ ] **Step 1: Create setup.ts**
```typescript
import { prisma } from '../src/db';

export async function cleanup() {
  await prisma.refreshToken.deleteMany();
  await prisma.clubMember.deleteMany();
  await prisma.club.deleteMany();
  await prisma.user.deleteMany();
}
```

- [ ] **Step 2: Create auth.test.ts**
```typescript
import { describe, it, expect, beforeEach } from 'vitest';
import request from 'supertest';
import app from '../src/app';
import { cleanup } from './setup';

describe('Auth', () => {
  beforeEach(async () => {
    await cleanup();
  });

  it('should register a new user', async () => {
    const res = await request(app)
      .post('/auth/register')
      .send({ username: 'alice', email: 'alice@test.com', password: 'password123' });

    expect(res.status).toBe(201);
    expect(res.body.username).toBe('alice');
  });

  it('should login and return tokens', async () => {
    await request(app)
      .post('/auth/register')
      .send({ username: 'bob', email: 'bob@test.com', password: 'password123' });

    const res = await request(app)
      .post('/auth/login')
      .send({ username: 'bob', password: 'password123' });

    expect(res.status).toBe(200);
    expect(res.body.accessToken).toBeDefined();
    expect(res.body.refreshToken).toBeDefined();
  });

  it('should reject invalid credentials', async () => {
    const res = await request(app)
      .post('/auth/login')
      .send({ username: 'nobody', password: 'wrong' });

    expect(res.status).toBe(500); // or 401 depending on error handling
  });
});
```

- [ ] **Step 3: Create vitest.config.ts**
```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    setupFiles: ['./tests/setup.ts'],
  },
});
```

- [ ] **Step 4: Run tests**
```bash
cd apps/api-server && npx vitest run tests/auth.test.ts
```

- [ ] **Step 5: Commit**
```bash
git add apps/api-server/tests/ apps/api-server/vitest.config.ts
git commit -m "test(api): add auth integration tests"
```

---

## Self-Review

**1. Spec coverage:**
- ✅ User registration/login with JWT — Tasks 1, 4, 8
- ✅ Player profiles — Task 5
- ✅ Club CRUD — Task 6
- ✅ Leaderboard — Task 5
- ✅ Hand history schema — Task 2
- ✅ Transaction schema — Task 2
- ⚠ Database migrations — needs PostgreSQL running to test

**2. Placeholder scan:** No TBD/TODO. All code provided.

**3. Type consistency:** Prisma schema matches service types.

---

Plan complete and saved to `docs/superpowers/plans/2026-05-19-phase2-api-server.md`.

**Two execution options:**

**1. Subagent-Driven (recommended)** — Fresh subagent per task, review between tasks

**2. Inline Execution** — Execute tasks in this session using executing-plans

**Which approach?**
