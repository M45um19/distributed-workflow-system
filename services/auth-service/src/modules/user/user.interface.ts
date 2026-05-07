import { Document, Types } from 'mongoose';

export interface IUser {
  full_name: string;
  email: string;
  password_hash: string;
  role?: string;
  is_active?: boolean;
  created_at?: Date
}

export interface IUserDocument extends IUser, Document {
  _id: Types.ObjectId;
}

export interface IUserRepository {
  exists(email: string): Promise<boolean>;
  create(data: IUser): Promise<IUserDocument>;
  findByEmail(email: string): Promise<IUserDocument | null>;
  findById(id: string): Promise<IUserDocument | null>;
}