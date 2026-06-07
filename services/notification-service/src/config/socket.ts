import { Server as HttpServer } from 'http';

import { createAdapter } from '@socket.io/redis-adapter';
import jwt from 'jsonwebtoken';
import { Server, Socket } from 'socket.io';

import { env } from './env.js';
import { grpcConfig, IVerifySessionResponse } from './grpc.js';
import { redisService } from './redis.js';

export interface ISocketConfig {
    init(httpServer: HttpServer): Server;
    initWorker(): void;
    getIO(): Server;
}

interface IAuthGrpcClient {
    VerifySession(
        argument: { token: string },
        callback: (err: Error | null, response: IVerifySessionResponse | null) => void
    ): void;
}

interface IJwtPayload extends jwt.JwtPayload {
    userId: string;
    deviceId: string;
}

interface AuthenticatedSocket extends Socket {
    userId?: string;
}

class SocketConfig implements ISocketConfig {
    private io: Server | null = null;
    private readonly jwtSecret: string;

    constructor() {
        this.jwtSecret = env.JWT_SECRET || 'your_jwt_secret';
    }

    public init(httpServer: HttpServer): Server {
        this.io = new Server(httpServer, {
            cors: {
                origin: '*',
                methods: ['GET', 'POST'],
            },
        });

        const pubClient = redisService.getClient();
        const subClient = pubClient.duplicate();
        this.io.adapter(createAdapter(pubClient, subClient));

        this.io.use((socket: AuthenticatedSocket, next) => {
            this.validateSocketToken(socket, next);
        });

        this.setupListeners();
        return this.io;
    }

    public initWorker(): void {
        const pubClient = redisService.getClient();
        const subClient = pubClient.duplicate();

        this.io = new Server();
        this.io.adapter(createAdapter(pubClient, subClient));
    }

    public getIO(): Server {
        if (!this.io) {
            this.initWorker();
        }

        if (!this.io) {
            throw new Error('[SocketConfig] Failed to initialize Socket.io Server instance.');
        }

        return this.io;
    }


    private async validateSocketToken(socket: AuthenticatedSocket, next: (err?: Error) => void): Promise<void> {
        try {
            const authHeader = socket.handshake.auth.token as string || socket.handshake.query.token as string;

            if (!authHeader || !authHeader.startsWith('Bearer ')) {
                return next(new Error('Authentication error: Token is required'));
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
                return next(new Error('Authentication error: Invalid or expired token'));
            }

            const { userId, deviceId } = decoded;
            const redisKey = `session:${userId}:${deviceId}`;

            const sessionData = await redisService.get(redisKey);
            if (sessionData === 'active') {
                socket.userId = userId;
                return next();
            }

            const client = grpcConfig.getAuthClient() as unknown as IAuthGrpcClient;

            const grpcResponse = await new Promise<IVerifySessionResponse | null>((resolve, reject) => {
                client.VerifySession({ token: tokenString }, (err: Error | null, response: IVerifySessionResponse | null) => {
                    if (err) return reject(err);
                    resolve(response);
                });
            });

            if (!grpcResponse || !grpcResponse.isValid) {
                return next(new Error('Authentication error: Session expired or invalid'));
            }

            await redisService.set(redisKey, 'active', 'EX', 15 * 60);

            socket.userId = grpcResponse.userId;
            return next();

        } catch (error) {
            return next(new Error(`Authentication error: ${(error as Error).message}`));
        }
    }
    private setupListeners(): void {
        if (!this.io) return;

        this.io.on('connection', (rawSocket) => {
            const socket = rawSocket as AuthenticatedSocket;

            const userId = socket.userId;

            if (userId) {
                void socket.join(`user_${userId}`);
                console.info(`[Socket Connected] User ${userId} successfully joined their personal room.`);
            }

            socket.on('disconnect', () => {
                console.info(`User Disconnected: ${socket.id} (UID: ${userId || 'UNKNOWN'})`);
            });
        });
    }
}

export const socketConfig = new SocketConfig();