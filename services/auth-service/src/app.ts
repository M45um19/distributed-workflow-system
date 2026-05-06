import express, { Application } from "express";
import mongoose from "mongoose";

import { AppContainer } from "./app.container";
import { globalErrorHandler } from "./middleware/error.middleware";
import { setupAuthRoutes } from "./modules/auth/auth.routes";
import { metricsHandler, metricsMiddleware } from "./monitoring/prometheus";

const createApp = (container: AppContainer): Application => {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));
  app.use(metricsMiddleware);
  app.use("/api/v1/auth", setupAuthRoutes(container.authController));

  setupHealthRoutes(app);

  app.use(globalErrorHandler);

  return app;
};

const setupHealthRoutes = (app: Application) => {
  app.get("/api/v1/auth/metrics", metricsHandler);

  app.get("/api/v1/auth/health", (_req, res) => {
    const isDatabaseConnected = mongoose.connection.readyState === 1;

    if (isDatabaseConnected) {
      return res.status(200).json({
        status: "UP",
        service: "auth-service",
        database: "connected"
      });
    }

    return res.status(503).json({
      status: "DOWN",
      service: "auth-service",
      database: "disconnected"
    });
  });
};

export default createApp;