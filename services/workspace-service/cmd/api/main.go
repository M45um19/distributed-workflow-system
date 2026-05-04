package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/app"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/workspace"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	conn, err := grpc.Dial(cfg.AuthServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("gRPC dial failed: %v", err)
	}
	defer conn.Close()
	grpcClient := pb.NewAuthServiceClient(conn)

	container := app.NewContainer(cfg, db, rdb, grpcClient)

	r := gin.Default()
	api := r.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "UP", "redis": rdb != nil})
		})

		workspace.RegisterRoutes(api, container.WorkspaceCtrl, container.AuthMid)
	}

	fmt.Printf("Workspace Service running on port %s\n", cfg.Port)
	r.Run(":" + cfg.Port)
}
