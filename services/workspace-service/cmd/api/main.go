package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/config"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/app"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/project"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/task"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/workspace"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/monitoring"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"

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

	rdb := config.ConnectRedis(cfg.RedisURI)

	authConn := config.NewAuthgRPCConnection(cfg.AuthServiceGRPCAddress)

	authGRPCClient := pb.NewAuthServiceClient(authConn)

	container := app.NewContainer(cfg, db, rdb, authGRPCClient, false)
	r := gin.Default()
	r.Use(middleware.GlobalErrorHandler(cfg.GoENV))
	r.Use(monitoring.MetricsMiddleware())
	api := r.Group("/api/v1/workspace")
	{

		setupHealthRoutes(api, db, rdb)
		api.GET("/metrics", monitoring.MetricsHandler())
		workspace.RegisterRoutes(api, container.WorkspaceCtrl, container.AuthMid)
		project.RegisterRoutes(api, container.ProjectCtrl, container.AuthMid)
		task.RegisterRoutes(api, container.TaskCtrl, container.AuthMid)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		fmt.Printf("Workspace API Server running on port %s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Listen failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	log.Printf("Received %v. Starting graceful shutdown for Workspace API Server...\n", sig)

	time.Sleep(5 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("HTTP Server stopped. Cleaning up background connections...")

	if container.TemporalClient != nil {
		container.TemporalClient.Close()
		log.Println("Temporal Client closed.")
	}
	if container.KafkaWriter != nil {
		if err := container.KafkaWriter.Close(); err != nil {
			log.Printf("Error closing Kafka Writer: %v", err)
		} else {
			log.Println("Kafka Writer flushed and closed safely.")
		}
	}
	if rdb != nil {
		rdb.Close()
		log.Println("Redis connection closed.")
	}

	authConn.Close()
	log.Println("Auth gRPC connection closed.")

	db.Close()
	log.Println("PostgreSQL connection closed safely.")

	log.Println("Workspace API Server shutdown completed successfully.")
}

func setupHealthRoutes(api *gin.RouterGroup, db *sqlx.DB, rdb *redis.Client) {
	api.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ALIVE"})
	})

	api.GET("/health", func(c *gin.Context) {
		dbConnected := db.DB.Ping() == nil

		redisConnected := false
		if rdb != nil {
			_, err := rdb.Ping(context.Background()).Result()
			redisConnected = err == nil
		}

		if dbConnected && redisConnected {
			c.JSON(http.StatusOK, gin.H{
				"status":   "UP",
				"service":  "workspace-service",
				"database": "connected",
				"redis":    "connected",
			})
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "DOWN",
				"service":  "workspace-service",
				"database": map[bool]string{true: "connected", false: "disconnected"}[dbConnected],
				"redis":    map[bool]string{true: "connected", false: "disconnected"}[redisConnected],
			})
		}
	})
}
