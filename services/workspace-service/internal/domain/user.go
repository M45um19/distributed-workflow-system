package domain

import (
	"context"
	"time"
)

type UserSnapshot struct {
	ID        string    `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type UserRepository interface {
	UpsertUser(ctx context.Context, u *UserSnapshot) error
}

type UserService interface {
	SyncUserSnapshot(ctx context.Context, user *UserSnapshot) error
}
