package workspace

import (
	"context"
	"time"
)

// Workspace Entity
type Workspace struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	OwnerID     string    `db:"owner_id" json:"owner_id"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type CreateWorkspaceInput struct {
	Name        string `json:"name" validate:"required,min=3,max=50"`
	Slug        string `json:"slug" validate:"required,min=3"`
	Description string `json:"description" validate:"max=255"`
}

type Repository interface {
	Create(ctx context.Context, ws *Workspace) error
	FindBySlug(ctx context.Context, slug string) (*Workspace, error)
}

type Service interface {
	CreateWorkspace(ctx context.Context, input CreateWorkspaceInput, ownerID string) (*Workspace, error)
}
