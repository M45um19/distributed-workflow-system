package workspace

import (
	"context"
	"database/sql"
	"errors"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type sqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) domain.WorkspaceRepository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, workspaceID string, ws *domain.Workspace) error {
	query := `INSERT INTO workspaces (id, name, slug, owner_id, description) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.db.QueryRowContext(ctx, query, ws.ID, ws.Name, ws.Slug, ws.OwnerID, ws.Description).Scan(&ws.ID)
}

func (r *sqlRepository) FindBySlug(ctx context.Context, workspaceID string, slug string) (*domain.Workspace, error) {
	var ws domain.Workspace
	var err error
	if workspaceID != "" {
		err = r.db.GetContext(ctx, &ws, "SELECT id, name, slug, owner_id, description, created_at FROM workspaces WHERE slug=$1 AND id=$2", slug, workspaceID)
	} else {
		err = r.db.GetContext(ctx, &ws, "SELECT id, name, slug, owner_id, description, created_at FROM workspaces WHERE slug=$1", slug)
	}
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *sqlRepository) GetByOwnerID(ctx context.Context, workspaceID string, ownerId string, limit, offset int) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace

	query := `SELECT id, name, slug, owner_id, description, created_at from workspaces WHERE owner_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &workspaces, query, ownerId, limit, offset)
	if err != nil {
		return nil, err
	}

	if workspaces == nil {
		workspaces = []domain.Workspace{}
	}

	return workspaces, nil
}

func (r *sqlRepository) CreateInvite(ctx context.Context, workspaceID string, invite *domain.WorkspaceInvitation) error {
	query := `
		INSERT INTO workspace_invitations (id, workspace_id, inviter_id, email, role, token, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	return r.db.QueryRowContext(
		ctx, query,
		invite.ID, invite.WorkspaceID, invite.InviterID, invite.Email,
		invite.Role, invite.Token, invite.Status, invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt)
}

func (r *sqlRepository) FindInviteByToken(ctx context.Context, workspaceID string, token string) (*domain.WorkspaceInvitation, error) {
	var invite domain.WorkspaceInvitation
	var err error
	if workspaceID != "" {
		query := `SELECT id, workspace_id, inviter_id, email, role, token, status, expires_at, created_at 
		          FROM workspace_invitations WHERE token = $1 AND workspace_id = $2`
		err = r.db.GetContext(ctx, &invite, query, token, workspaceID)
	} else {
		query := `SELECT id, workspace_id, inviter_id, email, role, token, status, expires_at, created_at 
		          FROM workspace_invitations WHERE token = $1`
		err = r.db.GetContext(ctx, &invite, query, token)
	}
	if err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *sqlRepository) UpdateInviteStatus(ctx context.Context, workspaceID string, id string, status string) error {
	query := `UPDATE workspace_invitations SET status = $1 WHERE id = $2 AND workspace_id = $3`
	_, err := r.db.ExecContext(ctx, query, status, id, workspaceID)
	return err
}

func (r *sqlRepository) IsMember(ctx context.Context, workspaceID string, userID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id = $1 AND user_id = $2)`
	err := r.db.GetContext(ctx, &exists, query, workspaceID, userID)
	return exists, err
}

func (r *sqlRepository) AddMember(ctx context.Context, workspaceID string, member *domain.WorkspaceMember) error {
	query := `
		INSERT INTO workspace_members (id, workspace_id, user_id, role) 
		VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, member.ID, member.WorkspaceID, member.UserID, member.Role)
	return err
}

func (r *sqlRepository) GetByMemberID(ctx context.Context, workspaceID string, userID string, limit, offset int) ([]domain.Workspace, error) {
	var workspaces []domain.Workspace

	query := `
        SELECT w.id, w.name, w.slug, w.owner_id, w.description, w.created_at 
        FROM workspaces w
        INNER JOIN workspace_members wm ON w.id = wm.workspace_id
        WHERE wm.user_id = $1
        ORDER BY w.created_at DESC
        LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &workspaces, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	if workspaces == nil {
		workspaces = []domain.Workspace{}
	}

	return workspaces, nil
}

func (r *sqlRepository) FindByID(ctx context.Context, workspaceID string, id string) (*domain.Workspace, error) {
	var ws domain.Workspace
	query := `SELECT id, name, slug, owner_id, description, created_at FROM workspaces WHERE id = $1`

	err := r.db.GetContext(ctx, &ws, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ws, nil
}

func (r *sqlRepository) GetMembers(ctx context.Context, workspaceID string) ([]domain.WorkspaceMemberResponse, error) {
	var members []domain.WorkspaceMemberResponse

	query := `
		SELECT 
			w.owner_id AS user_id,
			u.full_name,
			u.email,
			'OWNER' AS role,
			w.created_at AS joined_at
		FROM workspaces w
		INNER JOIN users u ON w.owner_id = u.id
		WHERE w.id = $1

		UNION ALL

		SELECT 
			wm.user_id,
			u.full_name,
			u.email,
			wm.role,
			wm.joined_at
		FROM workspace_members wm
		INNER JOIN users u ON wm.user_id = u.id
		WHERE wm.workspace_id = $1
		ORDER BY joined_at ASC`

	err := r.db.SelectContext(ctx, &members, query, workspaceID)
	if err != nil {
		return nil, err
	}

	if members == nil {
		members = []domain.WorkspaceMemberResponse{}
	}

	return members, nil
}

func (r *sqlRepository) GetMemberRole(ctx context.Context, workspaceID, userID string) (string, error) {
	var role string
	query := `SELECT role FROM workspace_members WHERE workspace_id = $1 AND user_id = $2`
	err := r.db.GetContext(ctx, &role, query, workspaceID, userID)
	if err != nil {
		return "", err
	}
	return role, nil
}
