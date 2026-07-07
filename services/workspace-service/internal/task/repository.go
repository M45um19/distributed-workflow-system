package task

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

func NewRepository(db *sqlx.DB) domain.TaskRepository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, t *domain.Task) error {
	query := `
		INSERT INTO tasks (project_id, title, description, status, priority, assignee_id, deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.Deadline).Scan(&t.ID, &t.CreatedAt)
}

func (r *sqlRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	var t domain.Task
	query := `
		SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.id = $1`
	err := r.db.GetContext(ctx, &t, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *sqlRepository) GetByProjectID(ctx context.Context, projectID string) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.project_id = $1 
		ORDER BY t.created_at DESC`
	err := r.db.SelectContext(ctx, &tasks, query, projectID)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	return tasks, nil
}

func (r *sqlRepository) GetByProjectIDAndStatus(ctx context.Context, projectID string, status string, limit, offset int) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.project_id = $1 AND t.status = $2 
		ORDER BY t.created_at DESC 
		LIMIT $3 OFFSET $4`
	err := r.db.SelectContext(ctx, &tasks, query, projectID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	return tasks, nil
}

func (r *sqlRepository) Update(ctx context.Context, t *domain.Task) error {
	query := `
		UPDATE tasks 
		SET title = $1, description = $2, priority = $3, assignee_id = $4, deadline = $5 
		WHERE id = $6`
	_, err := r.db.ExecContext(ctx, query, t.Title, t.Description, t.Priority, t.AssigneeID, t.Deadline, t.ID)
	return err
}

func (r *sqlRepository) UpdateStatus(ctx context.Context, taskID string, status string) error {
	query := `UPDATE tasks SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, taskID)
	return err
}

func (r *sqlRepository) CreateComment(ctx context.Context, c *domain.TaskComment) error {
	query := `
		INSERT INTO task_comments (task_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, c.TaskID, c.UserID, c.Content).Scan(&c.ID, &c.CreatedAt)
}

func (r *sqlRepository) GetCommentsByTaskID(ctx context.Context, taskID string) ([]domain.TaskComment, error) {
	var comments []domain.TaskComment
	query := `SELECT id, task_id, user_id, content, created_at FROM task_comments WHERE task_id = $1 ORDER BY created_at ASC`
	err := r.db.SelectContext(ctx, &comments, query, taskID)
	if err != nil {
		return nil, err
	}
	if comments == nil {
		comments = []domain.TaskComment{}
	}
	return comments, nil
}

func (r *sqlRepository) GetWorkspaceIDByProjectID(ctx context.Context, projectID string) (string, error) {
	var workspaceID string
	query := `SELECT workspace_id FROM projects WHERE id = $1`
	err := r.db.GetContext(ctx, &workspaceID, query, projectID)
	return workspaceID, err
}

func (r *sqlRepository) GetWorkspaceIDByTaskID(ctx context.Context, taskID string) (string, error) {
	var workspaceID string
	query := `SELECT p.workspace_id FROM tasks t JOIN projects p ON t.project_id = p.id WHERE t.id = $1`
	err := r.db.GetContext(ctx, &workspaceID, query, taskID)
	return workspaceID, err
}
