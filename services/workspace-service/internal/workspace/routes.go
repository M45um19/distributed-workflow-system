package workspace

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, ctrl *Controller, authMid *middleware.AuthMiddleware) {
	ws := r.Group("/", authMid.Protect())
	{
		ws.POST("", ctrl.CreateWorkspace)
		ws.GET("/owned", ctrl.ListWorkspacesByOwner)
		ws.GET("/joined", ctrl.ListWorkspacesByMember)

		ws.GET("/:id/members", ctrl.GetWorkspaceMembersHandler)
		ws.POST("/:id/invites", ctrl.InviteUser)
		ws.POST("/invites/accept", ctrl.AcceptInvite)
	}
}
