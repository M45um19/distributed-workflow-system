import { IUserRepository } from './user.interface';
import { User } from './user.model';

export class UserRepository implements IUserRepository {
  async create(userData: any): Promise<any> {
    const user = new User(userData);
    return await user.save();
  }

  async findByEmail(email: string): Promise<any> {
    return await User.findOne({ email }).select('+password_hash');
  }

  async exists(email: string): Promise<boolean | any> {
    return await User.exists({ email });
  }

  async findById(id: string): Promise<any> {
    return await User.findById(id);
  }
}