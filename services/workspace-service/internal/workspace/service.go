package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/temporal/workflow"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"go.temporal.io/sdk/client"
)

type service struct {
	wsRepo                  domain.WorkspaceRepository
	userRepo                domain.UserRepository
	temporalClient          client.Client
	notificationkafkaWriter *kafka.Writer
	frontendURL             string
	cache                   domain.WorkspaceCache
}

func NewService(
	wsRepo domain.WorkspaceRepository,
	userRepo domain.UserRepository,
	tempClient client.Client,
	notificationkafkaWriter *kafka.Writer,
	frontendURL string,
	cache domain.WorkspaceCache,
) domain.WorkspaceService {
	if !strings.HasPrefix(frontendURL, "http://") && !strings.HasPrefix(frontendURL, "https://") {
		frontendURL = "http://" + frontendURL
	}
	return &service{
		wsRepo:                  wsRepo,
		userRepo:                userRepo,
		temporalClient:          tempClient,
		notificationkafkaWriter: notificationkafkaWriter,
		frontendURL:             frontendURL,
		cache:                   cache,
	}
}

func (s *service) CreateWorkspace(ctx context.Context, input domain.WorkspaceCreateInput, ownerID string) (*domain.Workspace, error) {
	exists, _ := s.wsRepo.FindBySlug(ctx, "", input.Slug)
	if exists != nil {
		return nil, apperror.BadRequest("workspace with this slug already exists")
	}

	wsID, err := uuid.NewV7()
	if err != nil {
		return nil, apperror.InternalServer("failed to generate workspace ID: " + err.Error())
	}

	ws := &domain.Workspace{
		ID:          wsID.String(),
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		OwnerID:     ownerID,
	}

	if err := s.wsRepo.Create(ctx, ws.ID, ws); err != nil {
		return nil, err
	}

	// Update Cache
	if cacheErr := s.cache.SetWorkspaceMeta(ctx, ws); cacheErr != nil {
		log.Printf("[Cache Error] CreateWorkspace SetWorkspaceMeta failed: %v", cacheErr)
	}
	if cacheErr := s.cache.AddOwnedWorkspaceID(ctx, ownerID, ws.ID); cacheErr != nil {
		log.Printf("[Cache Error] CreateWorkspace AddOwnedWorkspaceID failed: %v", cacheErr)
	}

	return ws, nil
}

func sortWorkspacesDesc(workspaces []domain.Workspace) {
	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].ID > workspaces[j].ID
	})
}

func paginateWorkspaces(workspaces []domain.Workspace, cursor string, limit int) []domain.Workspace {
	var filtered []domain.Workspace
	for _, ws := range workspaces {
		if cursor == "" || ws.ID < cursor {
			filtered = append(filtered, ws)
		}
	}

	if len(filtered) > limit {
		return filtered[:limit]
	}
	return filtered
}

