package middlewares

import (
	"chat_app_backend/app/models"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVerifyOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 定義允許的來源列表
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://app.example.com",
		"https://admin.example.com",
	}

	t.Run("請求來自允許的來源 - 應該通過", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		c, w := setupTestRouter(req)

		// Execute
		VerifyOrigin(allowedOrigins)(c)

		// Assert
		assert.False(t, c.IsAborted(), "請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("請求來自另一個允許的來源 - 應該通過", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://app.example.com")
		c, w := setupTestRouter(req)

		// Execute
		VerifyOrigin(allowedOrigins)(c)

		// Assert
		assert.False(t, c.IsAborted(), "請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("請求來自不允許的來源 - 應該被拒絕", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://malicious-site.com")
		c, w := setupTestRouter(req)

		// Execute
		VerifyOrigin(allowedOrigins)(c)

		// Assert
		assert.True(t, c.IsAborted(), "請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, models.ErrInvalidToken, response.Code)
	})

	t.Run("請求沒有 Origin 標頭 - 應該被拒絕", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		// 不設置 Origin 標頭
		c, w := setupTestRouter(req)

		// Execute
		VerifyOrigin(allowedOrigins)(c)

		// Assert
		assert.True(t, c.IsAborted(), "請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("空的允許來源列表 - 所有請求都應該被拒絕", func(t *testing.T) {
		// Setup
		emptyOrigins := []string{}
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		c, w := setupTestRouter(req)

		// Execute
		VerifyOrigin(emptyOrigins)(c)

		// Assert
		assert.True(t, c.IsAborted(), "請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Origin 大小寫敏感測試 - 應該被拒絕", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "HTTP://LOCALHOST:3000") // 大寫
		c, w := setupTestRouter(req)

		// Execute
		VerifyOrigin(allowedOrigins)(c)

		// Assert
		assert.True(t, c.IsAborted(), "大小寫不匹配的請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestIsValidOrigin(t *testing.T) {
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://app.example.com",
	}

	t.Run("有效的來源應該返回 true", func(t *testing.T) {
		result := isValidOrigin("http://localhost:3000", allowedOrigins)
		assert.True(t, result)
	})

	t.Run("無效的來源應該返回 false", func(t *testing.T) {
		result := isValidOrigin("https://evil.com", allowedOrigins)
		assert.False(t, result)
	})

	t.Run("空字串應該返回 false", func(t *testing.T) {
		result := isValidOrigin("", allowedOrigins)
		assert.False(t, result)
	})

	t.Run("空的允許列表應該返回 false", func(t *testing.T) {
		result := isValidOrigin("http://localhost:3000", []string{})
		assert.False(t, result)
	})
}
