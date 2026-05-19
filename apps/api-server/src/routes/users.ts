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
