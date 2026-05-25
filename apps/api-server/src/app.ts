import express from 'express';
import authRoutes from './routes/auth';
import userRoutes from './routes/users';
import clubRoutes from './routes/clubs';
import chatRoutes from './routes/chat';
import friendRoutes from './routes/friends';
import handRoutes from './routes/hands';
import paymentRoutes from './routes/payments';
import simRoutes from './routes/sim';
import { errorHandler } from './middleware/error';
import { logger } from './logger';

const app = express();

app.use(express.json());

app.use((req, res, next) => {
  const start = Date.now();
  res.on('finish', () => {
    logger.info('http_request', {
      method: req.method,
      path: req.originalUrl,
      status: res.statusCode,
      duration_ms: Date.now() - start,
    });
  });
  next();
});

app.get('/', (_req, res) => {
  res.json({ status: 'ok', service: 'api-server', time: new Date().toISOString() });
});

app.use('/api/auth', authRoutes);
app.use('/api/users', userRoutes);
app.use('/api/clubs', clubRoutes);
app.use('/api/chat', chatRoutes);
app.use('/api/friends', friendRoutes);
app.use('/api/hands', handRoutes);
app.use('/api/payments', paymentRoutes);
app.use('/api/sim', simRoutes);

app.use(errorHandler);

export default app;
