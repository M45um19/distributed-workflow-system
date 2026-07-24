package task

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type service struct {
	taskRepo                     domain.TaskRepository
	wsRepo                       domain.WorkspaceRepository
	userRepo                     domain.UserRepository
	notificationkafkaWriter      *kafka.Writer
	taskCreatedkafkaWriter       *kafka.Writer
	taskUpdatedkafkaWriter       *kafka.Writer
	taskStatusUpdatedkafkaWriter *kafka.Writer
	taskCache                    domain.TaskCache
	wsCache                      domain.WorkspaceCache
}

func NewService(
	taskRepo domain.TaskRepository,
	wsRepo domain.WorkspaceRepository,
	userRepo domain.UserRepository,
	notificationkafkaWriter *kafka.Writer,
	taskCreatedkafkaWriter *kafka.Writer,
	taskUpdatedkafkaWriter *kafka.Writer,
	taskStatusUpdatedkafkaWriter *kafka.Writer,
	taskCache domain.TaskCache,
	wsCache domain.WorkspaceCache,
) domain.TaskService {
	return &service{
		taskRepo:                     taskRepo,
		wsRepo:                       wsRepo,
		userRepo:                     userRepo,
		notificationkafkaWriter:      notificationkafkaWriter,
		taskCreatedkafkaWriter:       taskCreatedkafkaWriter,
		taskUpdatedkafkaWriter:       taskUpdatedkafkaWriter,
		taskStatusUpdatedkafkaWriter: taskStatusUpdatedkafkaWriter,
		taskCache:                    taskCache,
		wsCache:                      wsCache,
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


func (s *service) sendTaskCreated(ctx context.Context, t *domain.Task) error {
	jsonData, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return s.taskCreatedkafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(t.ID),
		Value: jsonData,
	})
}

func (s *service) sendTaskUpdated(ctx context.Context, t *domain.Task) error {
	jsonData, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return s.taskUpdatedkafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(t.ID),
		Value: jsonData,
	})
}

func (s *service) sendTaskStatusUpdated(ctx context.Context, update domain.TaskStatusUpdate) error {
	jsonData, err := json.Marshal(update)
	if err != nil {
		return err
	}

	return s.taskStatusUpdatedkafkaWriter.WriteMessages(ctx, kafka.Message{
		Key:   []byte(update.TaskID),
		Value: jsonData,
	})
}

func (s *service) CreateTask(ctx context.Context, workspaceID string, projectID string, input domain.TaskCreateInput, userID string) (*domain.Task, error) {
	if err := s.checkAccess(ctx, workspaceID, userID, "ADMIN"); err != nil {
		return nil, err
	}

	taskID, err := uuid.NewV7()
	if err != nil {
		return nil, apperror.InternalServer("failed to generate task ID: " + err.Error())
	}

	var assigneeName *string
	if input.AssigneeID != "" {
		// 1. Try to find assignee in workspace members cache
		members, hit, err := s.wsCache.GetMembers(ctx, workspaceID)
		if err == nil && hit {
			for _, m := range members {
				if m.UserID == input.AssigneeID {
					assigneeName = &m.FullName
					break
				}
			}
		}

		// 2. Fall back to user repo if cache miss or not found in members cache
		if assigneeName == nil {
			u, err := s.userRepo.FindByID(ctx, input.AssigneeID)
			if err == nil && u != nil {
				assigneeName = &u.FullName
			}
		}
	}

	t := &domain.Task{
		ID:           taskID.String(),
		WorkspaceID:  workspaceID,
		ProjectID:    projectID,
		Title:        input.Title,
		Description:  input.Description,
		Status:       "TODO",
		Priority:     input.Priority,
		AssigneeID:   input.AssigneeID,
		AssigneeName: assigneeName,
		Deadline:     input.Deadline,
		CreatedAt:    time.Now(),
	}

	// 1. Add directly to Redis cache for immediate visibility
	_ = s.taskCache.AddTask(ctx, t)

	// 2. Publish event to Kafka task-created topic
	if err := s.sendTaskCreated(ctx, t); err != nil {
		return nil, apperror.InternalServer("failed to publish task created event: " + err.Error())
	}

	// 3. Send Notification to assignee asynchronously if applicable
	if t.AssigneeID != "" && t.AssigneeID != userID {
		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  t.AssigneeID,
			Title:   "Task Assigned",
			Message: fmt.Sprintf("You have been assigned the task '%s'", t.Title),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, t.AssigneeID)
	}

	return t, nil
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

	var assigneeName *string
	if input.AssigneeID != "" {
		if input.AssigneeID == t.AssigneeID {
			assigneeName = t.AssigneeName
		} else {
			// Find new assignee name
			members, hit, err := s.wsCache.GetMembers(ctx, workspaceID)
			if err == nil && hit {
				for _, m := range members {
					if m.UserID == input.AssigneeID {
						assigneeName = &m.FullName
						break
					}
				}
			}
			if assigneeName == nil {
				u, err := s.userRepo.FindByID(ctx, input.AssigneeID)
				if err == nil && u != nil {
					assigneeName = &u.FullName
				}
			}
		}
	}
	t.AssigneeName = assigneeName

	// 1. Update cache for immediate visibility
	_ = s.taskCache.UpdateTaskMeta(ctx, t)

	// 2. Publish updated event to Kafka
	if err := s.sendTaskUpdated(ctx, t); err != nil {
		return nil, apperror.InternalServer("failed to publish task updated event: " + err.Error())
	}

	// 3. Send Notification to assignee asynchronously if applicable
	if t.AssigneeID != "" && t.AssigneeID != userID {
		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  t.AssigneeID,
			Title:   "Task Updated",
			Message: fmt.Sprintf("Assigned task '%s' has been updated", t.Title),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, t.AssigneeID)
	}

	return t, nil
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

	// 1. Update cache for immediate visibility
	_ = s.taskCache.UpdateTaskStatus(ctx, t.ProjectID, taskID, t.Status, status)

	// 2. Publish status update to Kafka
	updatePayload := domain.TaskStatusUpdate{
		WorkspaceID: workspaceID,
		TaskID:      taskID,
		Status:      status,
	}
	if err := s.sendTaskStatusUpdated(ctx, updatePayload); err != nil {
		return apperror.InternalServer("failed to publish task status update event: " + err.Error())
	}

	// 3. Send Notification to assignee asynchronously if applicable
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

func (s *service) GetTaskComments(ctx context.Context, workspaceID string, taskID string, userID string, limit int, cursor string) ([]domain.TaskComment, string, error) {
	if err := s.checkAccess(ctx, workspaceID, userID); err != nil {
		return nil, "", err
	}

	// Fetch limit + 1 comments to determine if there is a next page
	comments, err := s.taskRepo.GetCommentsByTaskID(ctx, workspaceID, taskID, limit+1, cursor)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(comments) > limit {
		nextCursor = comments[limit-1].ID
		comments = comments[:limit]
	}

	return comments, nextCursor, nil
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
