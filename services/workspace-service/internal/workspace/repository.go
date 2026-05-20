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

func (r *sqlRepository) CreateInvite(ctx context.Context, invite *domain.WorkspaceInvitation) error {
	query := `
		INSERT INTO workspace_invitations (workspace_id, inviter_id, email, role, token, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`

	return r.db.QueryRowContext(
		ctx, query,
		invite.WorkspaceID, invite.InviterID, invite.Email,
		invite.Role, invite.Token, invite.Status, invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt)
}

func (r *sqlRepository) FindInviteByToken(ctx context.Context, token string) (*domain.WorkspaceInvitation, error) {
	var invite domain.WorkspaceInvitation
	query := `SELECT id, workspace_id, inviter_id, email, role, token, status, expires_at, created_at 
	          FROM workspace_invitations WHERE token = $1`

	err := r.db.GetContext(ctx, &invite, query, token)
	if err != nil {
		return nil, err
	}
	return &invite, nil
}
