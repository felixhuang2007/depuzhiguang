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

app.use('/auth', authRoutes);
app.use('/users', userRoutes);
app.use('/clubs', clubRoutes);
app.use('/chat', chatRoutes);
app.use('/friends', friendRoutes);
app.use('/hands', handRoutes);
app.use('/payments', paymentRoutes);
app.use('/sim', simRoutes);

app.use(errorHandler);

export default app;
