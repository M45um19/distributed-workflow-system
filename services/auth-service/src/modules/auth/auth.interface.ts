import { LoginUserDTO, RegisterUserDTO } from "./auth.validation";
import { Request } from 'express';

export interface IAuthService {
  register(data: RegisterUserDTO): Promise<any>;
  login(data: LoginUserDTO, deviceMeta: { deviceId: string; ip: any; deviceName: string }): Promise<any>;
  verifySession(token: string): Promise<any>
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