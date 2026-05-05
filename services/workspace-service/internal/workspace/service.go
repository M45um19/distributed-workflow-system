package workspace

import (
	"context"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
)

type service struct {
	repo Repository
}

type Service interface {
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput, ownerID string) (*Workspace, error)
	GetUserWorkspaces(ctx context.Context, ownerId string) ([]Workspace, error)
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput, ownerID string) (*Workspace, error) {
	exists, _ := s.repo.FindBySlug(ctx, input.Slug)
	if exists != nil {
		return nil, apperror.BadRequest("workspace with this slug already exists")
	}

	ws := &Workspace{
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *service) GetUserWorkspaces(ctx context.Context, ownerId string) ([]Workspace, error) {
	return s.repo.GetByOwnerID(ctx, ownerId)
}
