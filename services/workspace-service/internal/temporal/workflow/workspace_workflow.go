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
	Token string `json:"token"`
}

func InviteUserWorkflow(ctx workflow.Context, data WorkspaceInviteRequest) error {
	options := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute * 5,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second * 2,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute * 1,
			MaximumAttempts:    5,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var a *activity.WorkspaceActivities

	err := workflow.ExecuteActivity(ctx, a.SendInviteEmail, data.Email, data.Token).Get(ctx, nil)
	if err != nil {
		return err
	}

	_ = workflow.Sleep(ctx, time.Hour*24*10)

	var status string
	err = workflow.ExecuteActivity(ctx, a.CheckInviteStatus, data.Token).Get(ctx, &status)
	if err != nil {
		return err
	}

	if status == "PENDING" {
		err = workflow.ExecuteActivity(ctx, a.SendReminderEmail, data.Email).Get(ctx, nil)
		if err != nil {
			return err
		}

		_ = workflow.Sleep(ctx, time.Hour*24*4)

	}

	return nil
}
