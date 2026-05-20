import { Router } from 'express';
import { PrismaClient } from '@prisma/client';

const router = Router();
const prisma = new PrismaClient();

// POST /api/sim/actions - Record an action
router.post('/actions', async (req, res) => {
  try {
    const action = await prisma.simAction.create({ data: req.body });
    res.status(201).json(action);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// POST /api/sim/results - Record a hand result
router.post('/results', async (req, res) => {
  try {
    const result = await prisma.simAction.create({ data: req.body });
    res.status(201).json(result);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// GET /api/sim/leaderboard?metric=hands_won
router.get('/leaderboard', async (req, res) => {
  const metric = String(req.query.metric || 'hands_won');
  try {
    const entries = await prisma.simLeaderboard.findMany({
      where: { metric },
      orderBy: { rank: 'asc' },
      take: 20,
    });
    res.json(entries);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// GET /api/sim/users/:id/stats
router.get('/users/:id/stats', async (req, res) => {
  try {
    const user = await prisma.user.findUnique({
      where: { id: req.params.id },
      select: {
        id: true, username: true, nickname: true, gold: true,
        handsPlayed: true, handsWon: true, vpip: true, pfr: true,
      },
    });
    if (!user) return res.status(404).json({ error: 'User not found' });

    const recentActions = await prisma.simAction.findMany({
      where: { userId: req.params.id },
      orderBy: { timestamp: 'desc' },
      take: 100,
    });

    res.json({ user, recentActions });
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

// GET /api/sim/anomalies
router.get('/anomalies', async (req, res) => {
  try {
    const anomalies = await prisma.simAnomaly.findMany({
      orderBy: { detectedAt: 'desc' },
      take: 100,
    });
    res.json(anomalies);
  } catch (err) {
    res.status(500).json({ error: String(err) });
  }
});

export default router;