func (s *service) GetWorkspacesByOwner(ctx context.Context, ownerId string, limit int, cursor string) ([]domain.Workspace, error) {
	ids, exists, err := s.cache.GetOwnedWorkspaceIDs(ctx, ownerId, limit, cursor)
	if err != nil {
		log.Printf("[Cache Error] GetOwnedWorkspaceIDs failed: %v", err)
	}

	var workspaces []domain.Workspace

	if !exists {
		// Cache Miss: Query all workspaces from DB to populate the cache
		allDBWorkspaces, err := s.wsRepo.GetByOwnerID(ctx, "", ownerId, 10000, "")
		if err != nil {
			return nil, err
		}

		var dbIDs []string
		for _, ws := range allDBWorkspaces {
			dbIDs = append(dbIDs, ws.ID)
			if cacheErr := s.cache.SetWorkspaceMeta(ctx, &ws); cacheErr != nil {
				log.Printf("[Cache Error] SetWorkspaceMeta failed: %v", cacheErr)
			}
		}

		if cacheErr := s.cache.SetOwnedWorkspaceIDs(ctx, ownerId, dbIDs); cacheErr != nil {
			log.Printf("[Cache Error] SetOwnedWorkspaceIDs failed: %v", cacheErr)
		}

		workspaces = allDBWorkspaces
		sortWorkspacesDesc(workspaces)
		return paginateWorkspaces(workspaces, cursor, limit), nil
	}

	if len(ids) == 0 {
		return []domain.Workspace{}, nil
	}

	// Cache Hit: fetch metadata for only the returned paginated IDs!
	cachedMetas, missingIDs, err := s.cache.GetWorkspaceMetas(ctx, ids)
	if err != nil {
		log.Printf("[Cache Error] GetWorkspaceMetas failed: %v", err)
	}

	workspaces = cachedMetas

	if len(missingIDs) > 0 {
		for _, id := range missingIDs {
			ws, err := s.wsRepo.FindByID(ctx, id, id)
			if err != nil {
				return nil, err
			}
			if ws != nil {
				workspaces = append(workspaces, *ws)
				if cacheErr := s.cache.SetWorkspaceMeta(ctx, ws); cacheErr != nil {
					log.Printf("[Cache Error] SetWorkspaceMeta failed: %v", cacheErr)
				}
			}
		}
	}

	sortWorkspacesDesc(workspaces)
	return workspaces, nil
}

func (s *service) InviteUser(ctx context.Context, input domain.WorkspaceInviteRequest) (*domain.WorkspaceInviteResponse, error) {

	input.Token = uuid.New().String()

	inviteID, err := uuid.NewV7()
	if err != nil {
		return nil, apperror.InternalServer("failed to generate invitation ID: " + err.Error())
	}

	invite := &domain.WorkspaceInvitation{
		ID:          inviteID.String(),
		WorkspaceID: input.WorkspaceID,
		InviterID:   input.InviterID,
		Email:       input.Email,
		Role:        input.Role,
		Token:       input.Token,
		Status:      "PENDING",
		ExpiresAt:   time.Now().Add(time.Hour * 24 * 14),
	}

	if err := s.wsRepo.CreateInvite(ctx, invite.WorkspaceID, invite); err != nil {
		log.Printf("Failed to save invitation to DB: %v", err)
		return nil, apperror.InternalServer("Could not process invitation")
	}

	// Try to find if the invited user is registered in the system
	invitedUser, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err == nil && invitedUser != nil {
		notificationPayload := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  invitedUser.ID,
			Title:   "Workspace Invitation",
			Message: fmt.Sprintf("You have been invited to join the workspace as a %s", input.Role),
			Type:    "INFO",
		}
		s.sendNotification(ctx, notificationPayload, invitedUser.ID)
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("invite-user-%s", input.Email),
		TaskQueue: "workspace-task-queue",
	}

	wfInput := workflow.WorkspaceInviteRequest{
		Email: input.Email,
		Role:  input.Role,
		Token: input.Token,
	}
	log.Println(wfInput)
	we, err := s.temporalClient.ExecuteWorkflow(ctx, workflowOptions, workflow.InviteUserWorkflow, wfInput)
	if err != nil {
		log.Printf("Temporal Execution Failed: %v", err)
		return nil, apperror.InternalServer("Temporal error: " + err.Error())
	}

	log.Printf("Started workflow. WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())
	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", s.frontendURL, input.Token)
	return &domain.WorkspaceInviteResponse{InviteURL: inviteLink}, nil
}

