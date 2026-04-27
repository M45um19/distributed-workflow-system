import { z } from 'zod';

export const registerSchema = z.object({
  body: z.object({
    full_name: z.string().min(3, "Name is too short").max(50),
    email: z.string().email("Invalid email address"),
    password: z.string().min(8, "Password must be at least 8 characters"),
    avatar_url: z.string().url().optional(),
  }),
});

export const loginSchema = z.object({
  body: z.object({
    email: z.string().email("Invalid email address"),
    password: z.string().min(1, "Password is required"),
  }),
});

export type RegisterUserRequest = z.infer<typeof registerSchema>;
export type LoginUserRequest = z.infer<typeof loginSchema>;

export type RegisterUserDTO = RegisterUserRequest['body'];
export type LoginUserDTO = LoginUserRequest['body'];