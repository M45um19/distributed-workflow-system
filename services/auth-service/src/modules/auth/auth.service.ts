import { IAuthService } from './auth.interface';
import { IUserRepository } from '../user/user.interface';
import { RegisterUserDTO, LoginUserDTO } from './auth.validation'; // Zod থেকে ইনফার করা টাইপ
import bcrypt from 'bcryptjs';
import jwt from 'jsonwebtoken';
import { env } from '../../config/env';
import { AppError } from '../../utils/appError';

export class AuthService implements IAuthService {
  constructor(private userRepository: IUserRepository) {} // শর্টহ্যান্ড কনস্ট্রাক্টর

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

    // পাসওয়ার্ড চেক করার জন্য 'select: false' করা থাকলে তা ম্যানুয়ালি নিয়ে আসতে হতে পারে
    const user = await this.userRepository.findByEmail(email); 
    if (!user) throw new AppError('Invalid credentials', 400);

    const isMatch = await bcrypt.compare(password, user.password_hash);
    if (!isMatch) throw new AppError("Invalid credentials", 400); // সিকিউরিটির জন্য একই এরর মেসেজ রাখা ভালো

    const token = jwt.sign(
      { userId: user._id },
      env.JWT_SECRET as string,
      { expiresIn: '1d' }
    );

    return { 
      token, 
      user: { id: user._id, name: user.full_name, email: user.email } 
    };
  }
}