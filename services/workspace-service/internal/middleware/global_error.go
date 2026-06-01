package middleware

import (
	"runtime/debug"

	"github.com/M45um19/distributed-workflow-system/services/workspace-service/pkg/apperror"
	"github.com/gin-gonic/gin"
)

type errorDetails struct {
	StatusCode    int  `json:"statusCode"`
	IsOperational bool `json:"isOperational"`
}

type devErrorResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Stack   string       `json:"stack"`
	Error   errorDetails `json:"error"`
}

type prodErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func GlobalErrorHandler(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			statusCode := 500
			message := "Something went very wrong!"
			var stackTrace string
			isOperational := false

			// Check if it's a custom AppError
			if appErr, ok := err.(*apperror.AppError); ok {
				statusCode = appErr.StatusCode
				message = appErr.Message
				stackTrace = appErr.Stack
				isOperational = true
			}

			if env == "development" {
				if stackTrace == "" {
					stackTrace = string(debug.Stack())
				}

				res := devErrorResponse{
					Success: false,
					Message: message,
					Stack:   stackTrace,
					Error: errorDetails{
						StatusCode:    statusCode,
						IsOperational: isOperational,
					},
				}
				c.JSON(statusCode, res)
			} else {
				if !isOperational {
					message = "Something went very wrong!"
				}

				res := prodErrorResponse{
					Success: false,
					Message: message,
				}
				c.JSON(statusCode, res)
			}
			c.Abort()
		}
	}
}
