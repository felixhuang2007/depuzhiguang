import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import * as handHistoryService from '../services/handhistory';

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
