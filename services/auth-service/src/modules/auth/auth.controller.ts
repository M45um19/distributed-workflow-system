import { NextFunction, Request, Response } from 'express';
import { IAuthService } from './auth.interface';
import { sendResponse } from '../../utils/sendResponse';

export class AuthController {
  private authService: IAuthService;

  constructor(authService: IAuthService) {
    this.authService = authService;
  }

  register = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { full_name, email, password } = req.body;
      const data = await this.authService.register(full_name, email, password);
      sendResponse(res, {
        statusCode: 201,
        success: true,
        message: 'User registered successfully!',
        data,
      });
    } catch (error: any) {
      next(error)
    }
  };

  login = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { email, password } = req.body;
      const data = await this.authService.login(email, password);

      sendResponse(res, {
        statusCode: 200,
        success: true,
        message: 'User login successfully!',
        data,
      });
    } catch (error: any) {
      next(error)
    }
  };
}