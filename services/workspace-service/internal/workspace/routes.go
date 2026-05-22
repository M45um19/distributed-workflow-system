package workspace

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, ctrl *Controller, authMid *middleware.AuthMiddleware) {
	workspaceGroup := r.Group("/workspaces")
	{
		workspaceGroup.POST("/", authMid.Protect(), ctrl.CreateWorkspace)
		workspaceGroup.GET("/", authMid.Protect(), ctrl.ListWorkspaces)
		workspaceGroup.POST("/:id/invite", authMid.Protect(), ctrl.InviteUserHandler)
		workspaceGroup.POST("/invitations/accept", authMid.Protect(), ctrl.AcceptInvite)
	}
}