func (s *service) AcceptInvitation(ctx context.Context, token string, loggedInUserID string) error {
	invite, err := s.wsRepo.FindInviteByToken(ctx, "", token)
	if err != nil {
		return apperror.NotFound("Invitation link is invalid or does not exist")
	}

	if invite.Status != "PENDING" {
		return apperror.BadRequest("This invitation has already been " + invite.Status)
	}

	if time.Now().After(invite.ExpiresAt) {
		_ = s.wsRepo.UpdateInviteStatus(ctx, invite.WorkspaceID, invite.ID, "EXPIRED")
		return apperror.BadRequest("This invitation link has expired")
	}

	currentUser, err := s.userRepo.FindByID(ctx, loggedInUserID)
	if err != nil {
		return apperror.NotFound("Logged in user not found in the system")
	}

	if invite.Email != currentUser.Email {
		return apperror.Forbidden("This invitation was sent to a different email address. Please login with the correct account.")
	}

	alreadyMember, err := s.wsRepo.IsMember(ctx, invite.WorkspaceID, loggedInUserID)
	if err == nil && alreadyMember {
		_ = s.wsRepo.UpdateInviteStatus(ctx, invite.WorkspaceID, invite.ID, "ACCEPTED")
		return apperror.BadRequest("You are already a member of this workspace")
	}

	memberID, err := uuid.NewV7()
	if err != nil {
		return apperror.InternalServer("failed to generate member ID: " + err.Error())
	}

	member := &domain.WorkspaceMember{
		ID:          memberID.String(),
		WorkspaceID: invite.WorkspaceID,
		UserID:      loggedInUserID,
		Role:        invite.Role,
	}

	if err := s.wsRepo.AddMember(ctx, invite.WorkspaceID, member); err != nil {
		log.Printf("Failed to add member to workspace: %v", err)
		return apperror.InternalServer("Could not accept invitation")
	}

	if err := s.wsRepo.UpdateInviteStatus(ctx, invite.WorkspaceID, invite.ID, "ACCEPTED"); err != nil {
		log.Printf("Failed to update invite status: %v", err)
	}

	// Invalidate/Add to cache
	if cacheErr := s.cache.AddJoinedWorkspaceID(ctx, loggedInUserID, invite.WorkspaceID); cacheErr != nil {
		log.Printf("[Cache Error] AcceptInvitation AddJoinedWorkspaceID failed: %v", cacheErr)
	}

	memberRes := domain.WorkspaceMemberResponse{
		UserID:   loggedInUserID,
		FullName: currentUser.FullName,
		Email:    currentUser.Email,
		Role:     invite.Role,
		JoinedAt: time.Now(),
	}
	if cacheErr := s.cache.AddMember(ctx, invite.WorkspaceID, memberRes); cacheErr != nil {
		log.Printf("[Cache Error] AcceptInvitation AddMember failed: %v", cacheErr)
	}
	if cacheErr := s.cache.SetMemberRole(ctx, invite.WorkspaceID, loggedInUserID, invite.Role); cacheErr != nil {
		log.Printf("[Cache Error] AcceptInvitation SetMemberRole failed: %v", cacheErr)
	}

	ws, wsErr := s.wsRepo.FindByID(ctx, invite.WorkspaceID, invite.WorkspaceID)
	if wsErr == nil && ws != nil {
		if ws.OwnerID != loggedInUserID {
			ownerNotification := domain.NotificationEventPayload{
				Channel: "IN_APP",
				UserID:  ws.OwnerID,
				Title:   "Workspace Member Joined",
				Message: fmt.Sprintf("%s has accepted the invitation and joined the workspace %s", currentUser.FullName, ws.Name),
				Type:    "SUCCESS",
			}
			s.sendNotification(ctx, ownerNotification, ws.OwnerID)
		}

		userNotification := domain.NotificationEventPayload{
			Channel: "IN_APP",
			UserID:  loggedInUserID,
			Title:   "Workspace Joined",
			Message: fmt.Sprintf("You have successfully joined the workspace %s", ws.Name),
			Type:    "SUCCESS",
		}
		s.sendNotification(ctx, userNotification, loggedInUserID)
	}

	return nil
}

