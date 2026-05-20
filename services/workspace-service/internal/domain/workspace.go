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
	WorkspaceID string `json:"workspace_id"`
	Email       string `json:"email" binding:"required,email"`
	Role        string `json:"role" binding:"required,oneof=ADMIN MEMBER VIEWER"`
	InviterID   string `json:"-"`
	Token       string `json:"token"`
}

// model start
type Workspace struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Slug        string    `db:"slug" json:"slug"`
	OwnerID     string    `db:"owner_id" json:"owner_id"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type WorkspaceInvitation struct {
	ID          string    `db:"id" json:"id"`
	WorkspaceID string    `db:"workspace_id" json:"workspace_id"`
	InviterID   string    `db:"inviter_id" json:"inviter_id"`
	Email       string    `db:"email" json:"email"`
	Role        string    `db:"role" json:"role"`
	Token       string    `db:"token" json:"token"`
	Status      string    `db:"status" json:"status"`
	ExpiresAt   time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

//model end

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *Workspace) error
	FindBySlug(ctx context.Context, slug string) (*Workspace, error)
	GetByOwnerID(ctx context.Context, ownerId string) ([]Workspace, error)

	CreateInvite(ctx context.Context, invite *WorkspaceInvitation) error
	FindInviteByToken(ctx context.Context, token string) (*WorkspaceInvitation, error)
}

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, input WorkspaceCreateInput, ownerID string) (*Workspace, error)
	GetUserWorkspaces(ctx context.Context, ownerId string) ([]Workspace, error)
	InviteUser(ctx context.Context, input WorkspaceInviteRequest) error
}
