package app

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/user"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/workspace"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	WorkspaceCtrl *workspace.Controller
	AuthMid       *middleware.AuthMiddleware
	KafkaWorker   *KafkaWorker
}

func NewContainer(cfg *config.Config, db *sqlx.DB, rdb *redis.Client, grpc pb.AuthServiceClient) *Container {
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	userRegHandler := user.NewUserRegisteredHandler(userSvc)

	// Kafka Setup
	regReader := config.NewKafkaReader(cfg.KafkaBrokers, "user-registered", "workspace-registration-group")
	worker := NewKafkaWorker()
	worker.AddTopicHandler(regReader, userRegHandler)

	// Workspace Setup
	wsRepo := workspace.NewRepository(db)
	wsSvc := workspace.NewService(wsRepo)
	wsCtrl := workspace.NewController(wsSvc)

	authMid := middleware.NewAuthMiddleware(cfg.JWTSecret, rdb, grpc)

	return &Container{
		WorkspaceCtrl: wsCtrl,
		AuthMid:       authMid,
		KafkaWorker:   worker,
	}
}
