package task

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.RouterGroup, ctrl *Controller, authMid *middleware.AuthMiddleware) {
	protected := r.Group("/", authMid.Protect())
	{

		projectTasks := protected.Group("/projects/:projectId/tasks")
		{
			projectTasks.POST("", ctrl.CreateTask)
			projectTasks.GET("", ctrl.ListTasks)
		}

		singleTask := protected.Group("/tasks/:id")
		{
			singleTask.PUT("", ctrl.UpdateTask)
			singleTask.PATCH("/status", ctrl.UpdateStatus)
			singleTask.POST("/comments", ctrl.AddComment)
			singleTask.GET("/comments", ctrl.GetComments)
		}
	}
}
