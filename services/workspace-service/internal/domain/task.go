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
	ID          string    `db:"id" json:"id"`
	ProjectID   string    `db:"project_id" json:"project_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	Status      string    `db:"status" json:"status"`
	Priority    string    `db:"priority" json:"priority"`
	AssigneeID  string    `db:"assignee_id" json:"assignee_id"`
	Deadline    time.Time `db:"deadline" json:"deadline"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type TaskComment struct {
	ID        string    `db:"id" json:"id"`
	TaskID    string    `db:"task_id" json:"task_id"`
	UserID    string    `db:"user_id" json:"user_id"`
	Content   string    `db:"content" json:"content"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id string) (*Task, error)
	GetByProjectID(ctx context.Context, projectID string) ([]Task, error)
	Update(ctx context.Context, task *Task) error
	UpdateStatus(ctx context.Context, taskID string, status string) error

	CreateComment(ctx context.Context, comment *TaskComment) error
	GetCommentsByTaskID(ctx context.Context, taskID string) ([]TaskComment, error)
	GetWorkspaceIDByProjectID(ctx context.Context, projectID string) (string, error)
	GetWorkspaceIDByTaskID(ctx context.Context, taskID string) (string, error)
}

type TaskService interface {
	CreateTask(ctx context.Context, projectID string, input TaskCreateInput, userID string) (*Task, error)
	GetTasksByProject(ctx context.Context, projectID string, userID string) ([]Task, error)
	UpdateFullTask(ctx context.Context, taskID string, input TaskUpdateInput, userID string) (*Task, error)
	UpdateTaskStatus(ctx context.Context, taskID string, status string, userID string) error

	AddComment(ctx context.Context, taskID string, input CommentCreateInput, userID string) (*TaskComment, error)
	GetTaskComments(ctx context.Context, taskID string, userID string) ([]TaskComment, error)
}
