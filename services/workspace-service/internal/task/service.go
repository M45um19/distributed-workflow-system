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
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type service struct {
	taskRepo                domain.TaskRepository
	wsRepo                  domain.WorkspaceRepository
	userRepo                domain.UserRepository
	notificationkafkaWriter *kafka.Writer
	taskCache               domain.TaskCache
	wsCache                 domain.WorkspaceCache
}

func NewService(taskRepo domain.TaskRepository, wsRepo domain.WorkspaceRepository, userRepo domain.UserRepository, notificationkafkaWriter *kafka.Writer, taskCache domain.TaskCache, wsCache domain.WorkspaceCache) domain.TaskService {
	return &service{
		taskRepo:                taskRepo,
		wsRepo:                  wsRepo,
		userRepo:                userRepo,
		notificationkafkaWriter: notificationkafkaWriter,
		taskCache:               taskCache,
		wsCache:                 wsCache,
	}
}

func (s *service) checkAccess(ctx context.Context, workspaceID, userID string, allowedRoles ...string) error {
	var ws *domain.Workspace
	var err error

	// 1. Try to fetch workspace metadata from cache
	ws, err = s.wsCache.GetWorkspaceMeta(ctx, workspaceID)
	if err != nil || ws == nil {
		// Cache miss or error: fetch from DB
		ws, err = s.wsRepo.FindByID(ctx, workspaceID, workspaceID)
		if err != nil {
			return err
		}
		if ws == nil {
			return apperror.NotFound("Workspace not found")
		}
		// Populate cache
		_ = s.wsCache.SetWorkspaceMeta(ctx, ws)
	}

	if ws.OwnerID == userID {
		return nil
	}

	// 2. Try to fetch user's member role in the workspace from cache
	role, exists, err := s.wsCache.GetMemberRole(ctx, workspaceID, userID)
	if err != nil || !exists {
		// Cache miss or error: fetch from DB
		role, err = s.wsRepo.GetMemberRole(ctx, workspaceID, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// Cache negative lookup (empty role) to prevent redundant DB calls
				_ = s.wsCache.SetMemberRole(ctx, workspaceID, userID, "")
				return apperror.Forbidden("You are not a member of this workspace")
			}
			return err
		}
		// Populate cache
		_ = s.wsCache.SetMemberRole(ctx, workspaceID, userID, role)
	} else if role == "" {
		// Cached negative lookup
		return apperror.Forbidden("You are not a member of this workspace")
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

func (s *service) getTaskByID(ctx context.Context, workspaceID string, taskID string) (*domain.Task, error) {
	tasks, missingIDs, err := s.taskCache.GetTaskMetas(ctx, []string{taskID})
	if err == nil && len(missingIDs) == 0 && len(tasks) > 0 {
		t := &tasks[0]
		if t.WorkspaceID == workspaceID {
			return t, nil
		}
	}

	t, err := s.taskRepo.FindByID(ctx, workspaceID, taskID)
	if err != nil {
		return nil, err
	}
	if t != nil {
		_ = s.taskCache.SetTaskMeta(ctx, t)
	}
	return t, nil
}


func (s *service) CreateTask(ctx context.Context, workspaceID string, projectID string, input domain.TaskCreateInput, userID string) (*domain.Task, error) {
	if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
		return nil, err
	}

	taskID, err := uuid.NewV7()
	if err != nil {
		return nil, apperror.InternalServer("failed to generate task ID: " + err.Error())
	}

	t := &domain.Task{
		ID:          taskID.String(),
		WorkspaceID: workspaceID,
		ProjectID:   projectID,
		Title:       input.Title,
		Description: input.Description,
		Status:      "TODO",
		Priority:    input.Priority,
		AssigneeID:  input.AssigneeID,
		Deadline:    input.Deadline,
	}

	if err := s.taskRepo.Create(ctx, workspaceID, t); err != nil {
		return nil, err
	}

	createdTask, err := s.taskRepo.FindByID(ctx, workspaceID, t.ID)
	if err == nil && createdTask != nil {
		_ = s.taskCache.AddTask(ctx, createdTask)

		if createdTask.AssigneeID != "" && createdTask.AssigneeID != userID {
			notificationPayload := domain.NotificationEventPayload{
				Channel: "IN_APP",
				UserID:  createdTask.AssigneeID,
				Title:   "Task Assigned",
				Message: fmt.Sprintf("You have been assigned the task '%s'", createdTask.Title),
				Type:    "INFO",
			}
			s.sendNotification(ctx, notificationPayload, createdTask.AssigneeID)
		}
	}

	return createdTask, err
}

