import { NextFunction, Request, Response } from 'express';
import { IAuthService } from './auth.interface';
import { sendResponse } from '../../utils/sendResponse';

export class AuthController {
  constructor(private authService: IAuthService) {}

  register = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const result = await this.authService.register(req.body);

      sendResponse(res, {
        statusCode: 201,
        success: true,
        message: 'User registered successfully!',
        data: result,
      });
    } catch (error) {
      next(error);
    }
  };

  login = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const result = await this.authService.login(req.body);

      sendResponse(res, {
        statusCode: 200,
        success: true,
        message: 'User login successfully!',
        data: result,
      });
    } catch (error) {
      next(error);
    }
  };
}