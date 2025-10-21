package middlewares

import (
	"chat_app_backend/app/models"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestVerifyCSRFToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	dummyToken := "super-secret-csrf-token"

	t.Run("GET 請求應該通過", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		c, _ := setupTestRouter(req)

		// 執行
		VerifyCSRFToken()(c)

		// 斷言
		assert.False(t, c.IsAborted())
	})

	t.Run("POST 請求使用有效令牌應該通過", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-CSRF-TOKEN", dummyToken)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: dummyToken})
		c, _ := setupTestRouter(req)

		// 執行
		VerifyCSRFToken()(c)

		// 斷言
		assert.False(t, c.IsAborted())
	})

	t.Run("POST 請求缺少標頭令牌應該失敗", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: dummyToken})
		c, w := setupTestRouter(req)

		// 執行
		VerifyCSRFToken()(c)

		// 斷言
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, models.ErrInvalidToken, response.Code)
	})

	t.Run("POST 請求缺少 Cookie 令牌應該失敗", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X-CSRF-TOKEN", dummyToken)
		c, w := setupTestRouter(req)

		// 執行
		VerifyCSRFToken()(c)

		// 斷言
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, models.ErrInvalidToken, response.Code)
	})

	t.Run("POST 請求令牌不匹配應該失敗", func(t *testing.T) {
		// Setup
		req, _ := http.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set("X--CSRF-TOKEN", "token1")
		req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "token2"})
		c, w := setupTestRouter(req)

		// 執行
		VerifyCSRFToken()(c)

		// 斷言
		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, models.ErrInvalidToken, response.Code)
	})
}
