package user

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type Repository interface {
	UpsertUser(ctx context.Context, u *UserSnapshot) error
}

type repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

func (r *repository) UpsertUser(ctx context.Context, u *UserSnapshot) error {
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
