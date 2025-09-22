package models

// ErrorResponse 定義統一的錯誤回應格式
type ErrorResponse struct {
	Error string `json:"error"`
}
