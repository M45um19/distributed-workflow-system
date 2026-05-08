
import bcrypt from 'bcryptjs';
import jwt, { JwtPayload } from 'jsonwebtoken';

import { env } from '../../config/env';
import { kafkaConfig } from '../../config/kafka';
import { redisService } from '../../config/redis';
import { AppError } from '../../utils/appError';
import { IUserRepository } from '../user/user.interface';

import { AuthResponse, IAuthService, SessionVerification } from './auth.interface';
import { LoginUserDTO, RegisterUserDTO } from './auth.validation';


interface AccessTokenPayload extends JwtPayload {
  userId: string;
  deviceId: string;
}

export class AuthService implements IAuthService {
  constructor(private userRepository: IUserRepository) { }

  async register(data: RegisterUserDTO): Promise<AuthResponse["user"]> {
    const { full_name, email, password } = data;

    const isUserExist = await this.userRepository.exists(email);
    if (isUserExist) throw new AppError('User already exists', 400);

    const salt = await bcrypt.genSalt(12);
    const hashedPassword = await bcrypt.hash(password, salt);

    const newUser = await this.userRepository.create({
      full_name,
      email,
      password_hash: hashedPassword
    });

    const userResult = {
      id: newUser._id.toString(),
      name: newUser.full_name,
      email: newUser.email
    };


    await kafkaConfig.sendMessage('user-registered', {
      ...userResult,
      role: newUser.role || 'user',
      createdAt: newUser.created_at
    });

    return userResult;
  }

  async login(
    data: LoginUserDTO,
    deviceMeta: { deviceId: string, ip: string, deviceName: string }
  ): Promise<AuthResponse> {
    const { email, password } = data;
    const { deviceId, ip, deviceName } = deviceMeta;

    const user = await this.userRepository.findByEmail(email);
    if (!user) throw new AppError('Invalid credentials', 400);

    const isMatch = await bcrypt.compare(password, user.password_hash);
    if (!isMatch) throw new AppError("Invalid credentials", 400);

    const accessToken = jwt.sign(
      { userId: user._id, deviceId: deviceId },
      env.JWT_ACCESS_SECRET as string,
      { expiresIn: '15m' }
    );

    const refreshToken = jwt.sign(
      { userId: user._id },
      env.JWT_REFRESH_SECRET as string,
      { expiresIn: '7d' }
    );

    const sessionObject = {
      refreshToken,
      user: {
        id: user._id.toString(),
        role: user.role || 'user',
        name: user.full_name,
        email: user.email,
      },
      meta: { ip, deviceName }
    };

    const sessionKey = `session:${user._id.toString()}:${deviceId}`;
    const ttl = 7 * 24 * 60 * 60; // ৭ দিন সেকেন্ডে

    await redisService.set(
      sessionKey,
      JSON.stringify(sessionObject),
      'EX',
      ttl
    );

    return {
      accessToken,
      refreshToken,
      user: {
        id: user._id.toString(),
        name: user.full_name,
        email: user.email
      }
    };
  }

  async verifySession(token: string): Promise<SessionVerification> {
    try {
      const decoded = jwt.verify(token, env.JWT_ACCESS_SECRET as string) as AccessTokenPayload;
      const { userId, deviceId } = decoded;

      const sessionData = await redisService.get(`session:${userId}:${deviceId}`);
      if (!sessionData) return { isValid: false };

      const session = JSON.parse(sessionData) as {
        user: { id: string; role: string; email: string };
        meta: { ip: string; deviceName: string };
      };

      return {
        isValid: true,
        userId: session.user.id,
        role: session.user.role,
        email: session.user.email,
        deviceId: deviceId,
        ip: session.meta.ip,
        deviceName: session.meta.deviceName
      };
    } catch {
      return { isValid: false };
    }
  }
}