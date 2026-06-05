import { AuthMiddleware } from "./middleware/auth.middleware.js";
import { AuthController } from "./modules/auth/auth.controller.js";
import { IAuthService } from "./modules/auth/auth.interface.js";
import { AuthService } from "./modules/auth/auth.service.js";
import { UserRepository } from "./modules/user/user.repository.js";

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