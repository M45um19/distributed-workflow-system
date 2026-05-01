import mongoose, { Schema, Document } from 'mongoose';
import { IWorkspace } from './workpace.validation';

const workspaceSchema = new mongoose.Schema<IWorkspace>({
  name: { type: String, required: true, trim: true },
  slug: { type: String, required: true, unique: true, lowercase: true },
  owner_id: { type: Schema.Types.ObjectId, ref: 'User', required: true },
}, { 
  timestamps: true
});

export const Workspace = mongoose.model<IWorkspace>('Workspace', workspaceSchema);