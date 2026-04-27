import { z } from 'zod';

export const UserSchema = z.object({
  full_name: z.string().min(3).max(50),
  email: z.string().email(),
  password_hash: z.string().min(8),
  avatar_url: z.string().url().optional().default(''),
});

export type IUser = z.infer<typeof UserSchema>;