import { Response } from 'express';

import { AuthRequest } from '../../middleware/auth.middleware.js';
import { sendResponse } from '../../utils/sendResponse.js';

import { NotificationService } from './notification.service.js';



export class NotificationController {
    constructor(private readonly notificationService: NotificationService) { }

    public getNotifications = async (req: AuthRequest, res: Response): Promise<void> => {
        try {

            const userId = req.user?.userId;

            if (!userId) {
                res.status(401).json({ success: false, message: 'Unauthorized access' });
                return;
            }

            const data = await this.notificationService.getUserNotifications(userId);

            sendResponse(res, {
                statusCode: 200,
                success: true,
                message: 'Notifications fetched successfully',
                data: data.notifications,
                meta: {
                    total: data.unreadCount,
                },
            });
        } catch (error) {
            console.error(`[NotificationController] getNotifications Error: ${(error as Error).message}`);
            res.status(500).json({ success: false, message: 'Internal server error' });
        }
    };

    public markNotificationsAsRead = async (req: AuthRequest, res: Response): Promise<void> => {
        try {

            const userId = req.user?.userId;
            const { notificationIds } = req.body as { notificationIds: string[] };

            if (!userId) {
                res.status(401).json({ success: false, message: 'Unauthorized access' });
                return;
            }

            if (!notificationIds || !Array.isArray(notificationIds) || notificationIds.length === 0) {
                res.status(400).json({ success: false, message: 'Invalid or empty notificationIds array' });
                return;
            }

            await this.notificationService.updateNotificationsAsRead(userId, notificationIds);

            sendResponse(res, {
                statusCode: 200,
                success: true,
                message: 'Notifications successfully marked as read',
                data: null,
            });
        } catch (error) {
            console.error(`[NotificationController] markNotificationsAsRead Error: ${(error as Error).message}`);
            res.status(500).json({ success: false, message: 'Internal server error' });
        }
    };
}