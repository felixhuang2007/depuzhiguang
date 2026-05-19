import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import * as chatService from '../services/chat';

const router = Router();

router.post('/table/:tableId', authenticate, async (req, res, next) => {
  try {
    const { content } = req.body;
    const msg = await chatService.sendTableMessage(req.params.tableId, (req as any).userId, content);
    res.status(201).json(msg);
  } catch (err) { next(err); }
});

router.get('/table/:tableId', async (req, res, next) => {
  try {
    const limit = parseInt(req.query.limit as string) || 50;
    const messages = await chatService.getTableMessages(req.params.tableId, limit);
    res.json(messages);
  } catch (err) { next(err); }
});

router.post('/club/:clubId', authenticate, async (req, res, next) => {
  try {
    const { content } = req.body;
    const msg = await chatService.sendClubMessage(req.params.clubId, (req as any).userId, content);
    res.status(201).json(msg);
  } catch (err) { next(err); }
});

export default router;
