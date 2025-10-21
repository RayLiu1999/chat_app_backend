package middlewares

import (
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupHealthCheckConfig 建立測試用的配置
func setupHealthCheckConfig(mode config.ModeConfig) *config.Config {
	return &config.Config{
		Server: config.ServerConfig{
			Mode: mode,
			TrustedProxies: []string{
				"10.0.0.0/8",
				"192.168.1.100",
			},
		},
	}
}

func TestHealthCheckAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("本地 IP (127.0.0.1) 應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "127.0.0.1:12345"

		// Execute
		HealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "本地請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("本地 IP (::1) IPv6 應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "[::1]:12345"

		// Execute
		HealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "本地 IPv6 請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("生產模式 - 信任的代理網段應該被允許 (CIDR)", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "10.0.1.50:12345" // 在 10.0.0.0/8 網段內

		// Execute
		HealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "信任的代理網段請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("生產模式 - 信任的單個 IP 應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "192.168.1.100:12345"

		// Execute
		HealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "信任的 IP 請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("生產模式 - 不信任的外部 IP 應該被拒絕", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "203.0.113.1:12345" // 外部 IP

		// Execute
		HealthCheckAuth(cfg)(c)

		// Assert
		assert.True(t, c.IsAborted(), "外部 IP 請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "健康檢查API僅限本地或授權網段存取")
	})

	t.Run("開發模式 - 不信任的外部 IP 也應該被拒絕", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.DevelopmentMode)
		req, _ := http.NewRequest(http.MethodGet, "/health", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "203.0.113.1:12345"

		// Execute
		HealthCheckAuth(cfg)(c)

		// Assert
		assert.True(t, c.IsAborted(), "開發模式下外部 IP 請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestPublicHealthCheckAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("本地 IP 應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health/proxy", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "127.0.0.1:12345"

		// Execute
		PublicHealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "本地請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("生產模式 - 信任的代理應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health/proxy", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "10.0.1.50:12345"

		// Execute
		PublicHealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "信任的代理請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("生產模式 - 私有網路 IP 應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health/proxy", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "192.168.100.50:12345" // 私有網路，但不在信任列表

		// Execute
		PublicHealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "私有網路請求不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})

	t.Run("生產模式 - 外部公網 IP 應該被拒絕", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.ProductionMode)
		req, _ := http.NewRequest(http.MethodGet, "/health/proxy", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "8.8.8.8:12345" // Google DNS (公網 IP)

		// Execute
		PublicHealthCheckAuth(cfg)(c)

		// Assert
		assert.True(t, c.IsAborted(), "公網 IP 請求應該被中止")
		assert.Equal(t, http.StatusForbidden, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "此健康檢查API僅限內部網路存取")
	})

	t.Run("開發模式 - 所有請求都應該被允許", func(t *testing.T) {
		// Setup
		cfg := setupHealthCheckConfig(config.DevelopmentMode)
		req, _ := http.NewRequest(http.MethodGet, "/health/proxy", nil)
		c, w := setupTestRouter(req)
		c.Request.RemoteAddr = "8.8.8.8:12345" // 公網 IP

		// Execute
		PublicHealthCheckAuth(cfg)(c)

		// Assert
		assert.False(t, c.IsAborted(), "開發模式下所有請求都不應該被中止")
		assert.NotEqual(t, http.StatusForbidden, w.Code)
	})
}

func TestIsLocalhost(t *testing.T) {
	t.Run("127.0.0.1 應該被識別為 localhost", func(t *testing.T) {
		result := isLocalhost("127.0.0.1")
		assert.True(t, result)
	})

	t.Run("::1 應該被識別為 localhost", func(t *testing.T) {
		result := isLocalhost("::1")
		assert.True(t, result)
	})

	t.Run("localhost 字串應該被識別為 localhost", func(t *testing.T) {
		result := isLocalhost("localhost")
		assert.True(t, result)
	})

	t.Run("外部 IP 不應該被識別為 localhost", func(t *testing.T) {
		result := isLocalhost("192.168.1.1")
		assert.False(t, result)
	})

	t.Run("空字串不應該被識別為 localhost", func(t *testing.T) {
		result := isLocalhost("")
		assert.False(t, result)
	})
}

func TestIsTrustedProxy(t *testing.T) {
	trustedProxies := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.1.100",
	}

	t.Run("CIDR 網段內的 IP 應該被信任", func(t *testing.T) {
		result := isTrustedProxy("10.0.1.50", trustedProxies)
		assert.True(t, result)
	})

	t.Run("另一個 CIDR 網段內的 IP 應該被信任", func(t *testing.T) {
		result := isTrustedProxy("172.16.5.10", trustedProxies)
		assert.True(t, result)
	})

	t.Run("單個信任的 IP 應該被信任", func(t *testing.T) {
		result := isTrustedProxy("192.168.1.100", trustedProxies)
		assert.True(t, result)
	})

	t.Run("不在信任列表中的 IP 不應該被信任", func(t *testing.T) {
		result := isTrustedProxy("203.0.113.1", trustedProxies)
		assert.False(t, result)
	})

	t.Run("無效的 IP 格式不應該被信任", func(t *testing.T) {
		result := isTrustedProxy("invalid-ip", trustedProxies)
		assert.False(t, result)
	})

	t.Run("空字串不應該被信任", func(t *testing.T) {
		result := isTrustedProxy("", trustedProxies)
		assert.False(t, result)
	})
}

func TestIsPrivateNetwork(t *testing.T) {
	t.Run("10.x.x.x 網段應該被識別為私有網路", func(t *testing.T) {
		result := isPrivateNetwork("10.0.0.1")
		assert.True(t, result)

		result = isPrivateNetwork("10.255.255.255")
		assert.True(t, result)
	})

	t.Run("172.16.x.x - 172.31.x.x 網段應該被識別為私有網路", func(t *testing.T) {
		result := isPrivateNetwork("172.16.0.1")
		assert.True(t, result)

		result = isPrivateNetwork("172.31.255.255")
		assert.True(t, result)
	})

	t.Run("192.168.x.x 網段應該被識別為私有網路", func(t *testing.T) {
		result := isPrivateNetwork("192.168.0.1")
		assert.True(t, result)

		result = isPrivateNetwork("192.168.255.255")
		assert.True(t, result)
	})

	t.Run("公網 IP 不應該被識別為私有網路", func(t *testing.T) {
		result := isPrivateNetwork("8.8.8.8")
		assert.False(t, result)

		result = isPrivateNetwork("1.1.1.1")
		assert.False(t, result)
	})

	t.Run("無效的 IP 不應該被識別為私有網路", func(t *testing.T) {
		result := isPrivateNetwork("invalid-ip")
		assert.False(t, result)
	})

	t.Run("空字串不應該被識別為私有網路", func(t *testing.T) {
		result := isPrivateNetwork("")
		assert.False(t, result)
	})
}
