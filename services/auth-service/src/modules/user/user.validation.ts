import { z } from 'zod';

export const UserRole = z.enum(['ADMIN', 'USER']);
export type UserRole = z.infer<typeof UserRole>;

export const UserSchema = z.object({
  full_name: z.string().min(3).max(50),
  email: z.string().email(),
  password_hash: z.string().min(8),
  role: UserRole.optional().default('USER'),
  address: z.string().optional(),
  phone: z.string().optional(),
  bio: z.string().optional(),
  city: z.string().optional(),
  country: z.string().optional(),
  avatar_url: z.string().url().optional().default(''),
  is_active: z.boolean().optional().default(true),
  created_at: z.date().optional(),
});

export type IUser = z.infer<typeof UserSchema>;

export const updateProfileSchema = z.object({
  body: z.object({
    full_name: z.string().min(3, "Name is too short").max(50).optional(),
    avatar_url: z.string().url("Invalid URL format").optional().or(z.literal('')),
    address: z.string().optional(),
    phone: z.string().optional(),
    bio: z.string().optional(),
    city: z.string().optional(),
    country: z.string().optional(),
  }),
});

export type UpdateProfileRequest = z.infer<typeof updateProfileSchema>;
export type UpdateProfileDTO = UpdateProfileRequest['body'];