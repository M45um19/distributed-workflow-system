package workspace

import (
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
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
	workspaces, err := ctrl.service.GetWorkspacesByOwner(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "All workspace fetch successful", workspaces)
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
	err := ctrl.service.InviteUser(ctx, input)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusAccepted, true, "Invitation sending in progress", nil)
}

func (ctrl *Controller) AcceptInvite(c *gin.Context) {
	var req domain.AcceptInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")

	err := ctrl.service.AcceptInvitation(c.Request.Context(), req.Token, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully joined the workspace!"})
}

func (ctrl *Controller) ListWorkspacesByMember(c *gin.Context) {
	userID := c.GetString("user_id")

	workspaces, err := ctrl.service.GetWorkspacesByMember(c.Request.Context(), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Member workspaces fetched successfully", workspaces)
}
