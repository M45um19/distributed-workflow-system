package task

import (
	"context"
	"database/sql"
	"errors"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
)

type service struct {
	taskRepo domain.TaskRepository
	wsRepo   domain.WorkspaceRepository
}

func NewService(taskRepo domain.TaskRepository, wsRepo domain.WorkspaceRepository) domain.TaskService {
	return &service{
		taskRepo: taskRepo,
		wsRepo:   wsRepo,
	}
}

func (s *service) checkAccess(ctx context.Context, workspaceID, userID string, allowedRoles ...string) error {
	ws, err := s.wsRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return err
	}
	if ws == nil {
		return apperror.NotFound("Workspace not found")
	}

	if ws.OwnerID == userID {
		return nil
	}

	role, err := s.wsRepo.GetMemberRole(ctx, workspaceID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.Forbidden("You are not a member of this workspace")
		}
		return err
	}

	if len(allowedRoles) == 0 {
		return nil
	}

	for _, r := range allowedRoles {
		if r == role {
			return nil
		}
	}

	return apperror.Forbidden("You do not have enough permissions for this action")
}

func (s *service) CreateTask(ctx context.Context, projectID string, input domain.TaskCreateInput, userID string) (*domain.Task, error) {
	workspaceID, err := s.taskRepo.GetWorkspaceIDByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Project not found")
		}
		return nil, err
	}

	if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
		return nil, err
	}

	t := &domain.Task{
		ProjectID:   projectID,
		Title:       input.Title,
		Description: input.Description,
		Status:      "TODO",
		Priority:    input.Priority,
		AssigneeID:  input.AssigneeID,
		Deadline:    input.Deadline,
	}

	if err := s.taskRepo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) GetTasksByProject(ctx context.Context, projectID string, userID string) ([]domain.Task, error) {
	workspaceID, err := s.taskRepo.GetWorkspaceIDByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	return s.taskRepo.GetByProjectID(ctx, projectID)
}

func (s *service) UpdateFullTask(ctx context.Context, taskID string, input domain.TaskUpdateInput, userID string) (*domain.Task, error) {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, apperror.NotFound("Task not found")
	}

	workspaceID, err := s.taskRepo.GetWorkspaceIDByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
		return nil, err
	}

	t.Title = input.Title
	t.Description = input.Description
	t.Priority = input.Priority
	t.AssigneeID = input.AssigneeID
	t.Deadline = input.Deadline

	if err := s.taskRepo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *service) UpdateTaskStatus(ctx context.Context, taskID string, status string, userID string) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if t == nil {
		return apperror.NotFound("Task not found")
	}

	// Check if user is the assigned member
	if t.AssigneeID == nil || *t.AssigneeID != userID {
		return apperror.Forbidden("Only the assigned member can change this task status")
	}

	return s.taskRepo.UpdateStatus(ctx, taskID, status)
}

func (s *service) AddComment(ctx context.Context, taskID string, input domain.CommentCreateInput, userID string) (*domain.TaskComment, error) {
	workspaceID, err := s.taskRepo.GetWorkspaceIDByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("Task not found")
		}
		return nil, err
	}

	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	comment := &domain.TaskComment{
		TaskID:  taskID,
		UserID:  userID,
		Content: input.Content,
	}

	if err := s.taskRepo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *service) GetTaskComments(ctx context.Context, taskID string, userID string) ([]domain.TaskComment, error) {
	workspaceID, err := s.taskRepo.GetWorkspaceIDByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	return s.taskRepo.GetCommentsByTaskID(ctx, taskID)
}
