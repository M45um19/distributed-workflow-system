
import { Request } from 'express';
import { JwtPayload } from 'jsonwebtoken';

import { LoginUserDTO, RegisterUserDTO } from "./auth.validation.js";


export interface AuthResponse {
  accessToken?: string;
  refreshToken?: string;
  user: {
    id: string;
    full_name: string;
    email: string;
    avatar_url: string;
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

export interface logOutUserInput{
  userId: string;
  deviceId: string;
}
export interface IAuthService {
  register(data: RegisterUserDTO, deviceMeta: DeviceMeta): Promise<AuthResponse>;
  login(data: LoginUserDTO, deviceMeta: DeviceMeta): Promise<AuthResponse>; verifySession(token: string): Promise<SessionVerification>;
  logout(data: logOutUserInput): Promise<void>
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
  user?: AccessTokenPayload;
}

export interface AccessTokenPayload extends JwtPayload {
  userId: string;
  deviceId: string;
}