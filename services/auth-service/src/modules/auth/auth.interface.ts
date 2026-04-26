export interface IAuthService {
  register(fullName: string, email: string, password: string): Promise<any>;
  login(email: string, password: string): Promise<any>;
}