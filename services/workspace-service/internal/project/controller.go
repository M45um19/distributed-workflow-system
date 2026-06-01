package project

import (
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/internal/domain"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
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

	userID := c.GetString("user_id")
	projects, err := ctrl.service.GetProjectsByWorkspace(c.Request.Context(), workspaceID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SendResponse(c, http.StatusOK, true, "Projects fetched successfully", projects)
}
