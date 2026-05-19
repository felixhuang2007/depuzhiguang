import { describe, it, expect, beforeEach } from 'vitest';
import request from 'supertest';
import app from '../src/app';
import { cleanup } from './setup';

describe('Auth', () => {
  beforeEach(async () => {
    await cleanup();
  });

  it('should register a new user', async () => {
    const res = await request(app)
      .post('/auth/register')
      .send({ username: 'alice', email: 'alice@test.com', password: 'password123' });

    expect(res.status).toBe(201);
    expect(res.body.username).toBe('alice');
  });

  it('should reject duplicate username', async () => {
    await request(app)
      .post('/auth/register')
      .send({ username: 'alice', email: 'alice@test.com', password: 'password123' });

    const res = await request(app)
      .post('/auth/register')
      .send({ username: 'alice', email: 'alice2@test.com', password: 'password123' });

    expect(res.status).toBe(500);
  });

  it('should login and return tokens', async () => {
    await request(app)
      .post('/auth/register')
      .send({ username: 'bob', email: 'bob@test.com', password: 'password123' });

    const res = await request(app)
      .post('/auth/login')
      .send({ username: 'bob', password: 'password123' });

    expect(res.status).toBe(200);
    expect(res.body.accessToken).toBeDefined();
    expect(res.body.refreshToken).toBeDefined();
  });

  it('should reject invalid credentials', async () => {
    const res = await request(app)
      .post('/auth/login')
      .send({ username: 'nobody', password: 'wrong' });

    expect(res.status).toBe(500);
  });
});
