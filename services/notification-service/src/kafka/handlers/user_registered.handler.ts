import { UserService } from '../../modules/user/user.service.js';
import { IKafkaHandler } from '../worker.js';

interface IUserRegisteredEvent {
  id: string;
  email: string;
  full_name: string;
  role?: string;
  avatar_url?: string
}

export class UserRegisteredHandler implements IKafkaHandler {
  constructor(private readonly userService: UserService) {}

  public async handle(messageValue: string | null): Promise<void> {
    if (!messageValue) return;
    
    try {
      const payload = JSON.parse(messageValue) as IUserRegisteredEvent;
      
      await this.userService.syncUserSnapshot(payload);
    } catch (error) {
      throw new Error(`[UserRegisteredHandler] Error: ${(error as Error).message}`);
    }
  }
}