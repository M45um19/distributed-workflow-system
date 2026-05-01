import express from "express";
import dotenv from "dotenv";
import { authRouter } from "./modules/auth/auth.routes";
import { metricsHandler } from "./monitoring/prometheus";
import mongoose from "mongoose";
import { globalErrorHandler } from "./middleware/error.middleware";
import { WorkspaceRouter } from "./modules/workspace/workspace.routes";

dotenv.config();

const app = express();

// Middlewares
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use("/api/v1/auth", authRouter);
app.use("/api/v1/workspace", WorkspaceRouter)

// Health check (K8s readiness probe later)
app.get("/api/v1/auth/health", (_req, res) => {
  const isDatabaseConnected = mongoose.connection.readyState === 1;

  if (isDatabaseConnected) {
    return res.status(200).json({
      status: "UP",
      service: "auth-service",
      database: "connected"
    });
  } else {
    return res.status(503).json({
      status: "DOWN",
      service: "auth-service",
      database: "disconnected"
    });
  }
});

app.get("/api/v1/auth/metrics", metricsHandler)

app.use(globalErrorHandler)

export default app;