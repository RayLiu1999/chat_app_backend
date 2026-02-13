package middlewares

import (
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// 用於測試的輔助函數，建立一個 gin 上下文
func setupTestRouter(req *http.Request) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	return c, w
}

func setupTestAppConfig() {
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:        "test_secret",
			AccessExpireMinutes: 60,
		},
	}
}

func TestAuthMiddleware(t *testing.T) {
	// 設置 Gin 為測試模式並設置配置
	gin.SetMode(gin.TestMode)
	setupTestAppConfig()

	// 用於生成令牌的虛擬用戶 ID
	dummyUserID := "60d5ecb8b3920215a8204803" // 示例 ObjectID

	t.Run("HTTP 請求沒有令牌", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		c, w := setupTestRouter(req)

		// 執行
		Auth()(c)

		// 斷言
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, models.ErrUnauthorized, response.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("HTTP 請求使用無效令牌", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		c, w := setupTestRouter(req)

		// 執行
		Auth()(c)

		// 斷言
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, models.ErrInvalidToken, response.Code)
		assert.True(t, c.IsAborted())
	})

	t.Run("HTTP 請求使用有效令牌", func(t *testing.T) {
		// Setup
		tokenRes, err := utils.GenAccessToken(dummyUserID)
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer "+tokenRes.Token)
		c, w := setupTestRouter(req)

		// 新增虛擬處理程式以檢查是否被呼叫
		c.Next()

		// 執行
		Auth()(c)

		// 斷言
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		assert.False(t, c.IsAborted())
	})

	t.Run("WebSocket 請求使用有效令牌", func(t *testing.T) {
		// Setup
		tokenRes, err := utils.GenAccessToken(dummyUserID)
		assert.NoError(t, err)

		req, _ := http.NewRequest(http.MethodGet, "/?token="+tokenRes.Token, nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		c, w := setupTestRouter(req)

		// 執行
		Auth()(c)

		// 斷言
		assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		assert.False(t, c.IsAborted())
	})

	t.Run("WebSocket 請求沒有令牌", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Upgrade", "websocket")
		req.Header.Set("Connection", "Upgrade")
		c, w := setupTestRouter(req)

		// 執行
		Auth()(c)

		// 斷言
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.True(t, c.IsAborted())
	})
}
