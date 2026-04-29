import { Request, Response, NextFunction } from 'express';
import { authService } from '../modules/auth/auth.routes';
import { AppError } from '../utils/appError';

interface AuthRequest extends Request {
  user?: {
    id: string;
    role: string;
    email: string;
    deviceId: string;
  };
}

export const protect = async (req: AuthRequest, res: Response, next: NextFunction) => {
  try {
    let token;
    if (req.headers.authorization && req.headers.authorization.startsWith('Bearer')) {
      token = req.headers.authorization.split(' ')[1];
    }

    if (!token) {
      throw new AppError('Please log in', 401);
    }

    const session = await authService.verifySession(token);

    if (!session.isValid) {
      throw new AppError('Session invalid or expire', 401);
    }
    req.user = {
      id: session.userId,
      role: session.role,
      email: session.email,
      deviceId: session.deviceId
    };

    next();
  } catch (error) {
    next(error);
  }
};