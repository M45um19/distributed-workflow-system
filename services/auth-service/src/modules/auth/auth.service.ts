import { IAuthService } from './auth.interface';
import { IUserRepository } from '../user/user.interface';
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { env } from '../../config/env';
import { AppError } from '../../utils/appError';

export class AuthService implements IAuthService {
  private userRepository: IUserRepository;

  constructor(userRepository: IUserRepository) {
    this.userRepository = userRepository;
  }

  async register(fullName: string, email: string, password: string): Promise<any> {
    const isUserExist = await this.userRepository.exists(email);
    if (isUserExist) throw new AppError('User already exists', 400);

    const salt = await bcrypt.genSalt(12);
    const hashedPassword = await bcrypt.hash(password, salt);

    const newUser = await this.userRepository.create({
      full_name: fullName,
      email,
      password_hash: hashedPassword
    });

    return { id: newUser._id, name: newUser.full_name, email: newUser.email };
  }

  async login(email: string, password: string): Promise<any> {
    const user = await this.userRepository.findByEmail(email);
    if (!user) throw new AppError('Invalid credentials', 400);

    const isMatch = await bcrypt.compare(password, user.password_hash);
    if (!isMatch) throw new AppError("Password mismatch", 400);
    const token = jwt.sign(
      { userId: user._id },
      env.JWT_SECRET as string,
      { expiresIn: '1d' }
    );

    return { token, user: { id: user._id, name: user.full_name, email: user.email } };
  }
}