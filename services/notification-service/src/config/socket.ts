import { Server as HttpServer } from 'http';

import { createAdapter } from '@socket.io/redis-adapter';
import { Server } from 'socket.io';

import { redisService } from './redis.js';

export interface ISocketConfig {
    init(httpServer: HttpServer): Server;
    initWorker(): void;
    getIO(): Server;
}

class SocketConfig implements ISocketConfig {
    private io: Server | null = null;

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

    private setupListeners(): void {
        if (!this.io) return;

        this.io.on('connection', (socket) => {
            const userId = socket.handshake.query.userId as string;

            if (userId) {
                void socket.join(userId);
            }

            socket.on('join-room', (customUserId: string) => {
                void socket.join(customUserId);
            });

            socket.on('disconnect', () => {
                console.info(`User Disconnected: ${socket.id}`);
            });
        });
    }
}

export const socketConfig = new SocketConfig();