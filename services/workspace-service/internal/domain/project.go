package domain

import (
	"context"
	"time"
)

type ProjectCreateInput struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	Description string `json:"description" binding:"max=255"`
}

type Project struct {
	ID          string    `db:"id" json:"id"`
	WorkspaceID string    `db:"workspace_id" json:"workspace_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	Status      string    `db:"status" json:"status"`
	CreatedBy   string    `db:"created_by" json:"created_by"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByWorkspaceID(ctx context.Context, workspaceID string, limit, offset int) ([]Project, error)
}

type ProjectService interface {
	CreateProject(ctx context.Context, workspaceID string, input ProjectCreateInput, userID string) (*Project, error)
	GetProjectsByWorkspace(ctx context.Context, workspaceID string, userID string, limit, page int) ([]Project, error)
}
