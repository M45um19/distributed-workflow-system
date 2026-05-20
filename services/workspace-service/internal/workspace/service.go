package workspace

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/temporal/workflow"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"go.temporal.io/sdk/client"
)

type service struct {
	repo           domain.WorkspaceRepository
	temporalClient client.Client
}

func NewService(repo domain.WorkspaceRepository, tempClient client.Client) domain.WorkspaceService {
	return &service{
		repo:           repo,
		temporalClient: tempClient,
	}
}

func (s *service) CreateWorkspace(ctx context.Context, input domain.WorkspaceCreateInput, ownerID string) (*domain.Workspace, error) {
	exists, _ := s.repo.FindBySlug(ctx, input.Slug)
	if exists != nil {
		return nil, apperror.BadRequest("workspace with this slug already exists")
	}

	ws := &domain.Workspace{
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		OwnerID:     ownerID,
	}

	if err := s.repo.Create(ctx, ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (s *service) GetUserWorkspaces(ctx context.Context, ownerId string) ([]domain.Workspace, error) {
	return s.repo.GetByOwnerID(ctx, ownerId)
}

func (s *service) InviteUser(ctx context.Context, input domain.WorkspaceInviteRequest) error {
	b := make([]byte, 16)
	rand.Read(b)
	input.Token = hex.EncodeToString(b)

	invite := &domain.WorkspaceInvitation{
		WorkspaceID: input.WorkspaceID,
		InviterID:   input.InviterID,
		Email:       input.Email,
		Role:        input.Role,
		Token:       input.Token,
		Status:      "PENDING",
		ExpiresAt:   time.Now().Add(time.Hour * 24 * 14),
	}

	if err := s.repo.CreateInvite(ctx, invite); err != nil {
		log.Printf("Failed to save invitation to DB: %v", err)
		return apperror.InternalServer("Could not process invitation")
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
