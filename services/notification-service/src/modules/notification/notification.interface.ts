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

export interface INotificationDocument extends INotification, Document {}

export interface INotificationRepository {
  create(data: INotification): Promise<INotification>;
}