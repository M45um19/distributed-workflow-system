// workspace.service.ts
import { IWorkspaceRepository, IWorkspaceService } from './workspace.interface';
import { CreateWorkspaceDTO } from './workpace.validation';
import { AppError } from '../../utils/appError';

export class WorkspaceService implements IWorkspaceService {
  constructor(private workspaceRepository: IWorkspaceRepository) {}

  async createWorkspace(data: CreateWorkspaceDTO, ownerId: string): Promise<any> {
    const { name, slug, description } = data;

    const isExist = await this.workspaceRepository.findBySlug(slug);
    if (isExist) {
      throw new AppError('Workspace with this slug already exists!', 400);
    }

    const newWorkspace = await this.workspaceRepository.create({
      name,
      slug,
      description,
      owner_id: ownerId as any,
    });

    return {
      id: newWorkspace._id,
      name: newWorkspace.name,
      slug: newWorkspace.slug,
      owner_id: newWorkspace.owner_id
    };
  }
}