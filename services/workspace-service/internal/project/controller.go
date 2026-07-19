package project

import (
	"encoding/base64"
	"net/http"
	"strconv"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Controller struct {
	service domain.ProjectService
}

func NewController(s domain.ProjectService) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) CreateProject(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	var input domain.ProjectCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.CreateProject(c.Request.Context(), workspaceID, input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Project created successfully!", result)
}

func (ctrl *Controller) ListProjects(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

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

	userID := c.GetString("user_id")
	projects, nextCursor, err := ctrl.service.GetProjectsByWorkspace(c.Request.Context(), workspaceID, userID, limit, cursor)
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

	response.SendResponse(c, http.StatusOK, true, "Projects fetched successfully", projects, meta)
}
