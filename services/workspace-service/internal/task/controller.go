package task

import (
	"net/http"
	"strconv"
	"strings"

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
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	projectID := c.Param("projectId")
	var input domain.TaskCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.CreateTask(c.Request.Context(), workspaceID, projectID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Task created successfully!", result)
}

func (ctrl *Controller) ListTasks(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	projectID := c.Param("projectId")
	userID := c.GetString("user_id")

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 20 {
				limit = 20
			} else {
				limit = l
			}
		}
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	rawStatuses := c.QueryArray("status")
	var statuses []string
	for _, s := range rawStatuses {
		if strings.Contains(s, ",") {
			parts := strings.Split(s, ",")
			for _, p := range parts {
				statuses = append(statuses, strings.ToUpper(strings.TrimSpace(p)))
			}
		} else {
			statuses = append(statuses, strings.ToUpper(strings.TrimSpace(s)))
		}
	}

	tasks, err := ctrl.service.GetTasksByProject(c.Request.Context(), workspaceID, projectID, userID, statuses, limit, page)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Tasks fetched successfully", tasks)
}

func (ctrl *Controller) UpdateTask(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	taskID := c.Param("taskId")
	var input domain.TaskUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.UpdateFullTask(c.Request.Context(), workspaceID, taskID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Task updated successfully", result)
}

func (ctrl *Controller) UpdateStatus(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	taskID := c.Param("taskId")
	var input domain.TaskStatusUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	err := ctrl.service.UpdateTaskStatus(c.Request.Context(), workspaceID, taskID, input.Status, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Task status updated successfully", nil)
}

func (ctrl *Controller) AddComment(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	taskID := c.Param("taskId")
	var input domain.CommentCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.AddComment(c.Request.Context(), workspaceID, taskID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Comment added successfully", result)
}

func (ctrl *Controller) GetComments(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	taskID := c.Param("taskId")
	userID := c.GetString("user_id")

	comments, err := ctrl.service.GetTaskComments(c.Request.Context(), workspaceID, taskID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Comments fetched successfully", comments)
}
