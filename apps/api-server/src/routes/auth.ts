import { Router } from 'express';
import { z } from 'zod';
import { register, login, refresh } from '../services/auth';

const router = Router();

const registerSchema = z.object({
  username: z.string().min(3).max(30),
  email: z.string().email(),
  password: z.string().min(6),
  nickname: z.string().optional(),
  isSimUser: z.boolean().optional(),
  simStyle: z.string().optional(),
  simPersonality: z.string().optional(),
});

const loginSchema = z.object({
  username: z.string(),
  password: z.string(),
});

router.post('/register', async (req, res, next) => {
  try {
    const data = registerSchema.parse(req.body);
    const user = await register(data.username, data.email, data.password, {
      nickname: data.nickname,
      isSimUser: data.isSimUser,
      simStyle: data.simStyle,
      simPersonality: data.simPersonality,
    });
    res.status(201).json(user);
  } catch (err: any) {
    if (err.message === 'Username or email already exists') {
      res.status(409).json({ error: err.message });
      return;
    }
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
