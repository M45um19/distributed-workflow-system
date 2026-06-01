package user

import (
	"context"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
)

type service struct {
	userRepo domain.UserRepository
}

func NewService(userRepo domain.UserRepository) domain.UserService {
	return &service{
		userRepo: userRepo,
	}
}

func (s *service) SyncUserSnapshot(ctx context.Context, user *domain.UserSnapshot) error {
	return s.userRepo.UpsertUser(ctx, user)
}
