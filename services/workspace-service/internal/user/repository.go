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
        INSERT INTO users (id, full_name, email, avatar_url, role, created_at)
        VALUES (:id, :full_name, :email, :avatar_url, :role, :created_at)
        ON CONFLICT (id) DO UPDATE SET
            full_name = EXCLUDED.full_name,
            email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
            role = EXCLUDED.role`

	_, err := r.db.NamedExecContext(ctx, query, u)
	return err
}

func (r *repository) FindByID(ctx context.Context, id string) (*domain.UserSnapshot, error) {
	var u domain.UserSnapshot
	query := `SELECT id, full_name, email, avatar_url, role, created_at FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &u, query, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
