package task

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/jmoiron/sqlx"
)

type sqlRepository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) domain.TaskRepository {
	return &sqlRepository{db: db}
}

func (r *sqlRepository) Create(ctx context.Context, workspaceID string, t *domain.Task) error {
	t.WorkspaceID = workspaceID
	query := `
		INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority, assignee_id, deadline)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, t.ID, t.WorkspaceID, t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.AssigneeID, t.Deadline).Scan(&t.ID, &t.CreatedAt)
}

func (r *sqlRepository) FindByID(ctx context.Context, workspaceID string, id string) (*domain.Task, error) {
	var t domain.Task
	query := `
		SELECT t.id, t.workspace_id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.id = $1 AND t.workspace_id = $2`
	err := r.db.GetContext(ctx, &t, query, id, workspaceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *sqlRepository) GetByProjectID(ctx context.Context, workspaceID string, projectID string) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT t.id, t.workspace_id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.project_id = $1 AND t.workspace_id = $2 
		ORDER BY t.created_at DESC`
	err := r.db.SelectContext(ctx, &tasks, query, projectID, workspaceID)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	return tasks, nil
}

func (r *sqlRepository) GetByProjectIDAndStatus(ctx context.Context, workspaceID string, projectID string, status string, limit, offset int) ([]domain.Task, error) {
	var tasks []domain.Task
	query := `
		SELECT t.id, t.workspace_id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
		FROM tasks t
		LEFT JOIN users u ON t.assignee_id = u.id
		WHERE t.project_id = $1 AND t.status = $2 AND t.workspace_id = $3
		ORDER BY t.created_at DESC 
		LIMIT $4 OFFSET $5`
	err := r.db.SelectContext(ctx, &tasks, query, projectID, status, workspaceID, limit, offset)
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	return tasks, nil
}

func (r *sqlRepository) GetByProjectIDAndStatusCursor(ctx context.Context, workspaceID string, projectID string, status string, limit int, cursor string) ([]domain.Task, error) {
	var tasks []domain.Task
	var err error
	if cursor == "" {
		query := `
			SELECT t.id, t.workspace_id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
			FROM tasks t
			LEFT JOIN users u ON t.assignee_id = u.id
			WHERE t.project_id = $1 AND t.status = $2 AND t.workspace_id = $3
			ORDER BY t.created_at DESC, t.id DESC
			LIMIT $4`
		err = r.db.SelectContext(ctx, &tasks, query, projectID, status, workspaceID, limit)
	} else {
		query := `
			SELECT t.id, t.workspace_id, t.project_id, t.title, t.description, t.status, t.priority, t.assignee_id, u.full_name AS assignee_name, t.deadline, t.created_at 
			FROM tasks t
			LEFT JOIN users u ON t.assignee_id = u.id
			WHERE t.project_id = $1 AND t.status = $2 AND t.workspace_id = $3 AND t.id < $4
			ORDER BY t.created_at DESC, t.id DESC
			LIMIT $5`
		err = r.db.SelectContext(ctx, &tasks, query, projectID, status, workspaceID, cursor, limit)
	}
	if err != nil {
		return nil, err
	}
	if tasks == nil {
		tasks = []domain.Task{}
	}
	return tasks, nil
}

func (r *sqlRepository) Update(ctx context.Context, workspaceID string, t *domain.Task) error {
	query := `
		UPDATE tasks 
		SET title = $1, description = $2, priority = $3, assignee_id = $4, deadline = $5 
		WHERE id = $6 AND workspace_id = $7`
	_, err := r.db.ExecContext(ctx, query, t.Title, t.Description, t.Priority, t.AssigneeID, t.Deadline, t.ID, workspaceID)
	return err
}

func (r *sqlRepository) UpdateStatus(ctx context.Context, workspaceID string, taskID string, status string) error {
	query := `UPDATE tasks SET status = $1 WHERE id = $2 AND workspace_id = $3`
	_, err := r.db.ExecContext(ctx, query, status, taskID, workspaceID)
	return err
}

func (r *sqlRepository) CreateComment(ctx context.Context, workspaceID string, c *domain.TaskComment) error {
	c.WorkspaceID = workspaceID
	query := `
		INSERT INTO task_comments (id, workspace_id, task_id, user_id, content)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, c.ID, c.WorkspaceID, c.TaskID, c.UserID, c.Content).Scan(&c.ID, &c.CreatedAt)
}

func (r *sqlRepository) GetCommentsByTaskID(ctx context.Context, workspaceID string, taskID string, limit int, cursor string) ([]domain.TaskComment, error) {
	var comments []domain.TaskComment
	var err error
	if cursor == "" {
		query := `
			SELECT id, workspace_id, task_id, user_id, content, created_at 
			FROM task_comments 
			WHERE task_id = $1 AND workspace_id = $2 
			ORDER BY id ASC 
			LIMIT $3`
		err = r.db.SelectContext(ctx, &comments, query, taskID, workspaceID, limit)
	} else {
		query := `
			SELECT id, workspace_id, task_id, user_id, content, created_at 
			FROM task_comments 
			WHERE task_id = $1 AND workspace_id = $2 AND id > $3
			ORDER BY id ASC 
			LIMIT $4`
		err = r.db.SelectContext(ctx, &comments, query, taskID, workspaceID, cursor, limit)
	}
	if err != nil {
		return nil, err
	}
	if comments == nil {
		comments = []domain.TaskComment{}
	}
	return comments, nil
}

func (r *sqlRepository) GetWorkspaceIDByProjectID(ctx context.Context, workspaceID string, projectID string) (string, error) {
	var wID string
	query := `SELECT workspace_id FROM projects WHERE id = $1 AND workspace_id = $2`
	err := r.db.GetContext(ctx, &wID, query, projectID, workspaceID)
	return wID, err
}

func (r *sqlRepository) GetWorkspaceIDByTaskID(ctx context.Context, workspaceID string, taskID string) (string, error) {
	var wID string
	query := `SELECT p.workspace_id FROM tasks t JOIN projects p ON t.project_id = p.id AND t.workspace_id = p.workspace_id WHERE t.id = $1 AND t.workspace_id = $2`
	err := r.db.GetContext(ctx, &wID, query, taskID, workspaceID)
	return wID, err
}

func (r *sqlRepository) BulkCreate(ctx context.Context, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	numCols := 10
	query := `INSERT INTO tasks (id, workspace_id, project_id, title, description, status, priority, assignee_id, deadline, created_at) VALUES `
	
	vals := make([]interface{}, 0, len(tasks)*numCols)
	for i, t := range tasks {
		if i > 0 {
			query += ", "
		}
		offset := i * numCols
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10)
		
		var assigneeID interface{}
		if t.AssigneeID == "" {
			assigneeID = nil
		} else {
			assigneeID = t.AssigneeID
		}

		var deadline interface{}
		if t.Deadline.IsZero() {
			deadline = nil
		} else {
			deadline = t.Deadline
		}

		vals = append(vals, t.ID, t.WorkspaceID, t.ProjectID, t.Title, t.Description, t.Status, t.Priority, assigneeID, deadline, t.CreatedAt)
	}
	
	query += " ON CONFLICT (id, workspace_id) DO NOTHING"

	_, err := r.db.ExecContext(ctx, query, vals...)
	return err
}
