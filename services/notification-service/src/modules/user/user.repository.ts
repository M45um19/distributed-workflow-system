// 1. Internal modules (Alphabetical Order)
import { IUser, IUserDocument, IUserRepository } from './user.interface.js';
import { User } from './user.model.js';

function isConnectionError(err: any): boolean {
  if (!err) return false;
  const name = err.name || '';
  const message = err.message || '';
  return (
    name === 'MongoNetworkError' ||
    name === 'MongoServerSelectionError' ||
    name === 'MongoTimeoutError' ||
    message.toLowerCase().includes('connection') ||
    message.toLowerCase().includes('timeout') ||
    message.toLowerCase().includes('topology')
  );
}

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

  public async bulkSaveSnapshots(dataList: Partial<IUser>[]): Promise<Partial<IUser>[]> {
    if (dataList.length === 0) return [];

    try {
      const ops = dataList.map(data => ({
        updateOne: {
          filter: { _id: data._id },
          update: { $set: data },
          upsert: true,
        }
      }));
      // Using ordered: true to catch the first write failure and split recursively
      await User.bulkWrite(ops, { ordered: true });
      return [];
    } catch (error: any) {
      if (isConnectionError(error)) {
        throw error;
      }

      // Base case: only 1 document left in the sub-batch and it failed, treat as poisonous data
      if (dataList.length === 1) {
        console.error(`[UserRepository] Poisonous user snapshot isolated: ${JSON.stringify(dataList[0])}. Error: ${error.message || error}`);
        return dataList;
      }

      // Recursive case: split sub-batch in half
      const mid = Math.floor(dataList.length / 2);
      const left = dataList.slice(0, mid);
      const right = dataList.slice(mid);

      const leftFailed = await this.bulkSaveSnapshots(left);
      const rightFailed = await this.bulkSaveSnapshots(right);

      return [...leftFailed, ...rightFailed];
    }
  }

  public async findById(id: string): Promise<IUserDocument | null> {
    const result = await User.findById(id).exec();
    return result;
  }
}