package workspace

import (
	"net/http"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/response"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service // Using Interface
}

func NewController(s Service) *Controller {
	return &Controller{service: s}
}

func (ctrl *Controller) CreateWorkspace(c *gin.Context) {
	var input CreateWorkspaceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.SendResponse(c, http.StatusBadRequest, false, "Validation failed", err.Error())
		return
	}

	userID := c.GetString("user_id")
	result, err := ctrl.service.CreateWorkspace(c.Request.Context(), input, userID)
	if err != nil {
		response.SendResponse(c, http.StatusBadRequest, false, err.Error(), nil)
		return
	}

	response.SendResponse(c, http.StatusCreated, true, "Workspace created successfully!", result)
}
