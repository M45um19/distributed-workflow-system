import { NextFunction, Request, Response } from 'express';

import { sendResponse } from '../../utils/sendResponse';

import { IAuthService } from './auth.interface';

export class AuthController {
  constructor(private authService: IAuthService) { }

  register = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {

      const xForwardedFor = req.headers['x-forwarded-for'];
      const ip = (Array.isArray(xForwardedFor) ? xForwardedFor[0] : xForwardedFor)
        || req.socket.remoteAddress
        || 'unknown';

      const userAgent = req.headers['user-agent'];
      const deviceName = (Array.isArray(userAgent) ? userAgent[0] : userAgent) || 'unknown_device';

      const deviceId = Buffer.from(`${deviceName}-${ip}`).toString('base64').substring(0, 16);

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

      const userAgent = req.headers['user-agent'];
      const deviceName = (Array.isArray(userAgent) ? userAgent[0] : userAgent) || 'unknown_device';

      const deviceId = Buffer.from(`${deviceName}-${ip}`).toString('base64').substring(0, 16);

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
}