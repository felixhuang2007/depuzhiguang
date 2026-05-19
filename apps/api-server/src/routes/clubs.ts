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
