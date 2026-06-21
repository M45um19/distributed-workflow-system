import { INotification } from '../../modules/notification/notification.interface.js';
import { NotificationService } from '../../modules/notification/notification.service.js';
import { IKafkaHandler } from '../worker.js';

interface INotificationEventPayload {
  channel: 'IN_APP' | 'EMAIL' | 'BOTH';

  userId?: string;
  title?: string;
  message?: string;
  type?: 'INFO' | 'SUCCESS' | 'WARN' | 'ERROR';

  email?: string;
  emailSubject?: string;
  emailBody?: string;
}

export class NotificationHandler implements IKafkaHandler {
  constructor(private readonly notificationService: NotificationService) { }

  public async handle(messageValue: string | null): Promise<void> {
    if (!messageValue) {
      console.warn('[NotificationHandler] Received null or empty message.');
      return;
    }

    try {
      const rawData = JSON.parse(messageValue);
      const payload = rawData as INotificationEventPayload;

      console.warn(`[Kafka Trigger] Processing notification event for channel: ${payload.channel}`);

      if (payload.channel === 'IN_APP' || payload.channel === 'BOTH') {
        if (!payload.userId || !payload.title || !payload.message) {
          console.error(`[NotificationHandler] Validation Failed for IN_APP. Payload: ${messageValue}`);
          return;
        }

        const notificationData: INotification = {
          userId: payload.userId,
          title: payload.title,
          message: payload.message,
          type: payload.type || 'INFO',
          isRead: false,
        };

        await this.notificationService.sendNotification(notificationData);
      }

      if (payload.channel === 'EMAIL' || payload.channel === 'BOTH') {
        if (!payload.email || !payload.emailBody) {
          console.error(`[NotificationHandler] Validation Failed for EMAIL. Payload: ${messageValue}`);
          return;
        }

        console.info(`Email channel detected for: ${payload.email}. Triggering email service next...`);

      }

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown Kafka Notification Error';

      console.error(`[Critical][NotificationHandler] Failed to process event. Error: ${errorMessage}`);
      console.error(`[NotificationHandler] Faulty Payload: ${messageValue}`);
    }
  }
}