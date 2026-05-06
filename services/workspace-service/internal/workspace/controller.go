package workspace

import (
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(s Service) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) CreateWorkspace(c *gin.Context) {
	var input CreateWorkspaceInput
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
