import { socketConfig } from '../../config/socket.js';

import { IFetchNotificationsResponse, INotification, INotificationRepository } from './notification.interface.js';

export class NotificationService {
  constructor(private readonly notificationRepository: INotificationRepository) { }

  public async sendNotification(data: INotification): Promise<INotification> {
    const notification = await this.notificationRepository.create(data);

    const io = socketConfig.getIO();
    io.to(data.userId).emit('notification-received', notification);

    console.warn(`[NotificationService] Live notification pushed to user: ${data.userId}`);
    return notification;
  }

  public async getUserNotifications(userId: string): Promise<IFetchNotificationsResponse> {
    if (!userId) {
      throw new Error('User ID is required to fetch notifications');
    }
    return this.notificationRepository.fetchLatest(userId, 20);
  }

  public async updateNotificationsAsRead(userId: string, notificationIds: string[]): Promise<void> {
    if (!userId || !notificationIds || notificationIds.length === 0) {
      throw new Error('Invalid payload for marking notifications as read');
    }
    return this.notificationRepository.markAsRead(userId, notificationIds);
  }
}