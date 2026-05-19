import { prisma } from '../db';

export async function getUserById(id: string) {
  return prisma.user.findUnique({
    where: { id },
    select: {
      id: true,
      username: true,
      nickname: true,
      avatar: true,
      gold: true,
      bb: true,
      handsPlayed: true,
      handsWon: true,
      vpip: true,
      pfr: true,
      createdAt: true,
    },
  });
}

export async function updateProfile(userId: string, data: { nickname?: string; avatar?: string }) {
  return prisma.user.update({
    where: { id: userId },
    data,
    select: { id: true, username: true, nickname: true, avatar: true },
  });
}

export async function getLeaderboard(limit: number = 100) {
  return prisma.user.findMany({
    orderBy: { gold: 'desc' },
    take: limit,
    select: {
      id: true,
      username: true,
      nickname: true,
      avatar: true,
      gold: true,
      handsWon: true,
    },
  });
}
