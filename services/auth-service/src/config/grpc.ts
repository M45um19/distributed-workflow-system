import * as grpc from '@grpc/grpc-js';

const grpcServer = new grpc.Server();

export const startGrpcServer = (port: string | number = 50051) => {
  grpcServer.bindAsync(
    `0.0.0.0:${port}`,
    grpc.ServerCredentials.createInsecure(),
    (err, boundPort) => {
      if (err) {
        console.error(`gRPC Server failed to bind: ${err.message}`);
        return;
      }
      console.log(`gRPC Server running on port: ${boundPort}`);
    }
  );
};

export default grpcServer;