package activity

import (
	"context"
	"log"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/email"
)

type WorkspaceActivities struct {
	repo        domain.WorkspaceRepository
	emailClient email.EmailClient
}

func NewWorkspaceActivities(repo domain.WorkspaceRepository, mail email.EmailClient) *WorkspaceActivities {
	return &WorkspaceActivities{
		repo:        repo,
		emailClient: mail,
	}
}

func (a *WorkspaceActivities) SendInviteEmail(ctx context.Context, to string, token string) error {
	log.Printf("[Activity] Sending initial invite to %s", to)
	return a.emailClient.SendInvite(ctx, to, token)
}

func (a *WorkspaceActivities) SendReminderEmail(ctx context.Context, to string) error {
	log.Printf("[Activity] Sending 10-day reminder to %s", to)
	return a.emailClient.SendReminder(ctx, to)
}

func (a *WorkspaceActivities) CheckInviteStatus(ctx context.Context, token string) (string, error) {
	invite, err := a.repo.FindInviteByToken(ctx, token)
	if err != nil {
		return "", err
	}
	return invite.Status, nil
}

func (a *WorkspaceActivities) ExpireInvite(ctx context.Context, token string) error {
	log.Printf("[Activity] Expiring invite token: %s", token)

	invite, err := a.repo.FindInviteByToken(ctx, token)
	if err != nil {
		return err
	}

	if invite.Status != "PENDING" {
		return nil
	}

	return a.repo.UpdateInviteStatus(ctx, invite.ID, "EXPIRED")
}
