import { Router } from 'express';
import { authenticate } from '../middleware/auth';
import { joinTable, leaveTable, getTablePlayers, getPlayerAtTable } from '../services/table';

const router = Router();

// GET /api/tables/:id/players - 获取牌桌上的活跃玩家（公开）
router.get('/:id/players', async (req, res, next) => {
  try {
    const players = await getTablePlayers(req.params.id);
    res.json({ players });
  } catch (err) {
    next(err);
  }
});

// GET /api/tables/:id/me - 查询当前用户是否在该桌
router.get('/:id/me', authenticate, async (req, res, next) => {
  try {
    const userId = (req as any).userId;
    const player = await getPlayerAtTable(req.params.id, userId);
    res.json({ seated: !!player, player });
  } catch (err) {
    next(err);
  }
});

// POST /api/tables/:id/join - 加入牌桌
router.post('/:id/join', authenticate, async (req, res, next) => {
  try {
    const userId = (req as any).userId;
    const { seat, chips } = req.body;
    const result = await joinTable(req.params.id, userId, { seat, chips });
    res.status(201).json({ success: true, ...result });
  } catch (err) {
    next(err);
  }
});

// POST /api/tables/:id/leave - 离开牌桌
router.post('/:id/leave', authenticate, async (req, res, next) => {
  try {
    const userId = (req as any).userId;
    const result = await leaveTable(req.params.id, userId);
    res.json({ success: true, ...result });
  } catch (err) {
    next(err);
  }
});

export default router;
