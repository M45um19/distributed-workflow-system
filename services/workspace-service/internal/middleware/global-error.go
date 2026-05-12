// internal/middleware/error_handler.go
package middleware

import (
	"log"
	"runtime/debug"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/gin-gonic/gin"
)

type errorDetails struct {
	StatusCode    int  `json:"statusCode"`
	IsOperational bool `json:"isOperational"`
}

type errorResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Stack   string       `json:"stack,omitempty"`
	Error   errorDetails `json:"error"`
}

func GlobalErrorHandler(env string) gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Next()

		log.Printf("[DEBUG] Current Env: %s, Total Errors: %d", env, len(c.Errors))

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			statusCode := 500
			message := "Something went very wrong!"

			var stackTrace string

			if appErr, ok := err.(*apperror.AppError); ok {
				statusCode = appErr.StatusCode
				message = appErr.Message
				stackTrace = appErr.Stack
			}

			isOperational := false
			if _, ok := err.(*apperror.AppError); ok {
				isOperational = true
			}

			res := errorResponse{
				Success: false,
				Message: message,
				Error: errorDetails{
					StatusCode:    statusCode,
					IsOperational: isOperational,
				},
			}

			if env == "development" {

				if stackTrace != "" {
					res.Stack = stackTrace
				} else {
					res.Stack = string(debug.Stack())
				}
			}

			c.JSON(statusCode, res)
			c.Abort()
		}
	}
}
