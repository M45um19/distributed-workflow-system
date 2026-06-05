// 1. External modules
import mongoose, { Schema, Model } from 'mongoose';

// 2. Internal modules
import { IUserDocument } from './user.interface.js';

const userSnapshotSchema = new Schema<IUserDocument>(
  {
    _id: { type: String, required: true },
    email: { type: String, required: true, unique: true, lowercase: true, trim: true },
    full_name: { type: String, required: true, trim: true },
    role: { type: String, required: true, default: 'user' },
    avatar_url: { type: String, default: '' },
    is_active: { type: Boolean, default: true },
  },
  {
    timestamps: { createdAt: 'created_at', updatedAt: 'updated_at' },
    versionKey: false,
  }
);

export const User: Model<IUserDocument> = mongoose.models.UserSnapshot || 
  mongoose.model<IUserDocument>('UserSnapshot', userSnapshotSchema);