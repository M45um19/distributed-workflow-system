import path from 'path';
import { fileURLToPath } from 'url';

import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

import { grpcConfig } from '../../config/grpc.js';

import { IAuthService, SessionVerification } from './auth.interface.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PROTO_PATH = path.resolve(__dirname, '../../../../../shared-proto/auth/auth.proto');
const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

interface AuthPackage extends grpc.GrpcObject {
  auth: {
    AuthService: grpc.ServiceClientConstructor;
  };
}

const authProto = grpc.loadPackageDefinition(packageDefinition) as unknown as AuthPackage;

interface VerifySessionRequest {
  token: string;
}

interface VerifySessionResponse {
  isValid: boolean;
  userId: string;
  role: string;
  email: string;
  deviceId: string;
}

const verifySessionHandler = (authService: IAuthService): grpc.UntypedHandleCall =>
  async (
    call: grpc.ServerUnaryCall<VerifySessionRequest, VerifySessionResponse>,
    callback: grpc.sendUnaryData<VerifySessionResponse>
  ) => {
    try {
      const { token } = call.request;
      const result: SessionVerification = await authService.verifySession(token);

      callback(null, {
        isValid: result.isValid,
        userId: result.userId ?? '',
        role: result.role ?? '',
        email: result.email ?? '',
        deviceId: result.deviceId ?? '',
      });
    } catch (error: unknown) {
      const errorMessage = error instanceof Error ? error.message : "Internal gRPC Error";
      callback({
        code: grpc.status.INTERNAL,
        details: errorMessage,
      }, null);
    }
  };

export const registerAuthGrpcService = (authService: IAuthService): void => {
  const service = authProto.auth.AuthService.service;

  grpcConfig.getServer().addService(service, {
    VerifySession: verifySessionHandler(authService),
  });

  console.log("Auth gRPC service registered with injected AuthService");
};