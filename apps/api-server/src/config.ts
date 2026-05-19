import { z } from 'zod';

const envSchema = z.object({
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.string().default('3000'),
  DATABASE_URL: z.string().default('postgresql://localhost:5432/depuzhiguang'),
  JWT_SECRET: z.string().min(32).default('change-me-in-production-32-chars-min'),
  JWT_REFRESH_SECRET: z.string().min(32).default('change-me-too-in-production-32-chars'),
  JWT_ACCESS_EXPIRY: z.string().default('15m'),
  JWT_REFRESH_EXPIRY: z.string().default('7d'),
  API_BASE_URL: z.string().optional(),
  KBZPAY_MERCHANT_ID: z.string().optional(),
  KBZPAY_SECRET: z.string().optional(),
});

export const config = envSchema.parse(process.env);
