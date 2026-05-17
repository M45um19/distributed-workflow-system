package activity

import (
	"context"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
)

type WorkspaceActivities struct {
	repo domain.WorkspaceRepository
}

func NewWorkspaceActivities(repo domain.WorkspaceRepository) *WorkspaceActivities {
	return &WorkspaceActivities{repo: repo}
}

func (a *WorkspaceActivities) SendInviteEmail(ctx context.Context, email string, token string) error {
	log.Printf("[Temporal Activity] Email successfully sent to %s with token %s", email, token)
	return nil
}
