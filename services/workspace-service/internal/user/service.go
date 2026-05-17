package user

import (
	"context"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
)

type service struct {
	repo domain.UserRepository
}

func NewService(repo domain.UserRepository) domain.UserService {
	return &service{
		repo: repo,
	}
}

func (s *service) SyncUserSnapshot(ctx context.Context, user *domain.UserSnapshot) error {
	return s.repo.UpsertUser(ctx, user)
}
