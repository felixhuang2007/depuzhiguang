import { prisma } from '../db';

export async function sendTableMessage(tableId: string, playerId: string, content: string) {
  if (content.length > 200) {
    throw new Error('Message too long (max 200 chars)');
  }

  // Block URLs for accounts < 24h old
  if (content.includes('http') && content.includes('://')) {
    const player = await prisma.user.findUnique({ where: { id: playerId } });
    if (player && Date.now() - player.createdAt.getTime() < 24 * 60 * 60 * 1000) {
      throw new Error('New accounts cannot send URLs');
    }
  }

  return prisma.chatMessage.create({
    data: { tableId, playerId, content, channel: 'table' },
  });
}

export async function getTableMessages(tableId: string, limit: number = 50) {
  return prisma.chatMessage.findMany({
    where: { tableId, channel: 'table' },
    orderBy: { createdAt: 'desc' },
    take: limit,
    include: { player: { select: { id: true, username: true, avatar: true } } },
  });
}

export async function sendClubMessage(clubId: string, playerId: string, content: string) {
  if (content.length > 200) {
    throw new Error('Message too long (max 200 chars)');
  }

  return prisma.chatMessage.create({
    data: { clubId, playerId, content, channel: 'club' },
  });
}
