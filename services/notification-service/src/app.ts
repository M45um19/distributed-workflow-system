import express, { Application } from 'express';
import mongoose from 'mongoose';

import { AppContainer } from './app.container.js';
import { globalErrorHandler } from './middleware/error.middleware.js';
import { setupNotificationRoutes } from './modules/notification/notification.routes.js';
import { metricsHandler, metricsMiddleware } from './monitoring/prometheus.js';

const createApp = (container: AppContainer): Application => {
  const app = express();

  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));
  app.use(metricsMiddleware);
  app.use("/api/v1/notification", setupNotificationRoutes(container.notificationController, container.authMiddleware));


  setupHealthRoutes(app);

  app.use(globalErrorHandler);

  return app;
};

const setupHealthRoutes = (app: Application) => {
  app.get("/api/v1/notification/metrics", metricsHandler);
  app.get('/api/v1/notification/health', (_req, res) => {
    const isDatabaseConnected = mongoose.connection.readyState === 1;

    if (isDatabaseConnected) {
      return res.status(200).json({
        status: 'UP',
        service: 'notification-service',
        database: 'connected',
      });
    }

    return res.status(503).json({
      status: 'DOWN',
      service: 'notification-service',
      database: 'disconnected',
    });
  });
};

export default createApp;