package workspace

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, ws *Workspace) error {
	query := `INSERT INTO workspaces (name, slug, owner_id, description) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.db.QueryRowContext(ctx, query, ws.Name, ws.Slug, ws.OwnerID, ws.Description).Scan(&ws.ID)
}

func (r *repository) FindBySlug(ctx context.Context, slug string) (*Workspace, error) {
	var ws Workspace
	err := r.db.GetContext(ctx, &ws, "SELECT * FROM workspaces WHERE slug=$1", slug)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}
