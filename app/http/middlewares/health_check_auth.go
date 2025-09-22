package middlewares

import (
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// HealthCheckAuth 健康檢查授權中介軟體
func HealthCheckAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 檢查是否為本地連接
		if isLocalhost(clientIP) {
			c.Next()
			return
		}

		// 檢查是否為信任的代理網段（僅在生產環境）
		if cfg.Server.Mode == config.ProductionMode {
			if isTrustedProxy(clientIP, cfg.Server.TrustedProxies) {
				c.Next()
				return
			}
		}

		// 拒絕存取
		c.JSON(http.StatusForbidden, models.APIResponse{
			Status:  "error",
			Message: "健康檢查API僅限本地或授權網段存取",
			Data:    nil,
		})
		c.Abort()
	}
}

// PublicHealthCheckAuth 公開健康檢查授權中介軟體（用於 /health/proxy）
func PublicHealthCheckAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// /health/proxy 端點可以更寬鬆一些，但仍需要基本限制
		clientIP := c.ClientIP()

		// 檢查是否為本地連接
		if isLocalhost(clientIP) {
			c.Next()
			return
		}

		// 在生產環境中，檢查是否為信任的代理網段或內部網路
		if cfg.Server.Mode == config.ProductionMode {
			if isTrustedProxy(clientIP, cfg.Server.TrustedProxies) || isPrivateNetwork(clientIP) {
				c.Next()
				return
			}
		} else {
			// 開發環境允許所有請求（用於測試）
			c.Next()
			return
		}

		// 拒絕存取
		c.JSON(http.StatusForbidden, models.APIResponse{
			Status:  "error",
			Message: "此健康檢查API僅限內部網路存取",
			Data:    nil,
		})
		c.Abort()
	}
}

// isLocalhost 檢查是否為本地連接
func isLocalhost(ip string) bool {
	return ip == "127.0.0.1" || ip == "::1" || ip == "localhost"
}

// isTrustedProxy 檢查IP是否在信任的代理列表中
func isTrustedProxy(clientIP string, trustedProxies []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	for _, proxy := range trustedProxies {
		// 如果包含 CIDR 符號，解析為網段
		if strings.Contains(proxy, "/") {
			_, cidr, err := net.ParseCIDR(proxy)
			if err != nil {
				continue
			}
			if cidr.Contains(ip) {
				return true
			}
		} else {
			// 單個IP比較
			proxyIP := net.ParseIP(proxy)
			if proxyIP != nil && proxyIP.Equal(ip) {
				return true
			}
		}
	}
	return false
}

// isPrivateNetwork 檢查IP是否為私有網路地址
func isPrivateNetwork(clientIP string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	// 定義私有網路範圍
	privateNetworks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}

	for _, network := range privateNetworks {
		_, cidr, err := net.ParseCIDR(network)
		if err != nil {
			continue
		}
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
