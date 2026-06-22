import express, { Application } from "express";
import mongoose from "mongoose";

import { AppContainer } from "./app.container.js";
import { redisService } from "./config/redis.js";
import { globalErrorHandler } from "./middleware/error.middleware.js";
import { setupAuthRoutes } from "./modules/auth/auth.routes.js";
import { metricsHandler, metricsMiddleware } from "./monitoring/prometheus.js";


const createApp = (container: AppContainer): Application => {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));
  app.use(metricsMiddleware);
  app.use("/api/v1/auth", setupAuthRoutes(container.authController, container.authMiddleware));

  setupHealthRoutes(app);

  app.use(globalErrorHandler);

  return app;
};

const setupHealthRoutes = (app: Application) => {
  app.get("/api/v1/auth/metrics", metricsHandler);

  app.get("/api/v1/auth/live", (_req, res) => {
    res.status(200).json({ status: "ALIVE" });
  });
  
  app.get("/api/v1/auth/health", async (_req, res) => {
    const isDatabaseConnected = mongoose.connection.readyState === 1;

    let isRedisConnected = false;
    try {
      const pong = await redisService.ping();
      if (pong === "PONG") {
        isRedisConnected = true;
      }
    } catch (error) {
      isRedisConnected = false;
      console.error(error)
    }

    if (isDatabaseConnected && isRedisConnected) {
      return res.status(200).json({
        status: "UP",
        service: "auth-service",
        database: "connected",
        redis: "connected"
      });
    }
    return res.status(503).json({
      status: "DOWN",
      service: "auth-service",
      database: isDatabaseConnected ? "connected" : "disconnected",
      redis: isRedisConnected ? "connected" : "disconnected"
    });
  });
};

export default createApp;