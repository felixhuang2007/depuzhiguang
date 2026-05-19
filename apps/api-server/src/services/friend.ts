import { prisma } from '../db';

export async function sendFriendRequest(userId: string, friendId: string) {
  if (userId === friendId) {
    throw new Error('Cannot add yourself');
  }

  const existing = await prisma.friend.findUnique({
    where: { userId_friendId: { userId, friendId } },
  });
  if (existing) throw new Error('Friend request already exists');

  return prisma.friend.create({
    data: { userId, friendId, status: 'pending' },
  });
}

export async function acceptFriendRequest(userId: string, requestId: string) {
  const request = await prisma.friend.findUnique({ where: { id: requestId } });
  if (!request || request.friendId !== userId) throw new Error('Not authorized');

  await prisma.friend.update({
    where: { id: requestId },
    data: { status: 'accepted' },
  });

  // Create reciprocal entry if not exists
  const reciprocal = await prisma.friend.findUnique({
    where: { userId_friendId: { userId, friendId: request.userId } },
  });
  if (!reciprocal) {
    await prisma.friend.create({
      data: { userId, friendId: request.userId, status: 'accepted' },
    });
  }

  return { success: true };
}

export async function getFriends(userId: string) {
  return prisma.friend.findMany({
    where: { userId, status: 'accepted' },
    include: {
      friend: { select: { id: true, username: true, nickname: true, avatar: true } },
    },
  });
}

export async function getPendingRequests(userId: string) {
  return prisma.friend.findMany({
    where: { friendId: userId, status: 'pending' },
    include: {
      user: { select: { id: true, username: true, avatar: true } },
    },
  });
}
