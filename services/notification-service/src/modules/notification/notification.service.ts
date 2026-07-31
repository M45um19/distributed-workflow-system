import { socketConfig } from '../../config/socket.js';

import { 
  IFetchNotificationsResponse, 
  INotification, 
  INotificationDocument,
  INotificationRepository 
} from './notification.interface.js';

export class NotificationService {
  constructor(private readonly notificationRepository: INotificationRepository) { }

  public async sendNotification(data: INotification): Promise<INotification> {
    const notification = await this.notificationRepository.create(data);

    const io = socketConfig.getIO();
    io.to(`user_${data.userId}`).emit('notification-received', notification);

    console.warn(`[NotificationService] Live notification pushed to user: ${data.userId}`);
    return notification;
  }

  public async sendNotificationsBulk(dataList: INotification[]): Promise<INotificationDocument[]> {
    const notifications = await this.notificationRepository.createMany(dataList);

    try {
      const io = socketConfig.getIO();
      const adapter = io.of('/').adapter as any;

      if (adapter && adapter.pubClient) {
        const pubClient = adapter.pubClient;
        const parser = adapter.parser;
        const channelPrefix = adapter.channel;
        const uid = adapter.uid;

        const pipeline = pubClient.pipeline();

        for (const n of notifications) {
          const packet = {
            type: 2,
            data: ['notification-received', n],
            nsp: '/'
          };
          const rawOpts = {
            rooms: [`user_${n.userId}`],
            except: [],
            flags: {}
          };

          const msg = parser.encode([uid, packet, rawOpts]);
          const channel = `${channelPrefix}user_${n.userId}#`;

          pipeline.publish(channel, msg);
        }

        await pipeline.exec();
        console.warn(`[NotificationService] Bulk live notifications pushed: ${notifications.length} events via Redis pipeline.`);
      } else {
        const io = socketConfig.getIO();
        for (const n of notifications) {
          io.to(`user_${n.userId}`).emit('notification-received', n);
        }
        console.warn(`[NotificationService] Bulk live notifications pushed: ${notifications.length} events (socket.io fallback).`);
      }
    } catch (redisError) {
      console.error(`[NotificationService] Redis pipelined broadcast failed: ${redisError}`);
      const io = socketConfig.getIO();
      for (const n of notifications) {
        io.to(`user_${n.userId}`).emit('notification-received', n);
      }
    }

    return notifications;
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