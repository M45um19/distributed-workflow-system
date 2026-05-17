package workflow

import (
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/temporal/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type WorkspaceInviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	Token string `json:"-"`
}

func InviteUserWorkflow(ctx workflow.Context, data WorkspaceInviteRequest) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 1,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 1,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var a activity.WorkspaceActivities
	err := workflow.ExecuteActivity(ctx, a.SendInviteEmail, data.Email, data.Token).Get(ctx, nil)
	return err
}
