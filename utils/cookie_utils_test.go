package utils

import (
	"chat_app_backend/config"
	"net/http"
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
		c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")

		SetCookie(c, cfg, "test_cookie", "test_value", 3600, true)
		cookie := w.Header().Get("Set-Cookie")
		assert.Contains(t, cookie, "test_cookie=test_value")
		assert.Contains(t, cookie, "Domain=example.com")
		assert.Contains(t, cookie, "Secure")
		assert.Contains(t, cookie, "SameSite=None")
		assert.Contains(t, cookie, "HttpOnly")
		assert.Contains(t, cookie, "Max-Age=3600")
	})

	c, w = setupGin()
	cfg.Server.Mode = config.DevelopmentMode
	t.Run("開發模式", func(t *testing.T) {
		c.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
		SetCookie(c, cfg, "dev_cookie", "dev_value", 0, false)
		cookie := w.Header().Get("Set-Cookie")
		assert.Contains(t, cookie, "dev_cookie=dev_value")
		assert.NotContains(t, cookie, "Domain=")
		assert.NotContains(t, cookie, "Secure")
		assert.Contains(t, cookie, "SameSite=Lax")
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

func TestNormalizeCookieDomain(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{name: "純網域", input: "liu-yucheng.com", expect: "liu-yucheng.com"},
		{name: "含協定", input: "https://liu-yucheng.com", expect: "liu-yucheng.com"},
		{name: "含路徑", input: "liu-yucheng.com/api", expect: "liu-yucheng.com"},
		{name: "含 port", input: "liu-yucheng.com:8443", expect: "liu-yucheng.com"},
		{name: "localhost 不設 domain", input: "localhost", expect: ""},
		{name: "前綴點", input: ".liu-yucheng.com", expect: "liu-yucheng.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expect, normalizeCookieDomain(tt.input))
		})
	}
}

func TestShouldUseSecureCookie(t *testing.T) {
	t.Run("production + forwarded https", func(t *testing.T) {
		c, _ := setupGin()
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")

		cfg := &config.Config{Server: config.ServerConfig{Mode: config.ProductionMode}}
		assert.True(t, shouldUseSecureCookie(c, cfg))
	})

	t.Run("production + plain http", func(t *testing.T) {
		c, _ := setupGin()
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		cfg := &config.Config{Server: config.ServerConfig{Mode: config.ProductionMode}}
		assert.True(t, shouldUseSecureCookie(c, cfg))
	})

	t.Run("development 一律不 secure", func(t *testing.T) {
		c, _ := setupGin()
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("X-Forwarded-Proto", "https")

		cfg := &config.Config{Server: config.ServerConfig{Mode: config.DevelopmentMode}}
		assert.False(t, shouldUseSecureCookie(c, cfg))
	})
}
