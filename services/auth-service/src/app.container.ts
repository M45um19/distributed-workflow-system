import { AuthMiddleware } from "./middleware/auth.middleware.js";
import { AuthController } from "./modules/auth/auth.controller.js";
import { IAuthService } from "./modules/auth/auth.interface.js";
import { AuthService } from "./modules/auth/auth.service.js";
import { UserController } from "./modules/user/user.controller.js";
import { IUserService } from "./modules/user/user.interface.js";
import { UserRepository } from "./modules/user/user.repository.js";
import { UserService } from "./modules/user/user.service.js";

export class AppContainer {
  public authController: AuthController;
  public authService: IAuthService;
  public authMiddleware: AuthMiddleware;
  public userController: UserController;
  public userService: IUserService;

  constructor() {
    const userRepository = new UserRepository();
    this.authService = new AuthService(userRepository); 
    this.authController = new AuthController(this.authService);
    this.authMiddleware = new AuthMiddleware(this.authService);
    this.userService = new UserService(userRepository);
    this.userController = new UserController(this.userService);
  }
}