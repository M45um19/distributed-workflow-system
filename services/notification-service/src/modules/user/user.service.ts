// 1. Internal modules
import { IUser, IUserRepository } from './user.interface.js';

interface IUserRegisteredEvent {
  id: string;
  email: string;
  full_name: string;
  role?: string;
  avatar_url?: string;
}

export class UserService {
  constructor(private userRepository: IUserRepository) {}

  public async syncUserSnapshot(eventData: { id: string; email: string; full_name: string; role?: string, avatar_url?: string }): Promise<void> {
    const userSnapshot: Partial<IUser> = {
      _id: eventData.id,
      full_name: eventData.full_name,
      email: eventData.email,
      role: eventData.role || 'USER',
      avatar_url: eventData.avatar_url || ""
    };

    const result = await this.userRepository.saveSnapshot(userSnapshot);
    console.info(`[Snapshot Synced] Successfully saved snapshot for: ${result.email} (${result.id})`);
  }

  public async syncUserSnapshotsBulk(events: IUserRegisteredEvent[]): Promise<IUserRegisteredEvent[]> {
    if (events.length === 0) return [];

    const snapshots: Partial<IUser>[] = events.map(e => ({
      _id: e.id,
      full_name: e.full_name,
      email: e.email,
      role: e.role || 'USER',
      avatar_url: e.avatar_url || ""
    }));

    const failedSnapshots = await this.userRepository.bulkSaveSnapshots(snapshots);

    if (failedSnapshots.length === 0) {
      console.info(`[Bulk Snapshots Synced] Successfully saved all ${snapshots.length} user snapshots.`);
      return [];
    }

    const failedIds = new Set(failedSnapshots.map(s => s._id));
    const failedEvents = events.filter(e => failedIds.has(e.id));
    console.warn(`[Bulk Snapshots Synced] Completed. Successfully saved: ${snapshots.length - failedEvents.length}. Failed/Isolated: ${failedEvents.length}`);
    return failedEvents;
  }
}