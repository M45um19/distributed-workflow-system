/* eslint-disable-next-line*/
import sdk from './monitoring/tracing.js';

import { AppContainer } from './app.container.js';
import { dbConfig } from './config/db.js';
import { kafkaConfig } from './config/kafka.js';
import { redisService } from './config/redis.js';
import { socketConfig } from './config/socket.js';

const startWorker = async () => {
  try {
    await dbConfig.connect();
    await redisService.ping();

    await kafkaConfig.connect();

    const container = new AppContainer(true);
    socketConfig.initWorker();
    console.info('Activating Notification Event Consumer Loop...');
    await container.kafkaWorker?.start();

  } catch (error) {
    console.error('Notification Event Worker startup failed:', error);
    process.exit(1);
  }
};

const shutdown = async (signal: string) => {
    console.info(`Received ${signal}. Shutting down worker...`);

    try { await kafkaConfig.disconnect(); } catch (e) { console.error(e); }

    try {
        console.log("Shutting down OpenTelemetry SDK...");
        await sdk.shutdown();
        console.log("Tracing terminated cleanly.");
    } catch (e) {
        console.error("Tracing close error:", e);
    }

    try { await socketConfig.shutdown(); } catch (e) { console.error(e); }

    try { await dbConfig.disconnect(); } catch (e) { console.error(e); }

    try { await redisService.quit(); } catch (e) { console.error(e); }

    process.exit(0);
};

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT', () => shutdown('SIGINT'));

startWorker();