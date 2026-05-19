import { Request, Response, NextFunction } from 'express';

export function errorHandler(err: any, _req: Request, res: Response, _next: NextFunction) {
  console.error(err);

  if (err.name === 'ZodError') {
    res.status(400).json({ error: 'Validation error', details: err.errors });
    return;
  }

  if (err.code === 'P2002') {
    res.status(409).json({ error: 'Duplicate entry' });
    return;
  }

  res.status(err.status || 500).json({
    error: err.message || 'Internal server error',
  });
}
