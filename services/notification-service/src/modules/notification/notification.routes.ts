import { Router } from 'express';

import { AuthMiddleware } from '../../middleware/auth.middleware.js';

import { NotificationController } from './notification.controller.js';

export const setupNotificationRoutes = (
  notificationCtrl: NotificationController,
  authMiddleware: AuthMiddleware
): Router => {
  const router = Router();

  router.get('/', authMiddleware.protect, notificationCtrl.getNotifications);
  router.patch('/read', authMiddleware.protect, notificationCtrl.markNotificationsAsRead);

  return router;
};