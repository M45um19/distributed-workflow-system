package project

import (
	"context"
	"database/sql"
	"errors"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
)

type service struct {
	projectRepo domain.ProjectRepository
	wsRepo      domain.WorkspaceRepository
}

func NewService(projectRepo domain.ProjectRepository, wsRepo domain.WorkspaceRepository) domain.ProjectService {
	return &service{
		projectRepo: projectRepo,
		wsRepo:      wsRepo,
	}
}

func (s *service) CreateProject(ctx context.Context, workspaceID string, input domain.ProjectCreateInput, userID string) (*domain.Project, error) {
	ws, err := s.wsRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	if ws.OwnerID != userID {
		role, err := s.wsRepo.GetMemberRole(ctx, workspaceID, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperror.Forbidden("You are not a member of this workspace")
			}
			return nil, err
		}
		if role != "ADMIN" {
			return nil, apperror.Forbidden("Only workspace owners and admins can create projects")
		}
	}

	p := &domain.Project{
		WorkspaceID: workspaceID,
		Name:        input.Name,
		Description: input.Description,
		Status:      "ACTIVE",
		CreatedBy:   userID,
	}

	if err := s.projectRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *service) GetProjectsByWorkspace(ctx context.Context, workspaceID string, userID string) ([]domain.Project, error) {
	ws, err := s.wsRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	if ws.OwnerID != userID {
		isMember, err := s.wsRepo.IsMember(ctx, workspaceID, userID)
		if err != nil {
			return nil, err
		}
		if !isMember {
			return nil, apperror.Forbidden("You do not have access to this workspace")
		}
	}

	return s.projectRepo.GetByWorkspaceID(ctx, workspaceID)
}
