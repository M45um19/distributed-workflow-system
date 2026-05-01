import mongoose from 'mongoose';
import { z } from 'zod';

export const WorkspaceSchema = z.object({
  name: z
    .string("Workspace name is required")
    .min(3, "Name must be at least 3 characters")
    .max(50, "Name cannot exceed 50 characters")
    .trim(),

  slug: z
    .string("Slug is required")
    .min(3, "Slug must be at least 3 characters")
    .toLowerCase()
    .trim()
    .regex(/^[a-z0-9-]+$/, "Slug can only contain lowercase letters, numbers, and hyphens"),

  owner_id: z
    .string("Owner ID is required")
    .refine((val) => mongoose.Types.ObjectId.isValid(val), {
      message: "Invalid MongoDB ObjectId for Owner ID",
    })
    .transform((val) => new mongoose.Types.ObjectId(val)),

  description: z
    .string()
    .max(255, "Description is too long")
    .optional()
    .default(''),
});


export type IWorkspace = z.infer<typeof WorkspaceSchema>;

export const createWorkspaceSchema = z.object({
  body: WorkspaceSchema.omit({ owner_id: true }),
});

export type CreateWorkspaceRequest = z.infer<typeof createWorkspaceSchema>;
export type CreateWorkspaceDTO = CreateWorkspaceRequest['body'];