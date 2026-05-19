import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import { createGoldOrder, handleCallback } from '../services/kbzpay';

const router = Router();

router.post('/gold-order', authenticate, async (req, res, next) => {
  try {
    const { packId } = req.body;
    if (!packId || typeof packId !== 'string') {
      return res.status(400).json({ error: 'packId required' });
    }
    const result = await createGoldOrder((req as any).userId, packId);
    res.json(result);
  } catch (err) { next(err); }
});

router.post('/callback', async (req, res, next) => {
  try {
    const { payload, signature } = req.body;
    if (!payload || !signature) {
      return res.status(400).json({ error: 'payload and signature required' });
    }
    const result = await handleCallback(payload, signature);
    res.json(result);
  } catch (err) { next(err); }
});

export default router;
