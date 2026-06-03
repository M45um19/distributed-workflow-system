package app

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/kafka"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/project"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/task"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/temporal"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/temporal/activity"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/temporal/workflow"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/user"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/workspace"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/email"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	temporalWorker "go.temporal.io/sdk/worker"
)

type Container struct {
	WorkspaceCtrl  *workspace.Controller
	ProjectCtrl    *project.Controller
	TaskCtrl       *task.Controller
	AuthMid        *middleware.AuthMiddleware
	KafkaWorker    *kafka.Worker
	TemporalWorker *temporal.Worker
	TemporalClient client.Client
}

func NewContainer(cfg *config.Config, db *sqlx.DB, rdb *redis.Client, authGRPCClient pb.AuthServiceClient, isWorker bool) *Container {
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)

	wsRepo := workspace.NewRepository(db)
	projectRepo := project.NewRepository(db)
	taskRepo := task.NewRepository(db)

	tempClient := config.ConnectTemporal(cfg.TemporalHost)
	wsSvc := workspace.NewService(wsRepo, userRepo, tempClient)
	wsCtrl := workspace.NewController(wsSvc)

	projectSvc := project.NewService(projectRepo, wsRepo)
	projectCtrl := project.NewController(projectSvc)

	taskSvc := task.NewService(taskRepo, wsRepo)
	taskCtrl := task.NewController(taskSvc)

	authMid := middleware.NewAuthMiddleware(cfg.JWTSecret, rdb, authGRPCClient)

	container := &Container{
		WorkspaceCtrl:  wsCtrl,
		ProjectCtrl:    projectCtrl,
		TaskCtrl:       taskCtrl,
		AuthMid:        authMid,
		TemporalClient: tempClient,
	}

	if isWorker {
		emailClient := email.NewGmailClient(cfg.SmtpHost, cfg.SmtpPort, cfg.SmtpFrom, cfg.SmtpPassword)
		userRegHandler := kafka.NewUserRegisteredHandler(userSvc)
		regReader := config.NewKafkaReader(cfg.KafkaBrokers, "user-registered", "workspace-registration-group")
		kWorker := kafka.NewWorker()
		kWorker.AddTopicHandler(regReader, userRegHandler)
		container.KafkaWorker = kWorker

		tempWorker := temporal.NewWorker(tempClient, "workspace-task-queue")
		tempWorker.Register(func(w temporalWorker.Worker) {
			wsActivities := activity.NewWorkspaceActivities(wsRepo, emailClient)
			w.RegisterWorkflow(workflow.InviteUserWorkflow)
			w.RegisterActivity(wsActivities)
		})
		container.TemporalWorker = tempWorker
	}

	return container
}
