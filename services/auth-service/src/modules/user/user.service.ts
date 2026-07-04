import { AppError } from '../../utils/appError.js';

import { IUserRepository, IUserService, UserProfileResponse } from './user.interface.js';
import { IUser } from './user.validation.js';

export class UserService implements IUserService {
  constructor(private userRepository: IUserRepository) {}

  async getUserProfile(userId: string): Promise<UserProfileResponse> {
    const user = await this.userRepository.findById(userId);
    if (!user) {
      throw new AppError('User not found', 404);
    }
    return {
      id: user._id.toString(),
      full_name: user.full_name,
      email: user.email,
      role: user.role,
      avatar_url: user.avatar_url,
      address: user.address,
      phone: user.phone,
      bio: user.bio,
      city: user.city,
      country: user.country,
      created_at: user.created_at,
    };
  }

  async updateUserProfile(userId: string, data: Partial<IUser>): Promise<UserProfileResponse> {
    const updatedUser = await this.userRepository.updateById(userId, data);
    if (!updatedUser) {
      throw new AppError('User not found', 404);
    }
    return {
      id: updatedUser._id.toString(),
      full_name: updatedUser.full_name,
      email: updatedUser.email,
      role: updatedUser.role,
      avatar_url: updatedUser.avatar_url,
      address: updatedUser.address,
      phone: updatedUser.phone,
      bio: updatedUser.bio,
      city: updatedUser.city,
      country: updatedUser.country,
      created_at: updatedUser.created_at,
    };
  }
}
