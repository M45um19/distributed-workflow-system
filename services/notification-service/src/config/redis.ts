import { Redis } from 'ioredis';

import { env } from './env.js';

export interface IRedisService {
  set(key: string, value: string | number, mode?: 'EX', duration?: number): Promise<string | null>;
  get(key: string): Promise<string | null>;
  del(key: string): Promise<number>;
  ping(): Promise<string>;
  getClient(): Redis;
}

class RedisConfig implements IRedisService {
  private client: Redis;

  constructor() {
    const REDIS_URI = env.REDIS_URI as string;

    if (!REDIS_URI) {
      throw new Error('REDIS_URI is not defined in the environment variables.');
    }

    this.client = new Redis(REDIS_URI, {
      maxRetriesPerRequest: null,
      retryStrategy: (times) => {
        const delay = Math.min(times * 50, 2000);
        return delay;
      }
    });

    this.initEventListeners();
  }

  private initEventListeners(): void {
    this.client.on('connect', () => {
      console.info('Redis connected successfully');
    });

    this.client.on('error', (err) => {
      console.warn(`Redis Connection Error: ${String(err)}`);
    });
  }
  public getClient(): Redis {
    return this.client;
  }
  async set(key: string, value: string | number, mode?: 'EX', duration?: number): Promise<string | null> {
    if (mode === 'EX' && duration !== undefined) {
      return (await this.client.set(key, value, 'EX', duration)) as string | null;
    }
    return (await this.client.set(key, value)) as string | null;
  }

  async get(key: string): Promise<string | null> {
    return (await this.client.get(key)) as string | null;
  }

  async del(key: string): Promise<number> {
    return (await this.client.del(key)) as number;
  }

  async ping(): Promise<string> {
    return (await this.client.ping()) as string;
  }

  async quit(): Promise<void> {
    await this.client.disconnect();
  }
}

export const redisService = new RedisConfig();