package workspace

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
}

func NewService(wsRepo domain.WorkspaceRepository, userRepo domain.UserRepository, tempClient client.Client, notificationkafkaWriter *kafka.Writer) domain.WorkspaceService {
	return &service{
		wsRepo:                  wsRepo,
		userRepo:                userRepo,
		temporalClient:          tempClient,
		notificationkafkaWriter: notificationkafkaWriter,
	}
}

func (s *service) CreateWorkspace(ctx context.Context, input domain.WorkspaceCreateInput, ownerID string) (*domain.Workspace, error) {
	exists, _ := s.wsRepo.FindBySlug(ctx, input.Slug)
	if exists != nil {
		return nil, apperror.BadRequest("workspace with this slug already exists")
	}

	ws := &domain.Workspace{
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		OwnerID:     ownerID,
	}

	if err := s.wsRepo.Create(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *service) GetWorkspacesByOwner(ctx context.Context, ownerId string, limit, page int) ([]domain.Workspace, error) {
	offset := (page - 1) * limit
	return s.wsRepo.GetByOwnerID(ctx, ownerId, limit, offset)
}

func (s *service) InviteUser(ctx context.Context, input domain.WorkspaceInviteRequest) error {

	input.Token = uuid.New().String()

	invite := &domain.WorkspaceInvitation{
		WorkspaceID: input.WorkspaceID,
		InviterID:   input.InviterID,
		Email:       input.Email,
		Role:        input.Role,
		Token:       input.Token,
		Status:      "PENDING",
		ExpiresAt:   time.Now().Add(time.Hour * 24 * 14),
	}

	if err := s.wsRepo.CreateInvite(ctx, invite); err != nil {
		log.Printf("Failed to save invitation to DB: %v", err)
		return apperror.InternalServer("Could not process invitation")
	}

	notificationPayload := domain.NotificationEventPayload{
		Channel: "IN_APP",
		UserID:  input.InviterID,
		Title:   "Workspace Invitation",
		Message: fmt.Sprintf("You have been invited to join the workspace as a %s", input.Role),
		Type:    "INFO",
	}

	jsonData, err := json.Marshal(notificationPayload)
	if err != nil {
		log.Printf("Failed to marshal kafka notification payload: %v", err)
	} else {
		err = s.notificationkafkaWriter.WriteMessages(ctx, kafka.Message{
			Key:   []byte(input.Email),
			Value: jsonData,
		})
		if err != nil {
			log.Printf("Kafka failed to send message to send-notification: %v", err)
		} else {
			log.Printf("Successfully produced message to send-notification for: %s", input.Email)
		}
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
		return apperror.InternalServer("Temporal error: " + err.Error())
	}

	log.Printf("Started workflow. WorkflowID: %s, RunID: %s", we.GetID(), we.GetRunID())
	return nil
}

func (s *service) AcceptInvitation(ctx context.Context, token string, loggedInUserID string) error {
	invite, err := s.wsRepo.FindInviteByToken(ctx, token)
	if err != nil {
		return apperror.NotFound("Invitation link is invalid or does not exist")
	}

	if invite.Status != "PENDING" {
		return apperror.BadRequest("This invitation has already been " + invite.Status)
	}

	if time.Now().After(invite.ExpiresAt) {
		_ = s.wsRepo.UpdateInviteStatus(ctx, invite.ID, "EXPIRED")
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
		_ = s.wsRepo.UpdateInviteStatus(ctx, invite.ID, "ACCEPTED")
		return apperror.BadRequest("You are already a member of this workspace")
	}

	member := &domain.WorkspaceMember{
		WorkspaceID: invite.WorkspaceID,
		UserID:      loggedInUserID,
		Role:        invite.Role,
	}

	if err := s.wsRepo.AddMember(ctx, member); err != nil {
		log.Printf("Failed to add member to workspace: %v", err)
		return apperror.InternalServer("Could not accept invitation")
	}

	if err := s.wsRepo.UpdateInviteStatus(ctx, invite.ID, "ACCEPTED"); err != nil {
		log.Printf("Failed to update invite status: %v", err)
	}

	return nil
}

func (s *service) GetWorkspacesByMember(ctx context.Context, userID string, limit, page int) ([]domain.Workspace, error) {
	offset := (page - 1) * limit
	return s.wsRepo.GetByMemberID(ctx, userID, limit, offset)
}

func (s *service) GetWorkspaceMembers(ctx context.Context, workspaceID string, userID string) ([]domain.WorkspaceMemberResponse, error) {
	ws, err := s.wsRepo.FindByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	if ws.OwnerID == userID {
		return s.wsRepo.GetMembers(ctx, workspaceID)
	}

	isMember, err := s.wsRepo.IsMember(ctx, workspaceID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, apperror.Forbidden("You do not have permission to view this workspace's members")
	}

	return s.wsRepo.GetMembers(ctx, workspaceID)
}
