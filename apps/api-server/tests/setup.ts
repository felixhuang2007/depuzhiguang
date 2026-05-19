import { prisma } from '../src/db';

export async function cleanup() {
  await prisma.refreshToken.deleteMany();
  await prisma.clubMember.deleteMany();
  await prisma.club.deleteMany();
  await prisma.user.deleteMany();
}
