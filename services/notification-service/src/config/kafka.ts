import { Kafka, Consumer, Producer, logLevel } from 'kafkajs';

import { env } from './env.js';

export interface IKafkaConsumer {
  connect(): Promise<void>;
  subscribe(topic: string): Promise<void>;
  disconnect(): Promise<void>;
}

class KafkaConfig {
  private kafka: Kafka;
  private consumer: Consumer;
  private producer: Producer;
  private isConnected: boolean;
  private isProducerConnected: boolean;

  constructor() {
    this.isConnected = false;
    this.isProducerConnected = false;
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

    this.producer = this.kafka.producer();
  }

  getConsumer(): Consumer {
    return this.consumer;
  }

  getProducer(): Producer {
    return this.producer;
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

  async connectProducer(): Promise<void> {
    if (this.isProducerConnected) return;
    try {
      await this.producer.connect();
      this.isProducerConnected = true;
      console.info('Kafka Producer connected successfully');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown Kafka error';
      console.error('Kafka Producer connection failed:', errorMessage);
      throw new Error(`Kafka producer connection could not be established: ${errorMessage}`);
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

    if (this.isProducerConnected) {
      try {
        await this.producer.disconnect();
        this.isProducerConnected = false;
        console.info('Kafka Producer disconnected');
      } catch (error) {
        console.error('Kafka Producer disconnect error:', error);
      }
    }
  }
}

export const kafkaConfig = new KafkaConfig();