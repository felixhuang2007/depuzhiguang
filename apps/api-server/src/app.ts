import express from 'express';
import authRoutes from './routes/auth';
import userRoutes from './routes/users';
import clubRoutes from './routes/clubs';
import { errorHandler } from './middleware/error';

const app = express();

app.use(express.json());

app.use('/auth', authRoutes);
app.use('/users', userRoutes);
app.use('/clubs', clubRoutes);

app.use(errorHandler);

export default app;
