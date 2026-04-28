import { LoginUserDTO, RegisterUserDTO } from "./auth.validation";

export interface IAuthService {
  register(data: RegisterUserDTO): Promise<any>;
  login(data: LoginUserDTO, deviceMeta: { deviceId: string; ip: any; deviceName: string }): Promise<any>;
}