import { Producer } from 'kafkajs';
import { UserService } from '../../modules/user/user.service.js';
import { IKafkaHandler } from '../worker.js';

interface IUserRegisteredEvent {
  id: string;
  email: string;
  full_name: string;
  role?: string;
  avatar_url?: string;
}

export class UserRegisteredHandler implements IKafkaHandler {
  private buffer: { messageValue: string | null; resolveOffset: () => Promise<void> }[] = [];
  private flushTimeout: NodeJS.Timeout | null = null;
  private readonly maxBatchSize = 100;
  private readonly maxWaitMs = 2000;

  constructor(
    private readonly userService: UserService,
    private readonly producer: Producer
  ) {}

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

  private async sendToDLQ(messageValue: string | null, reason: string): Promise<void> {
    if (!messageValue) return;
    try {
      await this.producer.send({
        topic: 'user-registered-notification-dlq',
        messages: [{
          value: messageValue,
          headers: {
            'x-failure-reason': reason
          }
        }]
      });
      console.warn(`[UserRegisteredHandler] Poisonous user registration message isolated to DLQ: ${messageValue}`);
    } catch (err) {
      console.error(`[UserRegisteredHandler] Failed to publish message to DLQ: ${err}`);
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

    // Parse and validate events, keeping track of their mappings
    const parsedItems: { item: typeof currentBatch[0]; payload: IUserRegisteredEvent }[] = [];
    const processedItems: typeof currentBatch = [];

    for (const item of currentBatch) {
      if (!item.messageValue) {
        processedItems.push(item);
        continue;
      }
      try {
        const payload = JSON.parse(item.messageValue) as IUserRegisteredEvent;
        if (!payload.id || !payload.email || !payload.full_name) {
          const reason = `Validation Failed: id=${payload.id}, email=${payload.email}, full_name=${payload.full_name}`;
          console.error(`[UserRegisteredHandler] ${reason}. Payload: ${item.messageValue}`);
          await this.sendToDLQ(item.messageValue, reason);
          processedItems.push(item);
          continue;
        }
        parsedItems.push({ item, payload });
      } catch (error) {
        const reason = `JSON Parse Error: ${(error as Error).message}`;
        console.error(`[UserRegisteredHandler] ${reason}. Payload: ${item.messageValue}`);
        await this.sendToDLQ(item.messageValue, reason);
        processedItems.push(item);
      }
    }

    try {
      if (parsedItems.length > 0) {
        const payloads = parsedItems.map(p => p.payload);
        const failedEvents = await this.userService.syncUserSnapshotsBulk(payloads);

        const failedIds = new Set(failedEvents.map(e => e.id));

        for (const { item, payload } of parsedItems) {
          if (failedIds.has(payload.id)) {
            const reason = "Database Bulk Upsert Failure (Poisonous Data)";
            await this.sendToDLQ(item.messageValue, reason);
          }
          processedItems.push(item);
        }
      }

      // Commit offsets sequentially for successfully saved or DLQ-isolated messages
      for (const item of processedItems) {
        await item.resolveOffset();
      }
      console.warn(`[UserRegisteredHandler] Flushed and committed batch of ${currentBatch.length} user registrations.`);
    } catch (err) {
      // Re-throw database connection error to prevent committing unprocessed offsets
      console.error(`[UserRegisteredHandler] Critical failure in bulk sync: ${err}`);
      throw err;
    }
  }
}