func (s *service) GetWorkspacesByMember(ctx context.Context, userID string, limit int, cursor string) ([]domain.Workspace, error) {
	ids, exists, err := s.cache.GetJoinedWorkspaceIDs(ctx, userID, limit, cursor)
	if err != nil {
		log.Printf("[Cache Error] GetJoinedWorkspaceIDs failed: %v", err)
	}

	var workspaces []domain.Workspace

	if !exists {
		// Cache Miss: Query all workspaces from DB to populate the cache
		allDBWorkspaces, err := s.wsRepo.GetByMemberID(ctx, "", userID, 10000, "")
		if err != nil {
			return nil, err
		}

		var dbIDs []string
		for _, ws := range allDBWorkspaces {
			dbIDs = append(dbIDs, ws.ID)
			if cacheErr := s.cache.SetWorkspaceMeta(ctx, &ws); cacheErr != nil {
				log.Printf("[Cache Error] SetWorkspaceMeta failed: %v", cacheErr)
			}
		}

		if cacheErr := s.cache.SetJoinedWorkspaceIDs(ctx, userID, dbIDs); cacheErr != nil {
			log.Printf("[Cache Error] SetJoinedWorkspaceIDs failed: %v", cacheErr)
		}

		workspaces = allDBWorkspaces
		sortWorkspacesDesc(workspaces)
		return paginateWorkspaces(workspaces, cursor, limit), nil
	}

	if len(ids) == 0 {
		return []domain.Workspace{}, nil
	}

	// Cache Hit: fetch metadata for only the returned paginated IDs!
	cachedMetas, missingIDs, err := s.cache.GetWorkspaceMetas(ctx, ids)
	if err != nil {
		log.Printf("[Cache Error] GetWorkspaceMetas failed: %v", err)
	}

	workspaces = cachedMetas

	if len(missingIDs) > 0 {
		for _, id := range missingIDs {
			ws, err := s.wsRepo.FindByID(ctx, id, id)
			if err != nil {
				return nil, err
			}
			if ws != nil {
				workspaces = append(workspaces, *ws)
				if cacheErr := s.cache.SetWorkspaceMeta(ctx, ws); cacheErr != nil {
					log.Printf("[Cache Error] SetWorkspaceMeta failed: %v", cacheErr)
				}
			}
		}
	}

	sortWorkspacesDesc(workspaces)
	return workspaces, nil
}

func (s *service) GetWorkspaceMembers(ctx context.Context, workspaceID string, userID string) ([]domain.WorkspaceMemberResponse, error) {
	ws, err := s.wsRepo.FindByID(ctx, workspaceID, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	var role string
	var cached bool
	if ws.OwnerID == userID {
		role = "OWNER"
	} else {
		role, cached, err = s.cache.GetMemberRole(ctx, workspaceID, userID)
		if err != nil {
			log.Printf("[Cache Error] GetMemberRole failed: %v", err)
		}
		if !cached {
			role, err = s.wsRepo.GetMemberRole(ctx, workspaceID, userID)
			if err != nil {
				// User is not a member or DB error
			} else {
				if cacheErr := s.cache.SetMemberRole(ctx, workspaceID, userID, role); cacheErr != nil {
					log.Printf("[Cache Error] SetMemberRole failed: %v", cacheErr)
				}
			}
		}
	}

	if role == "" {
		return nil, apperror.Forbidden("You do not have permission to view this workspace's members")
	}

	members, exists, err := s.cache.GetMembers(ctx, workspaceID)
	if err != nil {
		log.Printf("[Cache Error] GetMembers failed: %v", err)
	}

	if !exists {
		members, err = s.wsRepo.GetMembers(ctx, workspaceID)
		if err != nil {
			return nil, err
		}

		if cacheErr := s.cache.SetMembers(ctx, workspaceID, members); cacheErr != nil {
			log.Printf("[Cache Error] SetMembers failed: %v", cacheErr)
		}
	} else {
		sort.Slice(members, func(i, j int) bool {
			return members[i].JoinedAt.Before(members[j].JoinedAt)
		})
	}

	return members, nil
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
