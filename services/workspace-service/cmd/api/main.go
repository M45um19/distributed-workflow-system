package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/app"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/project"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/task"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/workspace"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config load failed: %v", err)
	}

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer db.Close()

	rdb := config.ConnectRedis(cfg.RedisURI)

	authConn := config.NewAuthgRPCConnection(cfg.AuthServiceGRPCAddress)
	defer authConn.Close()

	authGRPCClient := pb.NewAuthServiceClient(authConn)

	container := app.NewContainer(cfg, db, rdb, authGRPCClient, false)
	r := gin.Default()
	r.Use(middleware.GlobalErrorHandler(cfg.GoENV))

	api := r.Group("/api/v1/workspace")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "UP", "redis": rdb != nil})
		})

		workspace.RegisterRoutes(api, container.WorkspaceCtrl, container.AuthMid)
		project.RegisterRoutes(api, container.ProjectCtrl, container.AuthMid)
		task.RegisterRoutes(api, container.TaskCtrl, container.AuthMid)
	}

	fmt.Printf("Workspace API Server running on port %s\n", cfg.Port)
	r.Run(":" + cfg.Port)
}
