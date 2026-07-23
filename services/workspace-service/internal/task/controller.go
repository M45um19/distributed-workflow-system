package task

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	cursorBase64 := c.Query("cursor")
	var cursor string
	if cursorBase64 != "" {
		var decoded []byte
		var err error
		decoded, err = base64.URLEncoding.DecodeString(cursorBase64)
		if err != nil {
			decoded, err = base64.StdEncoding.DecodeString(cursorBase64)
		}
		if err != nil {
			c.Error(apperror.BadRequest("Invalid cursor format: must be valid base64"))
			return
		}
		cursor = string(decoded)
		if _, err := uuid.Parse(cursor); err != nil {
			c.Error(apperror.BadRequest("Invalid cursor format: decoded cursor is not a valid UUID"))
			return
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

	tasks, nextCursors, err := ctrl.service.GetTasksByProject(c.Request.Context(), workspaceID, projectID, userID, statuses, limit, cursor)
	if err != nil {
		c.Error(err)
		return
	}

	nextCursorsBase64 := make(map[string]string)
	for status, nc := range nextCursors {
		if nc != "" {
			nextCursorsBase64[status] = base64.URLEncoding.EncodeToString([]byte(nc))
		} else {
			nextCursorsBase64[status] = ""
		}
	}

	meta := gin.H{
		"next_cursors": nextCursorsBase64,
	}

	response.SendResponse(c, http.StatusOK, true, "Tasks fetched successfully", tasks, meta)
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

	cursorBase64 := c.Query("cursor")
	var cursor string
	if cursorBase64 != "" {
		var decoded []byte
		var err error
		decoded, err = base64.URLEncoding.DecodeString(cursorBase64)
		if err != nil {
			decoded, err = base64.StdEncoding.DecodeString(cursorBase64)
		}
		if err != nil {
			c.Error(apperror.BadRequest("Invalid cursor format: must be valid base64"))
			return
		}
		cursor = string(decoded)
		if _, err := uuid.Parse(cursor); err != nil {
			c.Error(apperror.BadRequest("Invalid cursor format: decoded cursor is not a valid UUID"))
			return
		}
	}

	comments, nextCursor, err := ctrl.service.GetTaskComments(c.Request.Context(), workspaceID, taskID, userID, limit, cursor)
	if err != nil {
		c.Error(err)
		return
	}

	var nextCursorBase64 string
	if nextCursor != "" {
		nextCursorBase64 = base64.URLEncoding.EncodeToString([]byte(nextCursor))
	}

	meta := gin.H{
		"next_cursor": nextCursorBase64,
	}

	response.SendResponse(c, http.StatusOK, true, "Comments fetched successfully", comments, meta)
}
