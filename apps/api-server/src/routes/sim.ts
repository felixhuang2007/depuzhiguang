import { Router } from 'express';
import { PrismaClient } from '@prisma/client';

const router = Router();
const prisma = new PrismaClient();

// POST /api/sim/actions - Record an action
router.post('/actions', async (req, res, next) => {
  try {
    const { userId, tableId, handNumber, phase, action, amount, potBefore, potAfter, stackBefore, stackAfter, holeCards, community } = req.body;
    if (!userId || !tableId || !phase || !action) {
      res.status(400).json({ error: 'Missing required fields: userId, tableId, phase, action' });
      return;
    }
    const record = await prisma.simAction.create({
      data: {
        sessionId: req.body.sessionId || '',
        userId,
        tableId,
        handNumber: handNumber || 0,
        phase,
        action,
        amount: amount || 0,
        potBefore: potBefore || 0,
        potAfter: potAfter || 0,
        stackBefore: stackBefore || 0,
        stackAfter: stackAfter || 0,
        holeCards: holeCards || null,
        community: community || null,
      },
    });
    res.status(201).json(record);
  } catch (err) {
    next(err);
  }
});

// POST /api/sim/results - Record a hand result and update user stats
router.post('/results', async (req, res, next) => {
  try {
    const { userId, winAmount, isWinner } = req.body;
    if (!userId || typeof winAmount !== 'number') {
      res.status(400).json({ error: 'Missing userId or winAmount' });
      return;
    }

    const user = await prisma.user.update({
      where: { id: userId },
      data: {
        gold: { increment: winAmount },
        handsPlayed: { increment: 1 },
        handsWon: isWinner ? { increment: 1 } : undefined,
      },
    });

    res.json({ user });
  } catch (err) {
    next(err);
  }
});

// GET /api/sim/leaderboard?metric=hands_won
router.get('/leaderboard', async (req, res, next) => {
  const metric = String(req.query.metric || 'hands_won');
  try {
    const entries = await prisma.simLeaderboard.findMany({
      where: { metric },
      orderBy: { rank: 'asc' },
      take: 20,
    });
    res.json(entries);
  } catch (err) {
    next(err);
  }
});

// GET /api/sim/users/:id/stats
router.get('/users/:id/stats', async (req, res, next) => {
  try {
    const user = await prisma.user.findUnique({
      where: { id: req.params.id },
      select: {
        id: true, username: true, nickname: true, gold: true,
        handsPlayed: true, handsWon: true, vpip: true, pfr: true,
      },
    });
    if (!user) {
      res.status(404).json({ error: 'User not found' });
      return;
    }

    const recentActions = await prisma.simAction.findMany({
      where: { userId: req.params.id },
      orderBy: { timestamp: 'desc' },
      take: 100,
    });

    res.json({ user, recentActions });
  } catch (err) {
    next(err);
  }
});

// GET /api/sim/anomalies
router.get('/anomalies', async (req, res, next) => {
  try {
    const anomalies = await prisma.simAnomaly.findMany({
      orderBy: { detectedAt: 'desc' },
      take: 100,
    });
    res.json(anomalies);
  } catch (err) {
    next(err);
  }
});

export default router;
