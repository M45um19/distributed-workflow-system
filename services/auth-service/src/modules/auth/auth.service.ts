import { IAuthService } from './auth.interface';
import { IUserRepository } from '../user/user.interface';
import { RegisterUserDTO, LoginUserDTO } from './auth.validation'; 
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { env } from '../../config/env';
import { AppError } from '../../utils/appError';
import redisClient from '../../config/redis';

export class AuthService implements IAuthService {
  constructor(private userRepository: IUserRepository) {} 

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

async login(data: LoginUserDTO): Promise<any> {
    const { email, password } = data;

    const user = await this.userRepository.findByEmail(email); 
    if (!user) throw new AppError('Invalid credentials', 400);

    const isMatch = await bcrypt.compare(password, user.password_hash);
    if (!isMatch) throw new AppError("Invalid credentials", 400);

    const accessToken = jwt.sign(
      { userId: user._id },
      env.JWT_ACCESS_SECRET as string,
      { expiresIn: '15m' }
    );

    const refreshToken = jwt.sign(
      { userId: user._id },
      env.JWT_REFRESH_SECRET as string,
      { expiresIn: '7d' }
    );

    const REDIS_TTL = 7 * 24 * 60 * 60; 
    await redisClient.set(
      `refresh_token:${user._id}`,
      refreshToken,
      'EX',
      REDIS_TTL
    );

    return { 
      accessToken, 
      refreshToken, 
      user: { id: user._id, name: user.full_name, email: user.email } 
    };
  }
}