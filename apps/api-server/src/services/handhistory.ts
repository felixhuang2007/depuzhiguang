import { prisma } from '../db';

export async function saveHandHistory(userId: string, tableId: string, handData: any, result: string, wonAmount: number) {
  return prisma.handHistory.create({
    data: {
      userId,
      tableId,
      handData: JSON.stringify(handData),
      result,
      wonAmount,
    },
  });
}

export async function getHandHistory(userId: string, limit: number = 50) {
  return prisma.handHistory.findMany({
    where: { userId },
    orderBy: { playedAt: 'desc' },
    take: limit,
  });
}

export async function getHandById(handId: string, userId: string) {
  const hand = await prisma.handHistory.findUnique({ where: { id: handId } });
  if (!hand || hand.userId !== userId) throw new Error('Hand not found');
  return hand;
}
