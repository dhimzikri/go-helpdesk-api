package utils

import (
	"github.com/gin-gonic/gin"
)

type ResponseWithData struct {
	ResponseCode int    `json:"responsecode"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	Data         any    `json:"data"`
}

type ResponseWithoutData struct {
	ResponseCode int    `json:"responsecode"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

// RespondJSON will send a structured response with success, message, and data
func RespondJSON(c *gin.Context, status int, success bool, message string, data interface{}) {
	c.JSON(status, gin.H{
		"success": success,
		"message": message,
		"data":    data,
	})
}

// Response is the generic response handler for Gin, replacing the previous http.ResponseWriter approach
func Response(c *gin.Context, code int, message string, payload interface{}) {
	var response any
	status := "success"

	if code >= 400 {
		status = "failed"
	}

	if payload != nil {
		response = &ResponseWithData{
			ResponseCode: code,
			Status:       status,
			Message:      message,
			Data:         payload,
		}
	} else {
		response = &ResponseWithoutData{
			ResponseCode: code,
			Status:       status,
			Message:      message,
		}
	}

	// Send the response
	c.JSON(code, response)
}
