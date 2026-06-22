import * as grpc from '@grpc/grpc-js';

export interface IGrpcConfig {
  start(port: string | number): void;
  getServer(): grpc.Server;
  stop(): Promise<void>;
}

class GrpcConfig implements IGrpcConfig {
  private readonly server: grpc.Server;

  constructor() {
    this.server = new grpc.Server();
  }

  public getServer(): grpc.Server {
    return this.server;
  }

  public start(port: string | number = 50051): void {
    this.server.bindAsync(
      `0.0.0.0:${port}`,
      grpc.ServerCredentials.createInsecure(),
      (err, boundPort) => {
        if (err) {
          console.error(`gRPC Server failed to bind: ${err.message}`);
          return;
        }
        console.info(`gRPC Server running on port: ${boundPort}`);
      }
    );
  }

public async stop(): Promise<void> {
  return new Promise((resolve) => {
    let isResolved = false;

  
    const forceKillTimer = setTimeout(() => {
      if (!isResolved) {
        console.warn('gRPC shutdown timed out. Forcing stop...');
        this.server.forceShutdown();
        isResolved = true;
        resolve();
      }
    }, 2000);

    this.server.tryShutdown(() => {
      if (!isResolved) {
        clearTimeout(forceKillTimer);
        console.info('gRPC Server gracefully shut down');
        isResolved = true;
        resolve();
      }
    });
  });
}
}

export const grpcConfig = new GrpcConfig();