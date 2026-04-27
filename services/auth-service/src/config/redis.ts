import Redis from 'ioredis';
import { env } from './env';

const REDIS_URI = env.REDIS_URI as string;

const redisClient = new Redis(REDIS_URI, {
    maxRetriesPerRequest: null,
    retryStrategy: (times) => {
        const delay = Math.min(times * 50, 2000);
        return delay;
    }
});

redisClient.on('connect', () => {
    console.log('Redis connected');
});

redisClient.on('error', (err) => {
    console.error('ioredis: Connection Error', err);
});

export default redisClient;