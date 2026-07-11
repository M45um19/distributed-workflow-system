package project

import (
	"context"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type sqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) domain.ProjectRepository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, workspaceID string, p *domain.Project) error {
	log.Println(p)
	query := `
		INSERT INTO projects (id, workspace_id, name, description, status, created_by) 
		VALUES ($1, $2, $3, $4, $5, $6) 
		RETURNING id, status, created_at`
	return r.db.QueryRowContext(ctx, query, p.ID, p.WorkspaceID, p.Name, p.Description, p.Status, p.CreatedBy).Scan(&p.ID, &p.Status, &p.CreatedAt)
}

func (r *sqlRepository) GetByWorkspaceID(ctx context.Context, workspaceID string, limit, offset int) ([]domain.Project, error) {
	var projects []domain.Project
	query := `SELECT id, workspace_id, name, description, status, created_by, created_at FROM projects WHERE workspace_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &projects, query, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	if projects == nil {
		projects = []domain.Project{}
	}
	return projects, nil
}
