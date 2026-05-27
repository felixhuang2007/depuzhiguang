import { prisma } from '../db';

const DEFAULT_BUYIN_GOLD = 1000;
const DEFAULT_BUYIN_CHIPS = 10000;
const EXCHANGE_RATE = 10; // 1 gold = 10 chips

export async function joinTable(
  tableId: string,
  userId: string,
  opts?: { seat?: number; chips?: number }
) {
  const user = await prisma.user.findUnique({
    where: { id: userId },
    select: { gold: true },
  });

  const existing = await prisma.tablePlayer.findUnique({
    where: { tableId_userId: { tableId, userId } },
  });

  if (existing && existing.status === 'active') {
    // Already seated — return existing player without deducting gold again
    return { tablePlayer: existing, remainingGold: user?.gold ?? 0 };
  }

  if (!user) {
    throw new Error('User not found');
  }

  const buyinGold = opts?.chips
    ? Math.ceil(opts.chips / EXCHANGE_RATE)
    : DEFAULT_BUYIN_GOLD;
  const buyinChips = opts?.chips ?? DEFAULT_BUYIN_CHIPS;

  if (user.gold < buyinGold) {
    throw new Error('Insufficient gold');
  }

  let seat = opts?.seat;
  if (seat == null) {
    const activePlayers = await prisma.tablePlayer.findMany({
      where: { tableId, status: 'active' },
      select: { seat: true },
    });
    const takenSeats = new Set(activePlayers.map((p) => p.seat));
    for (let i = 0; i < 10; i++) {
      if (!takenSeats.has(i)) {
        seat = i;
        break;
      }
    }
  }

  if (seat == null) {
    throw new Error('Table is full');
  }

  const seatTaken = await prisma.tablePlayer.findUnique({
    where: { tableId_seat: { tableId, seat } },
  });

  if (seatTaken && seatTaken.status === 'active') {
    throw new Error('Seat is already taken');
  }

  const result = await prisma.$transaction(async (tx) => {
    const updatedUser = await tx.user.update({
      where: { id: userId },
      data: { gold: { decrement: buyinGold } },
      select: { gold: true },
    });

    const tablePlayer = await tx.tablePlayer.create({
      data: {
        tableId,
        userId,
        seat,
        chips: buyinChips,
        status: 'active',
      },
    });

    await tx.exchangeRecord.create({
      data: {
        userId,
        tableId,
        type: 'buyin',
        goldAmount: buyinGold,
        chipsAmount: buyinChips,
      },
    });

    return { tablePlayer, remainingGold: updatedUser.gold };
  });

  return result;
}

export async function leaveTable(tableId: string, userId: string) {
  const player = await prisma.tablePlayer.findUnique({
    where: { tableId_userId: { tableId, userId } },
  });

  if (!player || player.status !== 'active') {
    throw new Error('Not seated at this table');
  }

  const returnedGold = Math.floor(player.chips / EXCHANGE_RATE);

  const result = await prisma.$transaction(async (tx) => {
    const updatedUser = await tx.user.update({
      where: { id: userId },
      data: { gold: { increment: returnedGold } },
      select: { gold: true },
    });

    await tx.tablePlayer.update({
      where: { id: player.id },
      data: { status: 'left', leftAt: new Date() },
    });

    await tx.exchangeRecord.create({
      data: {
        userId,
        tableId,
        type: 'cashout',
        goldAmount: returnedGold,
        chipsAmount: player.chips,
      },
    });

    return { returnedGold, remainingGold: updatedUser.gold };
  });

  return result;
}

export async function getTablePlayers(tableId: string) {
  const players = await prisma.tablePlayer.findMany({
    where: { tableId, status: 'active' },
    include: {
      user: {
        select: {
          id: true,
          username: true,
          nickname: true,
          avatar: true,
        },
      },
    },
    orderBy: { seat: 'asc' },
  });

  return players.map((p) => ({
    id: p.id,
    seat: p.seat,
    chips: p.chips,
    userId: p.userId,
    username: p.user.username,
    nickname: p.user.nickname,
    avatar: p.user.avatar,
  }));
}

export async function getPlayerAtTable(tableId: string, userId: string) {
  const player = await prisma.tablePlayer.findUnique({
    where: { tableId_userId: { tableId, userId } },
    include: {
      user: {
        select: {
          id: true,
          username: true,
          nickname: true,
          avatar: true,
        },
      },
    },
  });

  if (!player || player.status !== 'active') {
    return null;
  }

  return {
    id: player.id,
    seat: player.seat,
    chips: player.chips,
    userId: player.userId,
    username: player.user.username,
    nickname: player.user.nickname,
    avatar: player.user.avatar,
  };
}
