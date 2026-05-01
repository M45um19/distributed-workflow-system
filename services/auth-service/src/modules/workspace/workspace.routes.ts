// workspace.routes.ts
import { Router } from 'express';
import { WorkspaceRepository } from './workspace.repository';
import { WorkspaceService } from './workspace.service';
import { WorkspaceController } from './workspace.controller';
import { validateRequest } from '../../middleware/validation.middleware';
import { createWorkspaceSchema } from './workpace.validation';
import { protect } from '../../middleware/auth.middleware';

const router = Router();

const workspaceRepo = new WorkspaceRepository();
const workspaceService = new WorkspaceService(workspaceRepo);
const workspaceController = new WorkspaceController(workspaceService);

router.post('/create', validateRequest(createWorkspaceSchema), protect, workspaceController.createWorkspace);

export const WorkspaceRouter = router;