import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { prisma } from '../db';
import { config } from '../config';

export async function register(username: string, email: string, password: string) {
  const existing = await prisma.user.findFirst({
    where: { OR: [{ username }, { email }] },
  });
  if (existing) {
    throw new Error('Username or email already exists');
  }

  const hashedPassword = await bcrypt.hash(password, 12);
  const user = await prisma.user.create({
    data: { username, email, password: hashedPassword },
  });

  return { id: user.id, username: user.username };
}

export async function login(username: string, password: string) {
  const user = await prisma.user.findUnique({ where: { username } });
  if (!user) {
    throw new Error('Invalid credentials');
  }

  const valid = await bcrypt.compare(password, user.password);
  if (!valid) {
    throw new Error('Invalid credentials');
  }

  const accessToken = jwt.sign({ userId: user.id }, config.JWT_SECRET as jwt.Secret, {
    expiresIn: config.JWT_ACCESS_EXPIRY,
  } as jwt.SignOptions);

  const refreshToken = jwt.sign({ userId: user.id }, config.JWT_REFRESH_SECRET as jwt.Secret, {
    expiresIn: config.JWT_REFRESH_EXPIRY,
  } as jwt.SignOptions);

  await prisma.refreshToken.create({
    data: {
      token: refreshToken,
      userId: user.id,
      expiresAt: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000),
    },
  });

  return { accessToken, refreshToken, user: { id: user.id, username: user.username } };
}

export async function refresh(token: string) {
  const stored = await prisma.refreshToken.findUnique({ where: { token } });
  if (!stored || stored.expiresAt < new Date()) {
    throw new Error('Invalid refresh token');
  }

  const payload = jwt.verify(token, config.JWT_REFRESH_SECRET) as { userId: string };

  const accessToken = jwt.sign({ userId: payload.userId }, config.JWT_SECRET as jwt.Secret, {
    expiresIn: config.JWT_ACCESS_EXPIRY,
  } as jwt.SignOptions);

  return { accessToken };
}

export function verifyAccessToken(token: string) {
  return jwt.verify(token, config.JWT_SECRET) as { userId: string };
}
