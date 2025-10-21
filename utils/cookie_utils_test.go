package utils

import (
	"chat_app_backend/config"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupGin() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func TestSetCookie(t *testing.T) {
	c, w := setupGin()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode:       config.ProductionMode,
			MainDomain: "example.com",
		},
	}

	t.Run("生產模式", func(t *testing.T) {
		SetCookie(c, cfg, "test_cookie", "test_value", 3600, true)
		cookie := w.Header().Get("Set-Cookie")
		assert.Contains(t, cookie, "test_cookie=test_value")
		assert.Contains(t, cookie, "Domain=example.com")
		assert.Contains(t, cookie, "Secure")
		assert.Contains(t, cookie, "HttpOnly")
		assert.Contains(t, cookie, "Max-Age=3600")
	})

	c, w = setupGin()
	cfg.Server.Mode = config.DevelopmentMode
	t.Run("開發模式", func(t *testing.T) {
		SetCookie(c, cfg, "dev_cookie", "dev_value", 0, false)
		cookie := w.Header().Get("Set-Cookie")
		assert.Contains(t, cookie, "dev_cookie=dev_value")
		assert.Contains(t, cookie, "Domain=localhost")
		assert.NotContains(t, cookie, "Secure")
		assert.NotContains(t, cookie, "HttpOnly")
	})
}

func TestClearCookie(t *testing.T) {
	c, w := setupGin()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Mode:       config.ProductionMode,
			MainDomain: "example.com",
		},
	}

	ClearCookie(c, cfg, "test_cookie")
	cookie := w.Header().Get("Set-Cookie")
	assert.Contains(t, cookie, "test_cookie=")
	// Gin 針對負數輸入將 Max-Age 設為 0 以立即使 cookie 過期。
	assert.Contains(t, cookie, "Max-Age=0")
}
