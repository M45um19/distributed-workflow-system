package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/workspace"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load config: %v", err)
	}

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Could not connect to DB: %v", err)
	}
	defer db.Close()

	rdb := config.ConnectRedis(cfg.RedisURI)
	if rdb == nil {
		log.Println("Warning: Redis is not connected")
	}

	conn, err := grpc.Dial(cfg.AuthServiceGRPCAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect to Auth gRPC service: %v", err)
	}
	defer conn.Close()
	grpcClient := pb.NewAuthServiceClient(conn)

	r := gin.Default()

	repo := workspace.NewRepository(db)
	svc := workspace.NewService(repo)
	ctrl := workspace.NewController(svc)

	api := r.Group("/api/v1")
	{
		api.GET("/workspaces/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "UP",
				"service": "Project Service",
				"redis":   rdb != nil,
			})
		})

		workspace.RegisterRoutes(api, ctrl, cfg.JWTSecret, rdb, grpcClient)
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	fmt.Printf("Project Service running on port %s\n", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
