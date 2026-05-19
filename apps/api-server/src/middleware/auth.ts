import { Request, Response, NextFunction } from 'express';
import { verifyAccessToken } from '../services/auth';

export function authenticate(req: Request, res: Response, next: NextFunction) {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith('Bearer ')) {
    res.status(401).json({ error: 'Unauthorized' });
    return;
  }

  try {
    const token = authHeader.slice(7);
    const payload = verifyAccessToken(token);
    (req as any).userId = payload.userId;
    next();
  } catch {
    res.status(401).json({ error: 'Invalid token' });
  }
}
