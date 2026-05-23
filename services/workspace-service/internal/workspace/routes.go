package workspace

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, ctrl *Controller, authMid *middleware.AuthMiddleware) {
	workspaceGroup := r.Group("/workspaces")
	{
		workspaceGroup.POST("/", authMid.Protect(), ctrl.CreateWorkspace)
		workspaceGroup.GET("/", authMid.Protect(), ctrl.ListWorkspacesByOwner)

		workspaceGroup.GET("/member", authMid.Protect(), ctrl.ListWorkspacesByMember)
		workspaceGroup.POST("/:id/invite", authMid.Protect(), ctrl.InviteUser)
		workspaceGroup.POST("/invitations/accept", authMid.Protect(), ctrl.AcceptInvite)
	}
}
