import { Request } from 'express';

import { LoginUserDTO, RegisterUserDTO } from "./auth.validation";

export interface AuthResponse {
  accessToken?: string;
  refreshToken?: string;
  user: {
    id: string;
    full_name: string;
    email: string;
  };
}

export interface SessionVerification {
  isValid: boolean;
  userId?: string;
  role?: string;
  email?: string;
  deviceId?: string;
  ip?: string;
  deviceName?: string;
}
export interface DeviceMeta {
  deviceId: string;
  ip: string;
  deviceName: string;
}
export interface IAuthService {
  register(data: RegisterUserDTO, deviceMeta: DeviceMeta): Promise<AuthResponse>;
  login(data: LoginUserDTO, deviceMeta: DeviceMeta): Promise<AuthResponse>; verifySession(token: string): Promise<SessionVerification>;
}

export interface AuthUser {
  userId: string;
  role: string;
  email: string;
  deviceId: string;  
  ip?: string;       
  deviceName?: string; 
}
export interface AuthRequest extends Request {
  user?: AuthUser;
}