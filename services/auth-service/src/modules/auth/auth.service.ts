import bcrypt from 'bcryptjs';
import jwt, { JwtPayload } from 'jsonwebtoken';

import { env } from '../../config/env.js';
import { kafkaConfig } from '../../config/kafka.js';
import { redisService } from '../../config/redis.js';
import { AppError } from '../../utils/appError.js';
import { IUserRepository } from '../user/user.interface.js';

import { AuthResponse, IAuthService, SessionVerification } from './auth.interface.js';
import { LoginUserDTO, RegisterUserDTO } from './auth.validation.js';

interface AccessTokenPayload extends JwtPayload {
  userId: string;
  deviceId: string;
}

export class AuthService implements IAuthService {
  constructor(private userRepository: IUserRepository) { }

  async register(data: RegisterUserDTO, deviceMeta: { deviceId: string, ip: string, deviceName: string }): Promise<AuthResponse> {
    const { full_name, email, password } = data;
    const { deviceId, ip, deviceName } = deviceMeta;

    const isUserExist = await this.userRepository.exists(email);
    if (isUserExist) throw new AppError('User already exists', 400);

    const salt = await bcrypt.genSalt(12);
    const hashedPassword = await bcrypt.hash(password, salt);

    const newUser = await this.userRepository.create({
      full_name,
      email,
      password_hash: hashedPassword,
      role: 'USER',
      avatar_url: '',
      is_active: true
    });

    const userResult: AuthResponse["user"] = {
      id: newUser._id.toString(),
      full_name: newUser.full_name,
      email: newUser.email
    };

    await kafkaConfig.sendMessage('user-registered', {
      ...userResult,
      role: newUser.role,
      createdAt: newUser.created_at
    });

    const accessToken = jwt.sign(
      { userId: newUser._id.toString(), deviceId: deviceId },
      env.JWT_ACCESS_SECRET as string,
      { expiresIn: '15m' }
    );

    const refreshToken = jwt.sign(
      { userId: newUser._id.toString(), deviceId: deviceId},
      env.JWT_REFRESH_SECRET as string,
      { expiresIn: '7d' }
    );

    const sessionObject = {
      refreshToken,
      user: {
        id: newUser._id.toString(),
        role: newUser.role || 'user',
        name: newUser.full_name,
        email: newUser.email,
      },
      meta: { ip, deviceName }
    };

    const sessionKey = `session:${newUser._id.toString()}:${deviceId}`;
    const ttl = 7 * 24 * 60 * 60;

    await redisService.set(
      sessionKey,
      JSON.stringify(sessionObject),
      'EX',
      ttl
    );

    return {accessToken, refreshToken, user: userResult};
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
      { userId: user._id.toString(), deviceId: deviceId },
      env.JWT_ACCESS_SECRET as string,
      { expiresIn: '15m' }
    );

    const refreshToken = jwt.sign(
      { userId: user._id.toString(), deviceId: deviceId },
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
    const ttl = 7 * 24 * 60 * 60;

    await redisService.set(
      sessionKey,
      JSON.stringify(sessionObject),
      'EX',
      ttl
    );

    const loginResult: AuthResponse = {
      accessToken,
      refreshToken,
      user: {
        id: user._id.toString(),
        full_name: user.full_name,
        email: user.email
      }
    };

    return loginResult;
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