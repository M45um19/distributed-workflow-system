import { AuthMiddleware } from "./middleware/auth.middleware";
import { AuthController } from "./modules/auth/auth.controller";
import { IAuthService } from "./modules/auth/auth.interface";
import { AuthService } from "./modules/auth/auth.service";
import { UserRepository } from "./modules/user/user.repository";

export class AppContainer {
  public authController: AuthController;
  public authService: IAuthService;
  public authMiddleware: AuthMiddleware;

  constructor() {
    const userRepository = new UserRepository();
    this.authService = new AuthService(userRepository); 
    this.authController = new AuthController(this.authService);
    this.authMiddleware = new AuthMiddleware(this.authService);
  }
}