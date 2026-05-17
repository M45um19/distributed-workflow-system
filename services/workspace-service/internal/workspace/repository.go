package workspace

import (
	"context"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type sqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) domain.WorkspaceRepository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, ws *domain.Workspace) error {
	query := `INSERT INTO workspaces (name, slug, owner_id, description) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query, ws.Name, ws.Slug, ws.OwnerID, ws.Description).Scan(&ws.ID)
}

func (r *sqlRepository) FindBySlug(ctx context.Context, slug string) (*domain.Workspace, error) {
	var ws domain.Workspace
	err := r.db.GetContext(ctx, &ws, "SELECT * FROM workspaces WHERE slug=$1", slug)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *sqlRepository) GetByOwnerID(ctx context.Context, ownerId string) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace

	query := `SELECT id, name, slug, owner_id, description, created_at from workspaces WHERE owner_id=$1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &workspaces, query, ownerId)
	if err != nil {
		return nil, err
	}

	if workspaces == nil {
		workspaces = []domain.Workspace{}
	}

	return workspaces, nil
}
