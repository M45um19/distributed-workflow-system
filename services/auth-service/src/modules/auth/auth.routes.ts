import { Router } from "express";

import { AuthMiddleware } from "../../middleware/auth.middleware.js";
import { validateRequest } from "../../middleware/validation.middleware.js";

import { AuthController } from "./auth.controller.js";
import { loginSchema, refreshTokenSchema, registerSchema } from "./auth.validation.js";



export const setupAuthRoutes = (authCtrl: AuthController, authMiddleware: AuthMiddleware) => {
  const router = Router();
  router.post("/register", validateRequest(registerSchema), authCtrl.register);
  router.post("/login", validateRequest(loginSchema), authCtrl.login);
  router.post("/logout", authMiddleware.protect, authCtrl.logOut);
  router.post("/refresh-token", validateRequest(refreshTokenSchema), authCtrl.refreshToken)
  return router;
};