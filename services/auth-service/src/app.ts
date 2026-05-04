import express, { Application } from "express";
import mongoose from "mongoose";
import { globalErrorHandler } from "./middleware/error.middleware";
import { metricsHandler } from "./monitoring/prometheus";
import { AppContainer } from "./app.container";
import { setupAuthRoutes } from "./modules/auth/auth.routes";

const createApp = (container: AppContainer): Application => {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));

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