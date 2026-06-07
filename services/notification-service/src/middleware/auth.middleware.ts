import { Request, Response, NextFunction } from 'express';
import jwt from 'jsonwebtoken';

import { env } from '../config/env.js';
import { grpcConfig, IVerifySessionResponse } from '../config/grpc.js';
import { redisService } from '../config/redis.js';
import { AppError } from '../utils/appError.js';

interface IJwtPayload extends jwt.JwtPayload {
    userId: string;
    deviceId: string;
}

interface IAuthGrpcClient {
    VerifySession(
        argument: { token: string },
        callback: (err: Error | null, response: IVerifySessionResponse | null) => void
    ): void;
}

export interface IAuthenticatedRequest extends Request {
    user_id?: string;
}

export class AuthMiddleware {
    private readonly jwtSecret: string;

    constructor() {
        this.jwtSecret = env.JWT_SECRET || 'your_jwt_secret';
    }

    public protect() {
        return async (req: IAuthenticatedRequest, _res: Response, next: NextFunction): Promise<void> => {
            try {
                const authHeader = req.headers.authorization;
                if (!authHeader || !authHeader.startsWith('Bearer ')) {
                    return next(new AppError('Authorization header is required', 401));
                }

                const tokenString = authHeader.split(' ')[1];
                if (!tokenString) {
                    return next(new Error('Authentication error: Invalid Token format'));
                }
                let decoded: IJwtPayload;
                try {
                    const verified = jwt.verify(tokenString, this.jwtSecret);
                    decoded = verified as unknown as IJwtPayload;
                } catch {
                    return next(new AppError('Invalid or expired token', 401));
                }

                const { userId, deviceId } = decoded;
                const redisKey = `session:${userId}:${deviceId}`;

                const sessionData = await redisService.get(redisKey);
                if (sessionData === 'active') {
                    req.user_id = userId;
                    return next();
                }

                const client = grpcConfig.getAuthClient() as unknown as IAuthGrpcClient;
                client.VerifySession({ token: tokenString }, (err: Error | null, response: IVerifySessionResponse | null) => {
                    if (err || !response || !response.isValid) {
                        return next(new AppError('Session expired or invalid', 401));
                    }

                    void redisService.set(redisKey, 'active', 'EX', 15 * 60);

                    req.user_id = response.userId;
                    return next();
                });

            } catch (error) {
                return next(new AppError(`Auth Middleware Error: ${(error as Error).message}`, 401));
            }
        };
    }
}