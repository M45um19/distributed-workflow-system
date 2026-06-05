// 1. Internal modules (Alphabetical Order)
import { IUser, IUserDocument, IUserRepository } from './user.interface.js';
import { User } from './user.model.js';

export class UserRepository implements IUserRepository {
  
  public async saveSnapshot(data: Partial<IUser>): Promise<IUserDocument> {
    const query = { _id: data._id };
    
    const result = await User.findOneAndUpdate(
      query, 
      { $set: data },
      { upsert: true, new: true }
    ).exec();

    return result as IUserDocument;
  }

  public async findById(id: string): Promise<IUserDocument | null> {
    const result = await User.findById(id).exec();
    return result;
  }
}