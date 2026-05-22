import { Document, Types } from 'mongoose';

import { IUser } from './user.validation';

export interface IUserDocument extends IUser, Document {
  _id: Types.ObjectId;
}

export interface IUserRepository {
  exists(email: string): Promise<boolean>;
  create(data: IUser): Promise<IUserDocument>;
  findByEmail(email: string): Promise<IUserDocument | null>;
  findById(id: string): Promise<IUserDocument | null>;
}