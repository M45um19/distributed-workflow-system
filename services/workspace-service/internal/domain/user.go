package domain

import (
	"context"
	"time"
)

type UserSnapshot struct {
	ID        string    `db:"id" json:"id"`
	FullName  string    `db:"full_name" json:"full_name"`
	Email     string    `db:"email" json:"email"`
	Role      string    `db:"role" json:"role" binding:"required,oneof=ADMIN USER"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type UserRepository interface {
	UpsertUser(ctx context.Context, u *UserSnapshot) error
	FindByID(ctx context.Context, id string) (*UserSnapshot, error)
}

type UserService interface {
	SyncUserSnapshot(ctx context.Context, user *UserSnapshot) error
}
