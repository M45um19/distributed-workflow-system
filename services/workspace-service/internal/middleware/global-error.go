// internal/middleware/error_handler.go
package middleware

import (
	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/gin-gonic/gin"
)

// internal/middleware/error_handler.go

type errorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   any    `json:"error,omitempty"`
	Stack   string `json:"stack,omitempty"`
}

func GlobalErrorHandler(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			statusCode := 500
			message := "Something went very wrong!"

			if appErr, ok := err.(*apperror.AppError); ok {
				statusCode = appErr.StatusCode
				message = appErr.Message
			}

			res := errorResponse{
				Success: false,
				Message: message,
			}

			if env == "development" {
				res.Error = err
				res.Stack = "stack trace logic here"
			}

			c.JSON(statusCode, res)
			c.Abort()
		}
	}
}
