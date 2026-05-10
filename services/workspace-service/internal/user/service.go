package user

import (
	"context"
)

type Service interface {
	SyncUserSnapshot(ctx context.Context, user *UserSnapshot) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) SyncUserSnapshot(ctx context.Context, user *UserSnapshot) error {
	return s.repo.UpsertUser(ctx, user)
}
