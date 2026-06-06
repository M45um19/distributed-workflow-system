import { socketConfig } from '../../config/socket.js';

import { INotification, INotificationRepository } from './notification.interface.js';

export class NotificationService {
  constructor(private readonly notificationRepository: INotificationRepository) {}

  public async sendNotification(data: INotification): Promise<INotification> {
    const notification = await this.notificationRepository.create(data);

    const io = socketConfig.getIO();
    io.to(data.userId).emit('notification-received', notification);

    console.warn(`[NotificationService] Live notification pushed to user: ${data.userId}`);
    return notification;
  }
}