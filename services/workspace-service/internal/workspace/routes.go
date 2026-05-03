package workspace

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	pb "github.com/M45um19/distributed-workflow-system/services/workspace-service/pb/auth"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(r *gin.RouterGroup, ctrl *Controller, jwtSecret string, rdb *redis.Client, grpcClient pb.AuthServiceClient) {
	workspaceGroup := r.Group("/workspaces")
	{
		workspaceGroup.POST("/create", middleware.Protect(jwtSecret, rdb, grpcClient), ctrl.CreateWorkspace)
	}
}
