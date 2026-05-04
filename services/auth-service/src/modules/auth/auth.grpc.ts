import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';
import grpcServer from '../../config/grpc';
import { AuthService } from './auth.service';
import { IAuthService } from './auth.interface';

const PROTO_PATH = path.resolve(__dirname, '../../../../../shared-proto/auth/auth.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const authProto = grpc.loadPackageDefinition(packageDefinition) as any;

const verifySessionHandler = (authService: IAuthService) => async (call: any, callback: any) => {
  try {
    const { token } = call.request; 
    const result = await authService.verifySession(token);
    callback(null, result);
  } catch (error: any) {
    callback({
      code: grpc.status.INTERNAL,
      message: error.message,
    });
  }
};

export const registerAuthGrpcService = (authService: IAuthService) => {
  grpcServer.addService(authProto.auth.AuthService.service, {
    VerifySession: verifySessionHandler(authService), 
  });
  console.log("Auth gRPC service registered with injected AuthService");
};