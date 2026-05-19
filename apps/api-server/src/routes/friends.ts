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

router.get('/pending', authenticate, async (req, res, next) => {
  try {
    const requests = await friendService.getPendingRequests((req as any).userId);
    res.json(requests);
  } catch (err) { next(err); }
});

export default router;
