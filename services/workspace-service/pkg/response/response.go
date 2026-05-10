package response

import (
	"github.com/gin-gonic/gin"
)

type IApiResponse struct {
	StatusCode int16       `json:"statusCode`
	Success    bool        `json:"success"`
	Message    string      `json:"message"`
	Meta       interface{} `json:"meta,omitempty"`
	Data       interface{} `json:"data"`
}

func SendResponse(c *gin.Context, statusCode int, success bool, message string, data interface{}, meta ...interface{}) {
	var m interface{}
	if len(meta) > 0 {
		m = meta[0]
	}

	c.JSON(statusCode, IApiResponse{
		StatusCode: int16(statusCode),
		Success:    success,
		Message:    message,
		Meta:       m,
		Data:       data,
	})
}
