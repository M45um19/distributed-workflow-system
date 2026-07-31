import { kafkaConfig } from '../config/kafka.js';

export interface IKafkaHandler {
  handle(message: string | null): Promise<void>;
}

export class KafkaWorker {
  private readonly handlers: Map<string, IKafkaHandler> = new Map<string, IKafkaHandler>();

  public addTopicHandler(topic: string, handler: IKafkaHandler): void {
    this.handlers.set(topic, handler);
  }

  public async start(): Promise<void> {
    const consumer = kafkaConfig.getConsumer();
    const topics = Array.from(this.handlers.keys());

    if (topics.length === 0) {
      console.warn('[KafkaWorker] No topics registered to consume.');
      return;
    }

    for (const topic of topics) {
      await consumer.subscribe({ topic, fromBeginning: true });
    }

    await consumer.run({
      eachMessage: async ({ topic, partition, message }) => {
        const value = message.value ? message.value.toString() : null;
        const handler = this.handlers.get(topic);
        if (!handler) return;

        // Delegate batch buffering or process immediately
        if (typeof (handler as any).handleBatchMessage === 'function') {
          const offset = message.offset;
          const resolve = async () => {
            const nextOffset = (BigInt(offset) + 1n).toString();
            await consumer.commitOffsets([
              { topic, partition, offset: nextOffset }
            ]);
          };
          await (handler as any).handleBatchMessage(value, resolve);
        } else {
          // Standard one-by-one handler
          await handler.handle(value);
        }
      }
    });
  }
}