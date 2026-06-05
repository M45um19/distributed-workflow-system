import { AppContainer } from './app.container.js';
import { dbConfig } from './config/db.js';
import { kafkaConfig } from './config/kafka.js';
import { redisService } from './config/redis.js';

const startWorker = async () => {
  try {
    await dbConfig.connect();
    await redisService.ping();
  

    await kafkaConfig.connect();

    const container = new AppContainer(true);
    
    console.info('Activating Notification Event Consumer Loop...');
    await container.kafkaWorker?.start();

  } catch (error) {
    console.error('Notification Event Worker startup failed:', error);
    process.exit(1);
  }
};

const shutdown = async () => {
  console.info('Gracefully stopping Notification Consumer Worker...');
  await kafkaConfig.disconnect();
  process.exit(0);
};

process.on('SIGINT', shutdown);
process.on('SIGTERM', shutdown);

startWorker();