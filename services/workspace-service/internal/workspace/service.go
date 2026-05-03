package workspace

import (
	"context"
	"errors"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput, ownerID string) (*Workspace, error) {
	exists, _ := s.repo.FindBySlug(ctx, input.Slug)
	if exists != nil {
		return nil, errors.New("workspace with this slug already exists")
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