func (s *service) GetTasksByProject(ctx context.Context, workspaceID string, projectID string, userID string, statuses []string, limit int, cursor string) (map[string][]domain.Task, map[string]string, error) {
	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, nil, err
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

	result := make(map[string][]domain.Task)
	nextCursors := make(map[string]string)

	for _, st := range filterStatuses {
		taskIDs, _, exists, err := s.taskCache.GetTaskIDs(ctx, projectID, st, limit, cursor)
		if err != nil || !exists {
			dbTasks, err := s.taskRepo.GetByProjectIDAndStatusCursor(ctx, workspaceID, projectID, st, limit, cursor)
			if err != nil {
				return nil, nil, err
			}
			result[st] = dbTasks

			_ = s.taskCache.SetColumnCache(ctx, projectID, st, dbTasks)

			if len(dbTasks) == limit {
				nextCursors[st] = dbTasks[len(dbTasks)-1].ID
			} else {
				nextCursors[st] = ""
			}
			continue
		}

		if len(taskIDs) == 0 {
			result[st] = []domain.Task{}
			nextCursors[st] = ""
			continue
		}

		cachedTasks, missingIDs, err := s.taskCache.GetTaskMetas(ctx, taskIDs)
		if err != nil {
			return nil, nil, err
		}

		var orderedTasks []domain.Task
		if len(missingIDs) > 0 {
			missingTasksMap := make(map[string]domain.Task)
			for _, id := range missingIDs {
				dbTask, err := s.taskRepo.FindByID(ctx, workspaceID, id)
				if err == nil && dbTask != nil {
					missingTasksMap[id] = *dbTask
					_ = s.taskCache.SetTaskMeta(ctx, dbTask)
				}
			}

			orderedTasks = make([]domain.Task, 0, len(taskIDs))
			for _, id := range taskIDs {
				found := false
				for _, ct := range cachedTasks {
					if ct.ID == id {
						orderedTasks = append(orderedTasks, ct)
						found = true
						break
					}
				}
				if !found {
					if dbTask, ok := missingTasksMap[id]; ok {
						orderedTasks = append(orderedTasks, dbTask)
					}
				}
			}
		} else {
			orderedTasks = make([]domain.Task, len(taskIDs))
			taskMap := make(map[string]domain.Task)
			for _, t := range cachedTasks {
				taskMap[t.ID] = t
			}
			for i, id := range taskIDs {
				orderedTasks[i] = taskMap[id]
			}
		}

		result[st] = orderedTasks

		if len(taskIDs) == limit {
			nextCursors[st] = taskIDs[len(taskIDs)-1]
		} else {
			nextCursors[st] = ""
		}
	}

	return result, nextCursors, nil
}

func (s *service) UpdateFullTask(ctx context.Context, workspaceID string, taskID string, input domain.TaskUpdateInput, userID string) (*domain.Task, error) {
	t, err := s.getTaskByID(ctx, workspaceID, taskID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, apperror.NotFound("Task not found")
	}

	if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
		return nil, err
	}

	t.Title = input.Title
	t.Description = input.Description
	t.Priority = input.Priority
	t.AssigneeID = input.AssigneeID
	t.Deadline = input.Deadline

	if err := s.taskRepo.Update(ctx, workspaceID, t); err != nil {
		return nil, err
	}

	updatedTask, err := s.taskRepo.FindByID(ctx, workspaceID, t.ID)
	if err == nil && updatedTask != nil {
		_ = s.taskCache.UpdateTaskMeta(ctx, updatedTask)

		if updatedTask.AssigneeID != "" && updatedTask.AssigneeID != userID {
			notificationPayload := domain.NotificationEventPayload{
				Channel: "IN_APP",
				UserID:  updatedTask.AssigneeID,
				Title:   "Task Updated",
				Message: fmt.Sprintf("Assigned task '%s' has been updated", updatedTask.Title),
				Type:    "INFO",
			}
			s.sendNotification(ctx, notificationPayload, updatedTask.AssigneeID)
		}
	}

	return updatedTask, err
}

func (s *service) UpdateTaskStatus(ctx context.Context, workspaceID string, taskID string, status string, userID string) error {
	t, err := s.getTaskByID(ctx, workspaceID, taskID)
	if err != nil {
		return err
	}
	if t == nil {
		return apperror.NotFound("Task not found")
	}

	if t.AssigneeID != userID {
		if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
			return apperror.Forbidden("Only the assigned member, admin, or owner can change this task status")
		}
	}

	if err := s.taskRepo.UpdateStatus(ctx, workspaceID, taskID, status); err != nil {
		return err
	}

	_ = s.taskCache.UpdateTaskStatus(ctx, t.ProjectID, taskID, t.Status, status)

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

func (s *service) AddComment(ctx context.Context, workspaceID string, taskID string, input domain.CommentCreateInput, userID string) (*domain.TaskComment, error) {
	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	commentID, err := uuid.NewV7()
	if err != nil {
		return nil, apperror.InternalServer("failed to generate comment ID: " + err.Error())
	}

	comment := &domain.TaskComment{
		ID:          commentID.String(),
		WorkspaceID: workspaceID,
		TaskID:      taskID,
		UserID:      userID,
		Content:     input.Content,
	}

	if err := s.taskRepo.CreateComment(ctx, workspaceID, comment); err != nil {
		return nil, err
	}

	t, err := s.getTaskByID(ctx, workspaceID, taskID)
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

func (s *service) GetTaskComments(ctx context.Context, workspaceID string, taskID string, userID string) ([]domain.TaskComment, error) {
	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, err
	}

	return s.taskRepo.GetCommentsByTaskID(ctx, workspaceID, taskID)
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
