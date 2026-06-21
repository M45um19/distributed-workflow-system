import { redisService } from '../../config/redis.js';
import { IKafkaHandler } from '../worker.js';

interface IUserLogoutPayload {
    userId: string;
    deviceId: string;
}

export class UserLogoutHandler implements IKafkaHandler {

    public async handle(messageValue: string | null): Promise<void> {
        if (!messageValue) return;

        try {
            const payload = JSON.parse(messageValue) as IUserLogoutPayload;

            const sessionData = await redisService.get(`session:${payload.userId}:${payload.deviceId}`);
            if (sessionData) {
                await redisService.del(`session:${payload.userId}:${payload.deviceId}`);
            }
        } catch (error) {
            console.error(`[UserLogoutHandler] Failed to process message. Error: ${(error as Error).message}`);
            console.error(`[UserLogoutHandler] Faulty Payload: ${messageValue}`);
        }
    }
}