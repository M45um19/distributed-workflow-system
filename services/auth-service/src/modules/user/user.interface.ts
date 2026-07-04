import { Document, Types } from 'mongoose';

import { IUser } from './user.validation.js';

export interface IUserDocument extends IUser, Document {
  _id: Types.ObjectId;
}

export interface IUserRepository {
  exists(email: string): Promise<boolean>;
  create(data: IUser): Promise<IUserDocument>;
  findByEmail(email: string): Promise<IUserDocument | null>;
  findById(id: string): Promise<IUserDocument | null>;
  updateById(id: string, data: Partial<IUser>): Promise<IUserDocument | null>;
}

export interface UserProfileResponse {
  id: string;
  full_name: string;
  email: string;
  role?: string | undefined;
  avatar_url: string;
  address?: string | undefined;
  phone?: string | undefined;
  bio?: string | undefined;
  city?: string | undefined;
  country?: string | undefined;
  created_at?: Date | undefined;
}

export interface IUserService {
  getUserProfile(userId: string): Promise<UserProfileResponse>;
  updateUserProfile(userId: string, data: Partial<IUser>): Promise<UserProfileResponse>;
}