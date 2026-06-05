import { Kafka, Producer, Partitioners, Message } from 'kafkajs';

import { env } from './env.js';

export interface IKafkaProducer {
    connect(): Promise<void>;
    sendMessage<T>(topic: string, message: T): Promise<void>;
    disconnect(): Promise<void>;
}

class KafkaConfig implements IKafkaProducer {
    private kafka: Kafka;
    private producer: Producer;
    private isConnected: boolean;

    constructor() {
        this.isConnected = false;
        const brokers = env.KAFKA_BROKERS ? env.KAFKA_BROKERS.split(',') : [];

        if (brokers.length === 0) {
            throw new Error('KAFKA_BROKERS are not defined in the environment variables.');
        }

        this.kafka = new Kafka({
            clientId: 'auth-service',
            brokers: brokers,
            retry: {
                initialRetryTime: 100,
                retries: 8
            }
        });

        this.producer = this.kafka.producer({
            createPartitioner: Partitioners.LegacyPartitioner,
            allowAutoTopicCreation: true,
        });
    }

    async connect(): Promise<void> {
        if (this.isConnected) return;

        try {
            await this.producer.connect();
            this.isConnected = true;
            console.info('Kafka Producer connected successfully');
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown Kafka error';
            console.error('Kafka Connection failed:', errorMessage);
            throw new Error(`Kafka connection could not be established: ${errorMessage}`);
        }
    }

    async sendMessage<T>(topic: string, message: T): Promise<void> {
        if (!this.isConnected) {
            await this.connect();
        }

        try {
            const kafkaMessage: Message = {
                value: JSON.stringify(message),
                timestamp: Date.now().toString(),
            };

            await this.producer.send({
                topic,
                messages: [kafkaMessage],
                acks: -1,
            });

        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Unknown error while sending message';
            console.error(`Kafka Send error in topic ${topic}:`, errorMessage);
            throw new Error(errorMessage);
        }
    }

    async disconnect(): Promise<void> {
        try {
            await this.producer.disconnect();
            this.isConnected = false;
            console.info('Kafka Producer disconnected');
        } catch (error: unknown) {
            const errorMessage = error instanceof Error ? error.message : 'Error disconnecting Kafka';
            console.error('Kafka Disconnect error:', errorMessage);
        }
    }
}

export const kafkaConfig = new KafkaConfig();