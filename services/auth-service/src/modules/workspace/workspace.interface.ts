// workspace.interface.ts
import { CreateWorkspaceDTO, IWorkspace } from './workpace.validation';

export interface IWorkspaceRepository {
  create(data: IWorkspace): Promise<any>;
  findBySlug(slug: string): Promise<any>;
}
export interface IWorkspaceService {
  createWorkspace(data: CreateWorkspaceDTO, ownerId: string): Promise<any>;
}