import { NextFunction, Response } from 'express';

import { AppError } from '../../utils/appError.js';
import { sendResponse } from '../../utils/sendResponse.js';
import { AuthRequest } from '../auth/auth.interface.js';

import { IUserService } from './user.interface.js';

export class UserController {
  constructor(private userService: IUserService) {}

  getProfile = async (req: AuthRequest, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = req.user?.userId;
      if (!userId) {
        throw new AppError('Unauthorized', 401);
      }
      const profile = await this.userService.getUserProfile(userId);
      sendResponse(res, {
        statusCode: 200,
        success: true,
        message: 'Profile fetched successfully',
        data: profile,
      });
    } catch (error) {
      next(error);
    }
  };

  updateProfile = async (req: AuthRequest, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = req.user?.userId;
      if (!userId) {
        throw new AppError('Unauthorized', 401);
      }
      const profile = await this.userService.updateUserProfile(userId, req.body);
      sendResponse(res, {
        statusCode: 200,
        success: true,
        message: 'Profile updated successfully',
        data: profile,
      });
    } catch (error) {
      next(error);
    }
  };
}
