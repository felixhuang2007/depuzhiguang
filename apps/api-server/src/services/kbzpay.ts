import crypto from 'crypto';
import { config } from '../config';
import { prisma } from '../db';

const KBZPAY_BASE_URL = config.NODE_ENV === 'production'
  ? 'https://api.kbzpay.com/merchant'
  : 'https://api.kbzpay.com/merchant/sandbox';

const goldPacks: Record<string, { gold: number; mmk: number }> = {
  starter: { gold: 60, mmk: 3000 },
  small: { gold: 300, mmk: 15000 },
  medium: { gold: 1280, mmk: 60000 },
  large: { gold: 3880, mmk: 150000 },
  whale: { gold: 8880, mmk: 300000 },
};

export async function createGoldOrder(userId: string, packId: string) {
  const pack = goldPacks[packId];
  if (!pack) throw new Error('Invalid pack');

  const order = await prisma.transaction.create({
    data: {
      userId,
      type: 'buy_gold',
      amount: pack.gold,
      currency: 'gold',
      status: 'pending',
      description: `Purchase ${pack.gold} gold for ${pack.mmk} MMK`,
    },
  });

  const payload = {
    merchant_id: config.KBZPAY_MERCHANT_ID || 'test_merchant',
    order_id: order.id,
    amount: pack.mmk,
    currency: 'MMK',
    callback_url: `${config.API_BASE_URL || 'http://localhost:3000'}/payments/callback`,
  };

  const signature = signPayload(payload, config.KBZPAY_SECRET || 'test_secret');

  return {
    orderId: order.id,
    paymentUrl: `${KBZPAY_BASE_URL}/pay?order_id=${order.id}&sig=${signature}`,
  };
}

export async function handleCallback(payload: any, signature: string) {
  const expectedSig = signPayload(payload, config.KBZPAY_SECRET || 'test_secret');
  if (signature !== expectedSig) {
    throw new Error('Invalid signature');
  }

  const { order_id, status, transaction_ref } = payload;

  const existing = await prisma.transaction.findUnique({ where: { id: order_id } });
  if (!existing || existing.status === 'completed') {
    return { success: true, alreadyProcessed: true };
  }

  if (status === 'SUCCESS') {
    await prisma.$transaction(async (tx) => {
      await tx.transaction.update({
        where: { id: order_id },
        data: { status: 'completed', externalRef: transaction_ref },
      });
      await tx.user.update({
        where: { id: existing.userId },
        data: { gold: { increment: existing.amount } },
      });
    });
    return { success: true };
  }

  await prisma.transaction.update({
    where: { id: order_id },
    data: { status: 'failed' },
  });
  return { success: false };
}

function signPayload(payload: any, secret: string): string {
  const str = JSON.stringify(payload) + secret;
  return crypto.createHmac('sha256', secret).update(str).digest('hex');
}
