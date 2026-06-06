import { INotification, INotificationRepository } from './notification.interface.js';
import { NotificationModel } from './notification.model.js';

export class NotificationRepository implements INotificationRepository {
  public async create(data: INotification): Promise<INotification> {
    const created = await NotificationModel.create(data);
    return created.toJSON() as INotification;
  }
}