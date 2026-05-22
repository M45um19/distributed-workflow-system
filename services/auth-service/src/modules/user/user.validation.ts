import { z } from 'zod';

export const UserRole = z.enum(['ADMIN', 'USER']);
export type UserRole = z.infer<typeof UserRole>;

export const UserSchema = z.object({
  full_name: z.string().min(3).max(50),
  email: z.string().email(),
  password_hash: z.string().min(8),
  role: UserRole.optional().default('USER'),
  avatar_url: z.string().url().optional().default(''),
  is_active: z.boolean().optional().default(true),
  created_at: z.date().optional(),
});

export type IUser = z.infer<typeof UserSchema>;