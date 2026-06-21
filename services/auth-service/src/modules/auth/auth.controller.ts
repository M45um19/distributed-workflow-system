import { NextFunction, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';

import { AppError } from '../../utils/appError.js';
import { sendResponse } from '../../utils/sendResponse.js';

import { AuthRequest, IAuthService } from './auth.interface.js';


export class AuthController {
  constructor(private authService: IAuthService) { }

  register = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const xForwardedFor = req.headers['x-forwarded-for'];
      const ip = (Array.isArray(xForwardedFor) ? xForwardedFor[0] : xForwardedFor)
        || req.socket.remoteAddress
        || 'unknown';

      const userAgent = req.headers['user-agent'] || 'unknown_device';
      const deviceName = Array.isArray(userAgent) ? userAgent[0] : userAgent;

      const deviceId = uuidv4();

      const result = await this.authService.register(req.body, { deviceId, ip, deviceName });

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
      const xForwardedFor = req.headers['x-forwarded-for'];
      const ip = (Array.isArray(xForwardedFor) ? xForwardedFor[0] : xForwardedFor)
        || req.socket.remoteAddress
        || 'unknown';

      const userAgent = req.headers['user-agent'] || 'unknown_device';
      const deviceName = Array.isArray(userAgent) ? userAgent[0] : userAgent;

      let deviceId = (req.headers['x-device-id'] as string) || req.body.deviceId;

      if (!deviceId) {
        deviceId = uuidv4();
      }

      const result = await this.authService.login(req.body, { deviceId, ip, deviceName });

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

  logOut = async (req: AuthRequest, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = req.user?.userId;
      const deviceId = req.user?.deviceId;
      if (!userId || !deviceId) {
        throw new AppError("Invalid session information", 400);
      }
      await this.authService.logout({ userId, deviceId });
      sendResponse(res, {
        statusCode: 200,
        success: true,
        message: 'User logout successfully!',
        data: null,
      });
    } catch (error) {
      next(error)
    }
  }
}