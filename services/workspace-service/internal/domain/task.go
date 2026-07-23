package domain

import (
	"context"
	"time"
)

type TaskCreateInput struct {
	Title       string    `json:"title" binding:"required,min=3,max=100"`
	Description string    `json:"description" binding:"max=1000"`
	Priority    string    `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH URGENT"`
	AssigneeID  string    `json:"assignee_id"`
	Deadline    time.Time `json:"deadline"`
}

type TaskUpdateInput struct {
	Title       string    `json:"title" binding:"required,min=3,max=100"`
	Description string    `json:"description" binding:"max=1000"`
	Priority    string    `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH URGENT"`
	AssigneeID  string    `json:"assignee_id"`
	Deadline    time.Time `json:"deadline"`
}

type TaskStatusUpdateInput struct {
	Status string `json:"status" binding:"required,oneof=TODO IN_PROGRESS REVIEW DONE"`
}

type CommentCreateInput struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type Task struct {
	ID           string    `db:"id" json:"id"`
	WorkspaceID  string    `db:"workspace_id" json:"workspace_id"`
	ProjectID    string    `db:"project_id" json:"project_id"`
	Title        string    `db:"title" json:"title"`
	Description  string    `db:"description" json:"description"`
	Status       string    `db:"status" json:"status"`
	Priority     string    `db:"priority" json:"priority"`
	AssigneeID   string    `db:"assignee_id" json:"assignee_id"`
	AssigneeName *string   `db:"assignee_name" json:"assignee_name"`
	Deadline     time.Time `db:"deadline" json:"deadline"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type TaskComment struct {
	ID          string    `db:"id" json:"id"`
	WorkspaceID string    `db:"workspace_id" json:"workspace_id"`
	TaskID      string    `db:"task_id" json:"task_id"`
	UserID      string    `db:"user_id" json:"user_id"`
	Content     string    `db:"content" json:"content"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type TaskRepository interface {
	Create(ctx context.Context, workspaceID string, task *Task) error
	FindByID(ctx context.Context, workspaceID string, id string) (*Task, error)
	GetByProjectID(ctx context.Context, workspaceID string, projectID string) ([]Task, error)
	GetByProjectIDAndStatus(ctx context.Context, workspaceID string, projectID string, status string, limit, offset int) ([]Task, error)
	GetByProjectIDAndStatusCursor(ctx context.Context, workspaceID string, projectID string, status string, limit int, cursor string) ([]Task, error)
	Update(ctx context.Context, workspaceID string, task *Task) error
	UpdateStatus(ctx context.Context, workspaceID string, taskID string, status string) error

	CreateComment(ctx context.Context, workspaceID string, comment *TaskComment) error
	GetCommentsByTaskID(ctx context.Context, workspaceID string, taskID string, limit int, cursor string) ([]TaskComment, error)
	BulkCreate(ctx context.Context, tasks []Task) error
	GetWorkspaceIDByProjectID(ctx context.Context, workspaceID string, projectID string) (string, error)
	GetWorkspaceIDByTaskID(ctx context.Context, workspaceID string, taskID string) (string, error)
}

type TaskService interface {
	CreateTask(ctx context.Context, workspaceID string, projectID string, input TaskCreateInput, userID string) (*Task, error)
	GetTasksByProject(ctx context.Context, workspaceID string, projectID string, userID string, statuses []string, limit int, cursor string) (map[string][]Task, map[string]string, error)
	UpdateFullTask(ctx context.Context, workspaceID string, taskID string, input TaskUpdateInput, userID string) (*Task, error)
	UpdateTaskStatus(ctx context.Context, workspaceID string, taskID string, status string, userID string) error

	AddComment(ctx context.Context, workspaceID string, taskID string, input CommentCreateInput, userID string) (*TaskComment, error)
	GetTaskComments(ctx context.Context, workspaceID string, taskID string, userID string, limit int, cursor string) ([]TaskComment, string, error)
}

type TaskCache interface {
	AddTask(ctx context.Context, task *Task) error
	GetTaskIDs(ctx context.Context, projectID string, status string, limit int, cursor string) ([]string, []float64, bool, error)
	GetTaskMetas(ctx context.Context, taskIDs []string) ([]Task, []string, error)
	SetTaskMeta(ctx context.Context, task *Task) error
	SetColumnCache(ctx context.Context, projectID string, status string, tasks []Task) error
	UpdateTaskMeta(ctx context.Context, task *Task) error
	UpdateTaskStatus(ctx context.Context, projectID string, taskID string, oldStatus string, newStatus string) error
	InvalidateTasks(ctx context.Context, projectID string) error
}

