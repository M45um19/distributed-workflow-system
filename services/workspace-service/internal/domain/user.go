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
	AvatarURL string    `db:"avatar_url" json:"avatar_url"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type UserLogoutPayload struct {
	UserID   string `db:"userId" json:"userId"`
	DeviceID string `db:"deviceId" json:"deviceId"`
}

type UserRepository interface {
	UpsertUser(ctx context.Context, u *UserSnapshot) error
	FindByID(ctx context.Context, id string) (*UserSnapshot, error)
}

type UserService interface {
	SyncUserSnapshot(ctx context.Context, user *UserSnapshot) error
}
