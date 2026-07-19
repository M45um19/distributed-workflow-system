package project

import (
	"context"
	"database/sql"
	"errors"
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

func (r *sqlRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var p domain.Project
	query := `SELECT id, workspace_id, name, description, status, created_by, created_at FROM projects WHERE id = $1`
	err := r.db.GetContext(ctx, &p, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *sqlRepository) GetByWorkspaceID(ctx context.Context, workspaceID string, limit int, cursor string) ([]domain.Project, error) {
	var projects []domain.Project
	var err error
	if cursor == "" {
		query := `SELECT id, workspace_id, name, description, status, created_by, created_at FROM projects WHERE workspace_id = $1 ORDER BY id DESC LIMIT $2`
		err = r.db.SelectContext(ctx, &projects, query, workspaceID, limit)
	} else {
		query := `SELECT id, workspace_id, name, description, status, created_by, created_at FROM projects WHERE workspace_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`
		err = r.db.SelectContext(ctx, &projects, query, workspaceID, cursor, limit)
	}
	if err != nil {
		return nil, err
	}
	if projects == nil {
		projects = []domain.Project{}
	}
	return projects, nil
}

