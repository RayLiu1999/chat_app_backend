package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuccessResponse sends a success JSON response.
func SuccessResponse(c *gin.Context, data any, message string) {
	c.JSON(http.StatusOK, models.APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// ErrorResponse sends an error JSON response.
func ErrorResponse(c *gin.Context, statusCode int, options models.MessageOptions) {
	// 僅在測試環境回傳詳細錯誤訊息
	var details any
	if !config.IsProduction() {
		details = options.Details
	}

	c.JSON(statusCode, models.APIResponse{
		Status:  "error",
		Code:    options.Code,
		Message: options.Message,
		Details: details,
	})
}
