// 1. Internal modules
import { IUser, IUserRepository } from './user.interface.js';

export class UserService {
  constructor(private userRepository: IUserRepository) {}

  public async syncUserSnapshot(eventData: { id: string; email: string; full_name: string; role?: string }): Promise<void> {
    const userSnapshot: Partial<IUser> = {
      _id: eventData.id,
      full_name: eventData.full_name,
      email: eventData.email,
      role: eventData.role || 'user',
    };

    const result = await this.userRepository.saveSnapshot(userSnapshot);
    console.info(`[Snapshot Synced] Successfully saved snapshot for: ${result.email} (${result.id})`);
  }
}