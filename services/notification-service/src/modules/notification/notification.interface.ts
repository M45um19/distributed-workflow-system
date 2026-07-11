import { Document } from 'mongoose';

export interface INotification {
  userId: string;
  title: string;
  message: string;
  type: 'INFO' | 'SUCCESS' | 'WARN' | 'ERROR';
  isRead: boolean;
  createdAt?: Date;
  updatedAt?: Date;
}

export interface INotificationDocument extends INotification, Document<string> {
  _id: string;
}

export interface IFetchNotificationsResponse {
  notifications: INotificationDocument[];
  unreadCount: number;
}

export interface INotificationRepository {
  create(data: INotification): Promise<INotification>;
  fetchLatest(userId: string, limit?: number): Promise<IFetchNotificationsResponse>;
  markAsRead(userId: string, notificationIds: string[]): Promise<void>;
}