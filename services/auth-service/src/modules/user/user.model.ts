import mongoose from 'mongoose';

import { IUser, UserRole } from './user.validation.js';

const userSchema = new mongoose.Schema<IUser>({
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
}, {
  timestamps: { createdAt: 'created_at', updatedAt: 'updated_at' },
  versionKey: false
});


export const User = mongoose.model<IUser>('User', userSchema);
 