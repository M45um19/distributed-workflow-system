package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/segmentio/kafka-go"
)

type service struct {
	taskRepo                domain.TaskRepository
	wsRepo                  domain.WorkspaceRepository
	userRepo                domain.UserRepository
	notificationkafkaWriter *kafka.Writer
}

func NewService(taskRepo domain.TaskRepository, wsRepo domain.WorkspaceRepository, userRepo domain.UserRepository, notificationkafkaWriter *kafka.Writer) domain.TaskService {
	return &service{
		taskRepo:                taskRepo,
		wsRepo:                  wsRepo,
		userRepo:                userRepo,
		notificationkafkaWriter: notificationkafkaWriter,
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

	createdTask, err := s.taskRepo.FindByID(ctx, t.ID)
	if err == nil && createdTask != nil && createdTask.AssigneeID != "" && createdTask.AssigneeID != userID {
		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  createdTask.AssigneeID,
			Title:   "Task Assigned",
			Message: fmt.Sprintf("You have been assigned the task '%s'", createdTask.Title),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, createdTask.AssigneeID)
	}

	return createdTask, err
}

func (s *service) GetTasksByProject(ctx context.Context, projectID string, userID string, statuses []string, limit, page int) (map[string][]domain.Task, error) {
	workspaceID, err := s.taskRepo.GetWorkspaceIDByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	validStatuses := map[string]bool{
		"TODO":        true,
		"IN_PROGRESS": true,
		"REVIEW":      true,
		"DONE":        true,
	}

	var filterStatuses []string
	if len(statuses) == 0 {
		filterStatuses = []string{"TODO", "IN_PROGRESS", "REVIEW", "DONE"}
	} else {
		for _, st := range statuses {
			if validStatuses[st] {
				filterStatuses = append(filterStatuses, st)
			}
		}
	}

	offset := (page - 1) * limit
	result := make(map[string][]domain.Task)

	for _, st := range filterStatuses {
		tasks, err := s.taskRepo.GetByProjectIDAndStatus(ctx, projectID, st, limit, offset)
		if err != nil {
			return nil, err
		}
		result[st] = tasks
	}

	return result, nil
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

	updatedTask, err := s.taskRepo.FindByID(ctx, t.ID)
	if err == nil && updatedTask != nil && updatedTask.AssigneeID != "" && updatedTask.AssigneeID != userID {
		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  updatedTask.AssigneeID,
			Title:   "Task Updated",
			Message: fmt.Sprintf("Assigned task '%s' has been updated", updatedTask.Title),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, updatedTask.AssigneeID)
	}

	return updatedTask, err
}

func (s *service) UpdateTaskStatus(ctx context.Context, taskID string, status string, userID string) error {
	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}
	if t == nil {
		return apperror.NotFound("Task not found")
	}

	// Check if user is the assigned member, workspace owner, or admin
	if t.AssigneeID != userID {
		workspaceID, err := s.taskRepo.GetWorkspaceIDByTaskID(ctx, taskID)
		if err != nil {
			return err
		}
		if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
			return apperror.Forbidden("Only the assigned member, admin, or owner can change this task status")
		}
	}

	if err := s.taskRepo.UpdateStatus(ctx, taskID, status); err != nil {
		return err
	}

	if t.AssigneeID != "" && t.AssigneeID != userID {
		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  t.AssigneeID,
			Title:   "Task Status Updated",
			Message: fmt.Sprintf("The status of your task '%s' was updated to %s", t.Title, status),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, t.AssigneeID)
	}

	return nil
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

	t, err := s.taskRepo.FindByID(ctx, taskID)
	if err == nil && t != nil && t.AssigneeID != "" && t.AssigneeID != userID {
		commenterName := "Someone"
		commenter, err := s.userRepo.FindByID(ctx, userID)
		if err == nil && commenter != nil {
			commenterName = commenter.FullName
		}

		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  t.AssigneeID,
			Title:   "New Comment on Your Task",
			Message: fmt.Sprintf("%s commented on your assigned task '%s'", commenterName, t.Title),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, t.AssigneeID)
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

func (s *service) sendNotification(ctx context.Context, payload domain.NotificationEventPayload, key string) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal kafka notification payload: %v", err)
		return
	}
	err = s.notificationkafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: jsonData,
	})
	if err != nil {
		log.Printf("Kafka failed to send notification: %v", err)
	} else {
		log.Printf("Successfully produced message to send-notification for: %s", key)
	}
}
