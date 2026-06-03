package task

import (
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service domain.TaskService
}

func NewController(s domain.TaskService) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) CreateTask(c *gin.Context) {
	projectID := c.Param("projectId")
	var input domain.TaskCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.CreateTask(c.Request.Context(), projectID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Task created successfully!", result)
}

func (ctrl *Controller) ListTasks(c *gin.Context) {
	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	tasks, err := ctrl.service.GetTasksByProject(c.Request.Context(), projectID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Tasks fetched successfully", tasks)
}

func (ctrl *Controller) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	var input domain.TaskUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.UpdateFullTask(c.Request.Context(), taskID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Task updated successfully", result)
}

func (ctrl *Controller) UpdateStatus(c *gin.Context) {
	taskID := c.Param("id")
	var input domain.TaskStatusUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	err := ctrl.service.UpdateTaskStatus(c.Request.Context(), taskID, input.Status, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Task status updated successfully", nil)
}

func (ctrl *Controller) AddComment(c *gin.Context) {
	taskID := c.Param("id")
	var input domain.CommentCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.AddComment(c.Request.Context(), taskID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Comment added successfully", result)
}

func (ctrl *Controller) GetComments(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.GetString("user_id")

	comments, err := ctrl.service.GetTaskComments(c.Request.Context(), taskID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Comments fetched successfully", comments)
}
