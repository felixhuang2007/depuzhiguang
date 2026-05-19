import express from 'express';
import authRoutes from './routes/auth';
import userRoutes from './routes/users';
import clubRoutes from './routes/clubs';
import chatRoutes from './routes/chat';
import friendRoutes from './routes/friends';
import handRoutes from './routes/hands';
import paymentRoutes from './routes/payments';
import { errorHandler } from './middleware/error';

const app = express();

app.use(express.json());

app.use('/auth', authRoutes);
app.use('/users', userRoutes);
app.use('/clubs', clubRoutes);
app.use('/chat', chatRoutes);
app.use('/friends', friendRoutes);
app.use('/hands', handRoutes);
app.use('/payments', paymentRoutes);

app.use(errorHandler);

export default app;
