package user

import (
	"context"
	"fmt"
	"log"
	"strings"

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

func (r *repository) BulkUpsertUsers(ctx context.Context, users []domain.UserSnapshot) ([]domain.UserSnapshot, error) {
	if len(users) == 0 {
		return nil, nil
	}

	err := r.bulkUpsertUsersBatch(ctx, users)
	if err == nil {
		return nil, nil
	}

	if isConnectionError(err) {
		return nil, err
	}

	log.Printf("BulkUpsertUsers of full batch failed (%v), starting recursive divide-and-conquer split", err)

	failedUsers, err := r.recursiveUpsert(ctx, users)
	if err != nil {
		return nil, err
	}

	return failedUsers, nil
}

func (r *repository) recursiveUpsert(ctx context.Context, users []domain.UserSnapshot) ([]domain.UserSnapshot, error) {
	if len(users) == 0 {
		return nil, nil
	}

	err := r.bulkUpsertUsersBatch(ctx, users)
	if err == nil {
		return nil, nil
	}

	if isConnectionError(err) {
		return nil, err
	}

	if len(users) == 1 {
		log.Printf("Poisonous user snapshot detected: %s (Error: %v)", users[0].ID, err)
		return users, nil
	}

	mid := len(users) / 2
	left := users[:mid]
	right := users[mid:]

	var failedLeft, failedRight []domain.UserSnapshot
	var leftErr, rightErr error

	failedLeft, leftErr = r.recursiveUpsert(ctx, left)
	if leftErr != nil {
		return nil, leftErr
	}

	failedRight, rightErr = r.recursiveUpsert(ctx, right)
	if rightErr != nil {
		return nil, rightErr
	}

	return append(failedLeft, failedRight...), nil
}

func (r *repository) bulkUpsertUsersBatch(ctx context.Context, users []domain.UserSnapshot) error {
	placeholders := make([]string, len(users))
	flatArgs := make([]interface{}, 0, len(users)*6)
	for i, u := range users {
		placeholders[i] = "(?, ?, ?, ?, ?, ?)"
		flatArgs = append(flatArgs, u.ID, u.FullName, u.Email, u.AvatarURL, u.Role, u.CreatedAt)
	}

	query := fmt.Sprintf(`
		INSERT INTO users (id, full_name, email, avatar_url, role, created_at)
		VALUES %s
		ON CONFLICT (id) DO UPDATE SET
			full_name = EXCLUDED.full_name,
			email = EXCLUDED.email,
			avatar_url = EXCLUDED.avatar_url,
			role = EXCLUDED.role`, strings.Join(placeholders, ", "))

	query, args, err := sqlx.In(query, flatArgs...)
	if err != nil {
		return err
	}

	query = r.db.Rebind(query)

	_, err = r.db.ExecContext(ctx, query, args...)
	return err
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "driver: bad connection") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "EOF")
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

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.UserSnapshot, error) {
	var u domain.UserSnapshot
	query := `SELECT id, full_name, email, avatar_url, role, created_at FROM users WHERE email = $1`
	err := r.db.GetContext(ctx, &u, query, email)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
