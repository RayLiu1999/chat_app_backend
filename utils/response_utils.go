// utils/response.go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse represents the standard API response structure.
type APIResponse struct {
	Status      string `json:"status"`      // "success" 或 "error"
	Code        int    `json:"code"`        // 自定義錯誤碼，例如 1001 表示 "用戶不存在"
	Message     string `json:"message"`     // 訊息內容
	Displayable bool   `json:"displayable"` // 是否可顯示給用戶
	Data        any    `json:"data"`        // 可選的數據
}

// SuccessResponse sends a success JSON response.
func SuccessResponse(c *gin.Context, data interface{}, message string, Code int, opts ...bool) {
	displayable := false
	if len(opts) > 0 {
		displayable = opts[0]
	}

	c.JSON(http.StatusOK, APIResponse{
		Status:      "success",
		Code:        Code,
		Message:     message,
		Displayable: displayable,
		Data:        data,
	})
}

// ErrorResponse sends an error JSON response.
func ErrorResponse(c *gin.Context, statusCode int, Code int, message string, opts ...bool) {
	displayable := false
	if len(opts) > 0 {
		displayable = opts[0]
	}

	c.JSON(statusCode, APIResponse{
		Status:      "error",
		Code:        Code,
		Message:     message,
		Displayable: displayable,
	})
}
