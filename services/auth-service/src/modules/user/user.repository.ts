import { v7 as uuidv7 } from 'uuid';

import { IUserDocument, IUserRepository } from './user.interface.js';
import { User } from './user.model.js';
import { IUser } from './user.validation.js';

export class UserRepository implements IUserRepository {
  async exists(email: string): Promise<boolean> {
    const user = await User.exists({ email });
    return !!user;
  }

  create(data: IUser): Promise<IUserDocument> {
    const id = uuidv7();
    return User.create({ _id: id, ...data });
  }

  async findByEmail(email: string): Promise<IUserDocument | null> {
    return User.findOne({ email }).select('+password_hash').exec();
  }

  findById(id: string): Promise<IUserDocument | null> {
    return User.findById(id).exec();
  }

  updateById(id: string, data: Partial<IUser>): Promise<IUserDocument | null> {
    return User.findByIdAndUpdate(id, data, { new: true, runValidators: true }).exec();
  }

  async deleteById(id: string): Promise<boolean> {
    const result = await User.deleteOne({ _id: id }).exec();
    return result.deletedCount > 0;
  }
}