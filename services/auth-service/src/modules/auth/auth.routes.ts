import { Router } from "express";

import { AuthMiddleware } from "../../middleware/auth.middleware.js";

import { AuthController } from "./auth.controller.js";



export const setupAuthRoutes = (authCtrl: AuthController, authMiddleware: AuthMiddleware) => {
  const router = Router();
  router.post("/register", authCtrl.register);
  router.post("/login", authCtrl.login);
  router.post("/logout", authMiddleware.protect, authCtrl.logOut);
  return router;
};