// workspace.repository.ts
import { Workspace } from './workspace.model';
import { IWorkspace } from './workpace.validation';
import { IWorkspaceRepository } from './workspace.interface';

export class WorkspaceRepository implements IWorkspaceRepository {
  async create(data: IWorkspace) {
    return await Workspace.create(data);
  }

  async findBySlug(slug: string) {
    return await Workspace.findOne({ slug });
  }
}