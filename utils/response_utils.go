// utils/response.go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type MessageOptions struct {
	Code        int
	Message     string
	Displayable bool
}

// MessageResponseOptions 是可選參數的設置函數
type MessageResponseOptions func(*MessageOptions)

// WithCode 設置錯誤代碼
func WithCode(code int) MessageResponseOptions {
	return func(o *MessageOptions) {
		o.Code = code
	}
}

// WithMessage 設置錯誤消息
func WithMessage(message string) MessageResponseOptions {
	return func(o *MessageOptions) {
		o.Message = message
	}
}

// WithDisplayable 設置是否可顯示
func WithDisplayable(displayable bool) MessageResponseOptions {
	return func(o *MessageOptions) {
		o.Displayable = displayable
	}
}

// APIResponse represents the standard API response structure.
type APIResponse struct {
	Status      string `json:"status"`            // "success" 或 "error"
	Code        int    `json:"code"`              // 自定義錯誤碼，例如 1001 表示 "用戶不存在"
	Message     string `json:"message,omitempty"` // 訊息內容，可選
	Displayable bool   `json:"displayable"`       // 是否可顯示給用戶
	Data        any    `json:"data,omitempty"`    // 可選的數據
}

// SuccessResponse sends a success JSON response.
func SuccessResponse(c *gin.Context, data interface{}, options MessageOptions) {
	c.JSON(http.StatusOK, APIResponse{
		Status:      "success",
		Code:        options.Code,
		Message:     options.Message,
		Displayable: options.Displayable,
		Data:        data,
	})
}

// ErrorResponse sends an error JSON response.
func ErrorResponse(c *gin.Context, statusCode int, options MessageOptions) {
	c.JSON(statusCode, APIResponse{
		Status:      "error",
		Code:        options.Code,
		Message:     options.Message,
		Displayable: options.Displayable,
	})
}
