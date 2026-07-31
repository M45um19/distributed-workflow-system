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
  private buffer: { messageValue: string | null; resolveOffset: () => Promise<void> }[] = [];
  private flushTimeout: NodeJS.Timeout | null = null;
  private readonly maxBatchSize = 1000;
  private readonly maxWaitMs = 2000;

  constructor(private readonly notificationService: NotificationService) { }

  public async handle(messageValue: string | null): Promise<void> {
    await this.handleBatchMessage(messageValue, async () => {});
  }

  public async handleBatchMessage(messageValue: string | null, resolveOffset: () => Promise<void>): Promise<void> {
    this.buffer.push({ messageValue, resolveOffset });

    if (this.buffer.length >= this.maxBatchSize) {
      await this.flush();
    } else if (!this.flushTimeout) {
      this.flushTimeout = setTimeout(() => {
        void this.flush();
      }, this.maxWaitMs);
    }
  }

  private async flush(): Promise<void> {
    if (this.flushTimeout) {
      clearTimeout(this.flushTimeout);
      this.flushTimeout = null;
    }

    if (this.buffer.length === 0) return;

    const currentBatch = [...this.buffer];
    this.buffer = [];

    const messageValues = currentBatch.map(item => item.messageValue);
    try {
      await this.handleBatch(messageValues);

      // Commit offsets sequentially
      for (const item of currentBatch) {
        await item.resolveOffset();
      }
      console.warn(`[NotificationHandler] Flushed and committed batch of ${currentBatch.length} notifications.`);
    } catch (err) {
      console.error(`[NotificationHandler] Failed to process batch or commit offsets: ${err}`);
    }
  }

  public async handleBatch(messageValues: (string | null)[]): Promise<void> {
    const validInAppNotifications: INotification[] = [];
    const validEmailNotifications: INotificationEventPayload[] = [];

    for (const messageValue of messageValues) {
      if (!messageValue) continue;
      try {
        const rawData = JSON.parse(messageValue);
        const payload = rawData as INotificationEventPayload;

        console.warn(`[Kafka Trigger] Processing notification event for channel: ${payload.channel}`);

        if (payload.channel === 'IN_APP' || payload.channel === 'BOTH') {
          if (!payload.userId || !payload.title || !payload.message) {
            console.error(`[NotificationHandler] Validation Failed for IN_APP. Payload: ${messageValue}`);
            continue;
          }

          validInAppNotifications.push({
            userId: payload.userId,
            title: payload.title,
            message: payload.message,
            type: payload.type || 'INFO',
            isRead: false,
          });
        }

        if (payload.channel === 'EMAIL' || payload.channel === 'BOTH') {
          if (!payload.email || !payload.emailBody) {
            console.error(`[NotificationHandler] Validation Failed for EMAIL. Payload: ${messageValue}`);
            continue;
          }
          validEmailNotifications.push(payload);
        }
      } catch (error) {
        const errorMessage = error instanceof Error ? error.message : 'Unknown Kafka Notification Parse Error';
        console.error(`[Critical][NotificationHandler] Failed to parse event. Error: ${errorMessage}`);
        console.error(`[NotificationHandler] Faulty Payload: ${messageValue}`);
      }
    }

    // 1. Bulk DB write and batch-dispatch
    if (validInAppNotifications.length > 0) {
      try {
        await this.notificationService.sendNotificationsBulk(validInAppNotifications);
      } catch (error) {
        console.error(`[NotificationHandler] Bulk processing failed: ${error}`);
      }
    }

    // 2. Email processing
    if (validEmailNotifications.length > 0) {
      for (const emailPayload of validEmailNotifications) {
        console.info(`Email channel detected for: ${emailPayload.email}. Triggering email service...`);
      }
    }
  }
}