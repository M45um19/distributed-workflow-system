import path from 'path';
import { fileURLToPath } from 'url';

import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

import { env } from './env.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PROTO_PATH = env.AUTH_PROTO_PATH && env.AUTH_PROTO_PATH.startsWith('/') 
  ? env.AUTH_PROTO_PATH 
  : path.resolve(__dirname, env.AUTH_PROTO_PATH || '../../../../shared-proto/auth/auth.proto');
  const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

interface IGrpcAuthPackage extends grpc.GrpcObject {
  auth: {
    AuthService: grpc.ServiceClientConstructor;
  };
}

const grpcObject = grpc.loadPackageDefinition(packageDefinition) as unknown as IGrpcAuthPackage;
const AuthServiceConstructor = grpcObject.auth.AuthService;

export interface IVerifySessionResponse {
  isValid: boolean;
  userId: string;
  role: string;
  email: string;
  deviceId: string;
}

export interface IGrpcConfig {
  start(port: string | number): void;
  getServer(): grpc.Server;
  getAuthClient(): grpc.Client;
  stop(): Promise<void>;
}

class GrpcConfig implements IGrpcConfig {
  private readonly server: grpc.Server;
  private authClient: grpc.Client | null = null;

  constructor() {
    this.server = new grpc.Server();
  }

  public getServer(): grpc.Server {
    return this.server;
  }


  public getAuthClient(): grpc.Client {
    if (!this.authClient) {
      const address = env.AUTH_SERVICE_GRPC_ADDRESS || 'localhost:50051';
      this.authClient = new AuthServiceConstructor(
        address,
        grpc.credentials.createInsecure()
      );
      console.info(`[gRPC Config] Auth gRPC Client connected to: ${address}`);
    }
    return this.authClient;
  }

  public start(port: string | number = 50052): void {
    this.server.bindAsync(
      `0.0.0.0:${port}`,
      grpc.ServerCredentials.createInsecure(),
      (err, boundPort) => {
        if (err) {
          console.warn(`[gRPC Server] Failed to bind: ${err.message}`);
          return;
        }
        console.info(`[gRPC Server] Running on port: ${boundPort}`);
      }
    );
  }

  public async stop(): Promise<void> {
    return new Promise((resolve) => {
      this.server.tryShutdown(() => {
        console.info('[gRPC Server] Gracefully shut down');
        resolve();
      });
    });
  }
}

export const grpcConfig = new GrpcConfig();