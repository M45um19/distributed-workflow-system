package user

import (
	"context"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) domain.UserRepository {
	return &repository{db: db}
}

func (r *repository) UpsertUser(ctx context.Context, u *domain.UserSnapshot) error {
	query := `
        INSERT INTO users (id, name, email, role, created_at)
        VALUES (:id, :name, :email, :role, :created_at)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            role = EXCLUDED.role`

	_, err := r.db.NamedExecContext(ctx, query, u)
	return err
}
