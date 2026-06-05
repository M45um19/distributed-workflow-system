import { Document } from 'mongoose';

export interface IUser {
  _id: string;
  email: string;
  full_name: string;
  role: string;
  avatar_url?: string;
  is_active?: boolean;
}

export interface IUserDocument extends Omit<Document, '_id'>, IUser {
  _id: string; 
  created_at: Date;
  updated_at: Date;
}

export interface IUserRepository {
  saveSnapshot(data: Partial<IUser>): Promise<IUserDocument>;
  findById(id: string): Promise<IUserDocument | null>;
}