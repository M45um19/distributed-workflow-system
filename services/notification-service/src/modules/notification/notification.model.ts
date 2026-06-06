import { Schema, model } from 'mongoose';

import { INotificationDocument } from './notification.interface.js';

const NotificationSchema = new Schema<INotificationDocument>(
  {
    userId: { type: String, required: true, index: true },
    title: { type: String, required: true },
    message: { type: String, required: true },
    type: {
      type: String,
      enum: ['INFO', 'SUCCESS', 'WARN', 'ERROR'],
      default: 'INFO',
    },
    isRead: { type: Boolean, default: false },
  },
  {
    timestamps: true,
  }
);

export const NotificationModel = model<INotificationDocument>('Notification', NotificationSchema);