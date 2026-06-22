import http from 'http';
import { Server } from 'http';

import { AppContainer } from './app.container.js';
import createApp from './app.js';
import { dbConfig } from './config/db.js';
import { env } from './config/env.js';
import { redisService } from './config/redis.js';
import { socketConfig } from './config/socket.js';

let httpServer: Server;
let isShuttingDown = false;
const isProduction = env.NODE_ENV === "production" || env.NODE_ENV === "staging";

const startServer = async () => {
  try {
    await dbConfig.connect();
    await redisService.ping();

    const container = new AppContainer(false);
    const app = createApp(container);

    httpServer = http.createServer(app);
    socketConfig.init(httpServer);
    
    const PORT = env.PORT;
    httpServer.listen(PORT, () => {
      console.log(`Notification API & Socket Server running on port: ${PORT}`);
    });
  } catch (error) {
    console.error('Critical: Notification Server startup failed:', error);
    process.exit(1);
  }
};

const releaseResources = async () => {
  try {
    await socketConfig.shutdown();
  } catch (err) { console.error("Socket close error:", err); }

  try {
    if (redisService && typeof redisService.quit === 'function') {
      await redisService.quit();
      console.log("Redis client disconnected.");
    }
  } catch (err) { console.error("Redis quit error:", err); }

  try {
    if (dbConfig && typeof dbConfig.disconnect === 'function') {
      await dbConfig.disconnect();
      console.log("Database connection closed safely.");
    }
  } catch (err) { console.error("Database disconnect error:", err); }
};

const shutdown = async (signal: string) => {
  if (isShuttingDown) return;
  isShuttingDown = true;

  console.log(`\nReceived ${signal}. Starting graceful shutdown for Notification Service...`);

  if (httpServer) {
    console.log("Stopping new HTTP/Socket connections...");
    await new Promise<void>((resolve) => {
      httpServer.close((err) => {
        if (err) console.error("Error during HTTP server close:", err);
        else console.log("HTTP server closed.");
        resolve();
      });
    });
  }

  if (isProduction) {
    console.log("Production Mode: Waiting 5 seconds for traffic drain...");
    await new Promise((resolve) => setTimeout(resolve, 5000));
  }

  await releaseResources();

  console.log("Graceful shutdown completed successfully. Exiting.");
  process.exit(0);
};

process.on("SIGTERM", () => shutdown("SIGTERM"));
process.on("SIGINT", () => shutdown("SIGINT"));

startServer();