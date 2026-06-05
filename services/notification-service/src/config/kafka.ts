import { Kafka, Consumer, logLevel } from 'kafkajs';

import { env } from './env.js';

export interface IKafkaConsumer {
  connect(): Promise<void>;
  subscribe(topic: string): Promise<void>;
  disconnect(): Promise<void>;
}

class KafkaConfig {
  private kafka: Kafka;
  private consumer: Consumer;
  private isConnected: boolean;

  constructor() {
    this.isConnected = false;
    const brokers = env.KAFKA_BROKERS ? env.KAFKA_BROKERS.split(',') : [];

    if (brokers.length === 0) {
      throw new Error('KAFKA_BROKERS are not defined in the environment variables.');
    }

    this.kafka = new Kafka({
      clientId: 'notification-service',
      brokers: brokers,
      logLevel: logLevel.ERROR,
    });

    this.consumer = this.kafka.consumer({
      groupId: 'notification-service-group',
    });
  }

  getConsumer(): Consumer {
    return this.consumer;
  }

  async connect(): Promise<void> {
    if (this.isConnected) return;
    try {
      await this.consumer.connect();
      this.isConnected = true;
      console.info('Kafka Consumer connected successfully');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown Kafka error';
      console.error('Kafka Consumer connection failed:', errorMessage);
      throw new Error(`Kafka connection could not be established: ${errorMessage}`);
    }
  }

  async disconnect(): Promise<void> {
    try {
      await this.consumer.disconnect();
      this.isConnected = false;
      console.info('Kafka Consumer disconnected');
    } catch (error) {
      console.error('Kafka Consumer disconnect error:', error);
    }
  }
}

export const kafkaConfig = new KafkaConfig();