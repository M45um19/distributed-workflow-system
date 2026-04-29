import { IAuthService } from './auth.interface';
import { IUserRepository } from '../user/user.interface';
import { RegisterUserDTO, LoginUserDTO } from './auth.validation';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { env } from '../../config/env';
import { AppError } from '../../utils/appError';
import redisClient from '../../config/redis';

export class AuthService implements IAuthService {
  constructor(private userRepository: IUserRepository) { }

  async register(data: RegisterUserDTO): Promise<any> {
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

    return {
      id: newUser._id,
      name: newUser.full_name,
      email: newUser.email
    };
  }

  async login(data: LoginUserDTO, deviceMeta: { deviceId: string, ip: any, deviceName: string }): Promise<any> {
    const { email, password } = data;

    const { deviceId, ip, deviceName } = deviceMeta

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
      refreshToken: refreshToken,
      user: {
        id: user._id,
        role: (user as any).role || 'user',
        name: user.full_name,
        email: user.email,
      },
      meta: { ip, deviceName }
    };

    await redisClient.set(
      `session:${user._id.toString()}:${deviceId}`,
      JSON.stringify(sessionObject),
      'EX',
      7 * 24 * 60 * 60
    );

    return {
      accessToken,
      refreshToken,
      user: { id: user._id, name: user.full_name, email: user.email }
    };
  }

  async verifySession(token: string): Promise<any> {
    try {
      const decoded = jwt.verify(token, env.JWT_ACCESS_SECRET as string) as any;
      const { userId, deviceId } = decoded;

      const sessionData = await redisClient.get(`session:${userId}:${deviceId}`);
      if (!sessionData) return { isValid: false };

      const session = JSON.parse(sessionData);

      return {
        isValid: true,
        userId: session.user.id,
        role: session.user.role,
        email: session.user.email,
        deviceId: deviceId,
        ip: session.meta.ip,
        deviceName: session.meta.deviceName
      };
    } catch (error) {
      return { isValid: false };
    }
  }
}