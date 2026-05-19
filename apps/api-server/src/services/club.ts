import { prisma } from '../db';

export async function createClub(ownerId: string, name: string, description?: string, joinType: string = 'approval') {
  const club = await prisma.club.create({
    data: {
      name,
      description,
      ownerId,
      joinType,
      members: {
        create: { userId: ownerId, role: 'owner' },
      },
    },
  });
  return club;
}

export async function getClubById(clubId: string) {
  return prisma.club.findUnique({
    where: { id: clubId },
    include: {
      owner: { select: { id: true, username: true } },
      members: {
        include: { user: { select: { id: true, username: true, avatar: true } } },
      },
    },
  });
}

export async function joinClub(clubId: string, userId: string) {
  const club = await prisma.club.findUnique({ where: { id: clubId } });
  if (!club) throw new Error('Club not found');
  if (club.joinType === 'invite') throw new Error('Club is invite-only');

  const existing = await prisma.clubMember.findUnique({
    where: { clubId_userId: { clubId, userId } },
  });
  if (existing) throw new Error('Already a member');

  const role = club.joinType === 'public' ? 'member' : 'pending';
  return prisma.clubMember.create({
    data: { clubId, userId, role },
  });
}

export async function getUserClubs(userId: string) {
  return prisma.clubMember.findMany({
    where: { userId },
    include: {
      club: {
        include: {
          owner: { select: { id: true, username: true } },
          _count: { select: { members: true } },
        },
      },
    },
  });
}
