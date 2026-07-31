import { v7 as uuidv7 } from 'uuid';

import { 
  IFetchNotificationsResponse, 
  INotification, 
  INotificationDocument,
  INotificationRepository 
} from './notification.interface.js';
import { NotificationModel } from './notification.model.js';

export class NotificationRepository implements INotificationRepository {

  public async create(data: INotification): Promise<INotification> {
    const id = uuidv7();
    const created = await NotificationModel.create({ _id: id, ...data });
    return created.toJSON() as INotification;
  }

  public async createMany(data: INotification[]): Promise<INotificationDocument[]> {
    const docs = data.map(item => ({
      _id: uuidv7(),
      ...item,
    }));
    const created = await NotificationModel.insertMany(docs);
    return created as unknown as INotificationDocument[];
  }

  public async fetchLatest(userId: string, limit: number): Promise<IFetchNotificationsResponse> {
    const finalLimit = limit || 20;
    const [notifications, unreadCount] = await Promise.all([
      NotificationModel
        .find({ userId })
        .sort({ createdAt: -1 })
        .limit(finalLimit)
        .exec(),
      NotificationModel
        .countDocuments({ userId, isRead: false })
        .exec()
    ]);

    return {
      notifications,
      unreadCount,
    };
  }

  public async markAsRead(userId: string, notificationIds: string[]): Promise<void> {
    await NotificationModel
      .updateMany(
        {
          userId,
          _id: { $in: notificationIds }
        },
        {
          $set: { isRead: true }
        }
      )
      .exec();
  }
}