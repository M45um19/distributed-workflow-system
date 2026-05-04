import { Response, NextFunction } from 'express';
import { IAuthService } from '../modules/auth/auth.interface';
import { AppError } from '../utils/appError';
import { AuthRequest } from '../modules/auth/auth.interface';

export class AuthMiddleware {
  constructor(private authService: IAuthService) {}

  public protect = async (req: AuthRequest, res: Response, next: NextFunction) => {
    try {
      let token;
      if (req.headers.authorization?.startsWith('Bearer')) {
        token = req.headers.authorization.split(' ')[1];
      }

      if (!token) throw new AppError('Please log in', 401);

      const session = await this.authService.verifySession(token);

      if (!session.isValid) throw new AppError('Session invalid', 401);

      req.user = {
        userId: session.userId,
        role: session.role,
        email: session.email,
        deviceId: session.deviceId,
        ip: session.ip,
        deviceName: session.deviceName
      };

      next();
    } catch (error) {
      next(error);
    }
  };
}