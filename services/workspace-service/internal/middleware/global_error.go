package middleware

import (
	"encoding/json"
	"log"
	"runtime/debug"
	"time"

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

type lokiLogPayload struct {
	Level       string `json:"level"`
	Timestamp   string `json:"timestamp"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Message     string `json:"message"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	StatusCode  int    `json:"status_code"`
	Stack       string `json:"stack,omitempty"`
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

			if appErr, ok := err.(*apperror.AppError); ok {
				statusCode = appErr.StatusCode
				message = appErr.Message
				stackTrace = appErr.Stack
				isOperational = true
			}

			if stackTrace == "" {
				stackTrace = string(debug.Stack())
			}

			logLevel := "error"
			if isOperational && statusCode < 500 {
				logLevel = "warn"
			}

			logPayload := lokiLogPayload{
				Level:       logLevel,
				Timestamp:   time.Now().UTC().Format(time.RFC3339),
				Service:     "workspace-service",
				Environment: env,
				Message:     err.Error(),
				Method:      c.Request.Method,
				Path:        c.Request.URL.Path,
				StatusCode:  statusCode,
			}

			if logLevel == "error" {
				logPayload.Stack = stackTrace
			}

			if jsonLog, err := json.Marshal(logPayload); err == nil {
				log.Println(string(jsonLog))
			} else {
				log.Printf("[%s] ERROR: %v", logLevel, err)
			}

			if env == "development" {
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
					statusCode = 500
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
