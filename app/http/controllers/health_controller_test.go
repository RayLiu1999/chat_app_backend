package controllers

import (
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewHealthController 測試創建 HealthController
func TestNewHealthController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}

	controller := NewHealthController(cfg, nil)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
}

// TestHealthController_HealthCheck 測試健康檢查
func TestHealthController_HealthCheck(t *testing.T) {
	t.Run("健康檢查 - MongoDB 未初始化", func(t *testing.T) {
		controller := NewHealthController(&config.Config{}, nil)

		router := setupTestRouter()
		router.GET("/health", controller.HealthCheck)

		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		// 檢查返回的數據結構
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)

		// 檢查 status 欄位（實際 API 回應的是 "status" 而非 "success"）
		status, exists := data["status"]
		assert.True(t, exists, "data['status'] should exist")
		assert.Equal(t, "ok", status)

		services, ok := data["services"].(map[string]interface{})
		assert.True(t, ok)

		mongo, ok := services["mongo"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "not_initialized", mongo["status"])
	})
}
