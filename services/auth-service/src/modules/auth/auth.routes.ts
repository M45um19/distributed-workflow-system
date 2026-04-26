import { Router } from 'express';
import { AuthController } from './auth.controller';
import { AuthService } from './auth.service';
import { UserRepository } from '../user/user.repository';

const router = Router();

const userRepository = new UserRepository(); // Concrete Implementation
const authService = new AuthService(userRepository); // Injecting Repo into Service
const authController = new AuthController(authService); // Injecting Service into Controller

// API Endpoints
router.post('/register', authController.register);
router.post('/login', authController.login);

export const authRouter = router;