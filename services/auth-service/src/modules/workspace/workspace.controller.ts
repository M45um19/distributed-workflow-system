// workspace.controller.ts
import { Response, NextFunction, Request } from 'express';
import { IWorkspaceService } from './workspace.interface';
import { sendResponse } from '../../utils/sendResponse';
import { CreateWorkspaceDTO } from './workpace.validation';

export class WorkspaceController {
  constructor(private workspaceService: IWorkspaceService) {}

  createWorkspace = async (req: any, res: Response, next: NextFunction): Promise<void> => {
    try {
      const body: CreateWorkspaceDTO = req.body;
      const ownerId = req.user.id;

      const result = await this.workspaceService.createWorkspace(body, ownerId);

      sendResponse(res, {
        statusCode: 201,
        success: true,
        message: 'Workspace created successfully!',
        data: result,
      });
    } catch (error) {
      next(error);
    }
  };
}