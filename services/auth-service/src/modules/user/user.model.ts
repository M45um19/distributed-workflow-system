import mongoose from 'mongoose';

import { IUserDocument } from './user.interface.js';
import { IUser, UserRole } from './user.validation.js';

const userSchema = new mongoose.Schema<IUserDocument>({
  _id: {
    type: String,
    required: true,
  },
  full_name: {
    type: String,
    required: true,
    trim: true,
  },
  email: {
    type: String,
    required: true,
    unique: true,
    lowercase: true,
    trim: true,
  },
  password_hash: {
    type: String,
    required: true,
    select: false,
  },
  role: {
    type: String,
    enum: Object.values(UserRole.enum)
  },
  avatar_url: {
    type: String,
    default: '',
  },
  address: {
    type: String,
    default: '',
  },
  phone: {
    type: String,
    default: '',
  },
  bio: {
    type: String,
    default: '',
  },
  city: {
    type: String,
    default: '',
  },
  country: {
    type: String,
    default: '',
  },
}, {
  timestamps: { createdAt: 'created_at', updatedAt: 'updated_at' },
  versionKey: false
});


export const User = mongoose.model<IUserDocument>('User', userSchema);
 