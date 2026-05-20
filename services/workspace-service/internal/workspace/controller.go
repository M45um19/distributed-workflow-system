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

func (ctrl *Controller) ListWorkspaces(c *gin.Context) {
	workspaces, err := ctrl.service.GetUserWorkspaces(c.Request.Context(), c.GetString("user_id"))
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "All workspace fetch successful", workspaces)
}

func (ctrl *Controller) InviteUserHandler(c *gin.Context) {
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
