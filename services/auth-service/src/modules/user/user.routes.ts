import { Router } from "express";

import { AuthMiddleware } from "../../middleware/auth.middleware.js";
import { validateRequest } from "../../middleware/validation.middleware.js";

import { UserController } from "./user.controller.js";
import { updateProfileSchema } from "./user.validation.js";

export const setupUserRoutes = (userCtrl: UserController, authMiddleware: AuthMiddleware) => {
  const router = Router();
  
  router.get("/profile", authMiddleware.protect, userCtrl.getProfile);
  router.put("/profile", authMiddleware.protect, validateRequest(updateProfileSchema), userCtrl.updateProfile);
  router.get("/sessions", authMiddleware.protect, userCtrl.getSessions);
  
  return router;
};
