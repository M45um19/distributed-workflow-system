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

type AcceptInviteRequest struct {
	Token string `json:"token" binding:"required"`
}

type WorkspaceInviteResponse struct {
	InviteURL string `json:"invite_url"`
}

type WorkspaceMemberResponse struct {
	UserID   string    `db:"user_id" json:"user_id"`
	FullName string    `db:"full_name" json:"full_name"`
	Email    string    `db:"email" json:"email"`
	Role     string    `db:"role" json:"role"`
	JoinedAt time.Time `db:"joined_at" json:"joined_at"`
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
	Status      string    `db:"status" json:"status" binding:"required,oneof=PENDING ACCEPTED EXPIRED REVOKED"`
	ExpiresAt   time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type WorkspaceMember struct {
	ID          string    `db:"id" json:"id"`
	WorkspaceID string    `db:"workspace_id" json:"workspace_id"`
	UserID      string    `db:"user_id" json:"user_id"`
	Role        string    `db:"role" json:"role" binding:"required,oneof=ADMIN MEMBER VIEWER"`
	JoinedAt    time.Time `db:"joined_at" json:"joined_at"`
}

//model end

type WorkspaceRepository interface {
	Create(ctx context.Context, ws *Workspace) error
	FindBySlug(ctx context.Context, slug string) (*Workspace, error)
	GetByOwnerID(ctx context.Context, ownerId string, limit, offset int) ([]Workspace, error)
	GetByMemberID(ctx context.Context, userID string, limit, offset int) ([]Workspace, error)

	CreateInvite(ctx context.Context, invite *WorkspaceInvitation) error
	FindInviteByToken(ctx context.Context, token string) (*WorkspaceInvitation, error)

	UpdateInviteStatus(ctx context.Context, id string, status string) error

	AddMember(ctx context.Context, member *WorkspaceMember) error
	IsMember(ctx context.Context, workspaceID, userID string) (bool, error)
	FindByID(ctx context.Context, id string) (*Workspace, error)
	GetMembers(ctx context.Context, workspaceID string) ([]WorkspaceMemberResponse, error)
	GetMemberRole(ctx context.Context, workspaceID, userID string) (string, error)
}

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, input WorkspaceCreateInput, ownerID string) (*Workspace, error)
	GetWorkspacesByOwner(ctx context.Context, ownerId string, limit, page int) ([]Workspace, error)
	GetWorkspacesByMember(ctx context.Context, userID string, limit, page int) ([]Workspace, error)
	InviteUser(ctx context.Context, input WorkspaceInviteRequest) (*WorkspaceInviteResponse, error)
	AcceptInvitation(ctx context.Context, token string, loggedInUserID string) error
	GetWorkspaceMembers(ctx context.Context, workspaceID string, userID string) ([]WorkspaceMemberResponse, error)
}
