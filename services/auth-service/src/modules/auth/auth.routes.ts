import { Router } from "express";
import { AuthController } from "./auth.controller";

export const setupAuthRoutes = (authCtrl: AuthController) => {
  const router = Router();
  router.post("/register", authCtrl.register);
  router.post("/login", authCtrl.login);
  return router;
};