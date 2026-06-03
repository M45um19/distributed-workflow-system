package project

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, ctrl *Controller, authMid *middleware.AuthMiddleware) {
	projectGroup := r.Group("/:id/projects", authMid.Protect())
	{
		projectGroup.POST("", ctrl.CreateProject)
		projectGroup.GET("", ctrl.ListProjects)
	}
}
