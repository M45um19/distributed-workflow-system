import { IUserDocument, IUserRepository } from './user.interface';
import { User } from './user.model';
import { IUser } from './user.validation';

export class UserRepository implements IUserRepository {
  async exists(email: string): Promise<boolean> {
    const user = await User.exists({ email });
    return !!user;
  }

  create(data: IUser): Promise<IUserDocument> {
    return User.create(data);
  }

  async findByEmail(email: string): Promise<IUserDocument | null> {
    return User.findOne({ email }).select('+password_hash').exec();
  }

  findById(id: string): Promise<IUserDocument | null> {
    return User.findById(id).exec();
  }
}