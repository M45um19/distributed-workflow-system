import { Server } from "http";

import { AppContainer } from "./app.container.js";
import createApp from "./app.js";
import { dbConfig } from "./config/db.js";
import { env } from "./config/env.js";
import { grpcConfig } from "./config/grpc.js";
import { kafkaConfig } from "./config/kafka.js";
import { redisService } from "./config/redis.js";
import { registerAuthGrpcService } from "./modules/auth/auth.grpc.js";
import sdk from "./monitoring/tracing.js";

let httpServer: Server;
let isShuttingDown = false;
const isProduction = env.NODE_ENV === "production" || env.NODE_ENV === "staging";

const startServer = async () => {
  try {
    await dbConfig.connect();
    await redisService.ping();
    await kafkaConfig.connect();

    const container = new AppContainer();
    const app = createApp(container);

    registerAuthGrpcService(container.authService);
    grpcConfig.start(50051);

    const PORT = env.PORT || 5000;
    httpServer = app.listen(PORT, () => {
      console.log(`Auth Service running on port: ${PORT}`);
    });

  } catch (error) {
    console.error("Critical: Server startup failed:", error);
    process.exit(1);
  }
};

const releaseResources = async () => {
  try {
    console.log("Shutting down OpenTelemetry SDK...");
    await sdk.shutdown();
    console.log("Tracing terminated cleanly.");
    if (grpcConfig && typeof grpcConfig.stop === 'function') {
      await grpcConfig.stop();
      console.log("gRPC server stopped safely.");
    }
  } catch (err) { console.error("gRPC stop error:", err); }

  try {
    if (kafkaConfig && typeof kafkaConfig.disconnect === 'function') {
      await kafkaConfig.disconnect();
      console.log("Kafka producer disconnected.");
    }
  } catch (err) { console.error("Kafka disconnect error:", err); }

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

  console.log(`\nReceived ${signal}. Starting perfect graceful shutdown...`);

  if (httpServer) {
    console.log("Initiating HTTP server close. Stopping new connection acceptance...");
    if (typeof httpServer.closeIdleConnections === 'function') {
      httpServer.closeIdleConnections();
    }
    
    await new Promise<void>((resolve) => {
      httpServer.close((err) => {
        if (err) console.error("Error during HTTP server close:", err);
        else console.log("HTTP server closed. No longer accepting new HTTP requests.");
        resolve(); 
      });
    });
  }

  if (isProduction) {
    console.log("Production Mode: Waiting 5 seconds for existing traffic to drain and K8s endpoints to propagate...");
    await new Promise((resolve) => setTimeout(resolve, 5000));
    console.log("Drain period completed. Releasing connections...");
  } else {
    console.log("⚡ Development Mode: Skipping drain delay. Quick cleaning resources...");
  }

  await releaseResources();

  console.log("Graceful shutdown completed successfully. Exiting process.");
  process.exit(0);
};

if (isProduction) {
  process.on("SIGTERM", () => shutdown("SIGTERM"));
  process.on("SIGINT", () => shutdown("SIGINT"));
} else {
  process.on("SIGINT", () => shutdown("Local_SIGINT"));

  process.once("SIGUSR2", async () => {
    console.log("\n[Nodemon] Restarting... Cleaning up local connections.");
    await releaseResources();
    process.kill(process.pid, "SIGUSR2");
  });
}

startServer();