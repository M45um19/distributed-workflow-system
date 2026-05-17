package domain

import (
	"context"
	"time"
)

type WorkspaceCreateInput struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Slug        string `json:"slug" binding:"required,min=3"`
	Description string `json:"description" binding:"max=255"`
}

type WorkspaceInviteRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
	Token string `json:"-"`
}

type Workspace struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	OwnerID     string    `db:"owner_id" json:"owner_id"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *Workspace) error
	FindBySlug(ctx context.Context, slug string) (*Workspace, error)
	GetByOwnerID(ctx context.Context, ownerId string) ([]Workspace, error)
}

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, input WorkspaceCreateInput, ownerID string) (*Workspace, error)
	GetUserWorkspaces(ctx context.Context, ownerId string) ([]Workspace, error)
	InviteUser(ctx context.Context, input WorkspaceInviteRequest) error
}
