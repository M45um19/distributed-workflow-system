import { Response, NextFunction, Request } from 'express';

import { IAuthService, AuthRequest, AuthUser } from '../modules/auth/auth.interface';
import { AppError } from '../utils/appError';

export class AuthMiddleware {
  constructor(private authService: IAuthService) { }

  public protect = async (req: Request, _res: Response, next: NextFunction) => {
    try {
      let token: string | undefined;

      if (req.headers.authorization?.startsWith('Bearer')) {
        token = req.headers.authorization.split(' ')[1];
      }

      if (!token) throw new AppError('Please log in', 401);

      const session = await this.authService.verifySession(token);

      if (!session.isValid || !session.userId || !session.role || !session.email || !session.deviceId) {
        throw new AppError('Incomplete session data', 401);
      }

      const userData = {
        userId: session.userId,
        role: session.role,
        email: session.email,
        deviceId: session.deviceId,
        ip: session.ip ?? undefined,
        deviceName: session.deviceName ?? undefined
      } as AuthUser;

      (req as AuthRequest).user = userData;

      next();
    } catch (error) {
      next(error);
    }
  };
}