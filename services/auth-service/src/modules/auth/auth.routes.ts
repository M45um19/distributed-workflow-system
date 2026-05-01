import { Router } from 'express';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';
import { UserRepository } from '../user/user.repository';
import { validateRequest } from '../../middleware/validation.middleware';
import { loginSchema, registerSchema } from './auth.validation';

const router = Router();

const userRepository = new UserRepository();
export const authService = new AuthService(userRepository);
const authController = new AuthController(authService);

router.post(
  '/register',
  validateRequest(registerSchema), 
  authController.register
);

router.post(
  '/login',
  validateRequest(loginSchema), 
  authController.login
);

export const authRouter = router;