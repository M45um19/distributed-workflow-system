import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import path from 'path';
import grpcServer from '../../config/grpc';
import { authService } from './auth.routes';

const PROTO_PATH = path.resolve(__dirname, '../../../../../shared-proto/auth/auth.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

const authProto = grpc.loadPackageDefinition(packageDefinition) as any;

export const verifySessionHandler = async (call: any, callback: any) => {
  const { token } = call.request; 
  const result = await authService.verifySession(token);
  callback(null, result);
};


export const registerAuthGrpcService = () => {
  grpcServer.addService(authProto.auth.AuthService.service, {
    VerifySession: verifySessionHandler, 
  });
  console.log("Auth gRPC service registered");
};