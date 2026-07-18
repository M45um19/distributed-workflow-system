package workspace

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
	service domain.WorkspaceService
}

func NewController(s domain.WorkspaceService) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) CreateWorkspace(c *gin.Context) {
	var input domain.WorkspaceCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.CreateWorkspace(c.Request.Context(), input, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Workspace created successfully!", result)
}

func (ctrl *Controller) ListWorkspacesByOwner(c *gin.Context) {
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

	workspaces, err := ctrl.service.GetWorkspacesByOwner(c.Request.Context(), c.GetString("user_id"), limit, cursor)
	if err != nil {
		c.Error(err)
		return
	}

	var nextCursor string
	if len(workspaces) > 0 && len(workspaces) == limit {
		lastWS := workspaces[len(workspaces)-1]
		nextCursor = base64.URLEncoding.EncodeToString([]byte(lastWS.ID))
	}

	meta := gin.H{
		"next_cursor": nextCursor,
	}

	response.SendResponse(c, http.StatusOK, true, "All workspace fetch successful", workspaces, meta)
}

func (ctrl *Controller) InviteUser(c *gin.Context) {
	var input domain.WorkspaceInviteRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	input.InviterID = c.GetString("user_id")

	input.WorkspaceID = c.Param("id")

	if input.WorkspaceID == "" {
		c.Error(apperror.BadRequest("workspace_id is required"))
		return
	}
	if input.Role == "" {
		input.Role = "MEMBER"
	}

	ctx := c.Request.Context()
	inviteRes, err := ctrl.service.InviteUser(ctx, input)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusAccepted, true, "Invitation sending in progress", inviteRes)
}

func (ctrl *Controller) AcceptInvite(c *gin.Context) {
	var req domain.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperror.BadRequest("Validation failed: " + err.Error()))
		return
	}

	userID := c.GetString("user_id")

	err := ctrl.service.AcceptInvitation(c.Request.Context(), req.Token, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Successfully joined the workspace!", nil)
}

func (ctrl *Controller) ListWorkspacesByMember(c *gin.Context) {
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

	workspaces, err := ctrl.service.GetWorkspacesByMember(c.Request.Context(), userID, limit, cursor)
	if err != nil {
		c.Error(err)
		return
	}

	var nextCursor string
	if len(workspaces) > 0 && len(workspaces) == limit {
		lastWS := workspaces[len(workspaces)-1]
		nextCursor = base64.URLEncoding.EncodeToString([]byte(lastWS.ID))
	}

	meta := gin.H{
		"next_cursor": nextCursor,
	}

	response.SendResponse(c, http.StatusOK, true, "Member workspaces fetched successfully", workspaces, meta)
}

func (ctrl *Controller) GetWorkspaceMembersHandler(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.Error(apperror.BadRequest("Workspace ID is required"))
		return
	}

	userID := c.GetString("user_id")

	members, err := ctrl.service.GetWorkspaceMembers(c.Request.Context(), workspaceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Workspace members fetched successfully", members)
}
