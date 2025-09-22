package controllers

import (
	"chat_app_backend/app/providers"
	"chat_app_backend/config"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthController struct {
	config       *config.Config
	mongoConnect *providers.MongoWrapper
}

func NewHealthController(cfg *config.Config, mongodb *providers.MongoWrapper) *HealthController {
	return &HealthController{
		config:       cfg,
		mongoConnect: mongodb,
	}
}

func (hc *HealthController) HealthCheck(c *gin.Context) {
	status := "ok"

	// 預設 mongo 狀態
	mongoStatus := "ok"
	mongoError := ""

	if hc.mongoConnect == nil {
		mongoStatus = "not_initialized"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		defer cancel()
		if err := hc.mongoConnect.Ping(ctx); err != nil {
			mongoStatus = "error"
			mongoError = err.Error()
			// 只有資料庫異常時將整體狀態標記為 degraded
			status = "degraded"
		}
	}

	SuccessResponse(c, gin.H{
		"success": true,
		"status":  status, // ok | degraded
		"services": gin.H{
			"mongo": gin.H{
				"status": mongoStatus,
				"error":  mongoError,
			},
		},
		"timestamp": time.Now().UTC(),
	}, "Health check completed")
}

// ProxyCheck 代理配置檢查
func (hc *HealthController) ProxyCheck(c *gin.Context) {
	// 獲取客戶端IP信息
	clientIP := c.ClientIP()
	remoteAddr := c.Request.RemoteAddr
	forwardedFor := c.GetHeader("X-Forwarded-For")
	realIP := c.GetHeader("X-Real-IP")
	forwardedProto := c.GetHeader("X-Forwarded-Proto")

	// 獲取請求信息
	userAgent := c.GetHeader("User-Agent")
	requestURI := c.Request.RequestURI
	method := c.Request.Method

	SuccessResponse(c, gin.H{
		"success": true,
		"proxy_info": gin.H{
			"client_ip":         clientIP,
			"remote_addr":       remoteAddr,
			"x_forwarded_for":   forwardedFor,
			"x_real_ip":         realIP,
			"x_forwarded_proto": forwardedProto,
		},
		"request_info": gin.H{
			"method":     method,
			"uri":        requestURI,
			"user_agent": userAgent,
		},
		"timestamp": time.Now().UTC(),
	}, "Proxy check completed")
}

// DetailedHealthCheck 詳細健康檢查
func (hc *HealthController) DetailedHealthCheck(c *gin.Context) {
	status := "ok"

	// 預設 mongo 狀態
	mongoStatus := "ok"
	mongoError := ""

	if hc.mongoConnect == nil {
		mongoStatus = "not_initialized"
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
		defer cancel()
		if err := hc.mongoConnect.Ping(ctx); err != nil {
			mongoStatus = "error"
			mongoError = err.Error()
			// 只有資料庫異常時將整體狀態標記為 degraded
			status = "degraded"
		}
	}

	// 代理配置檢查
	clientIP := c.ClientIP()
	forwardedFor := c.GetHeader("X-Forwarded-For")
	realIP := c.GetHeader("X-Real-IP")

	// 組合響應
	SuccessResponse(c, gin.H{
		"success": true,
		"status":  status, // ok | degraded
		"services": gin.H{
			"mongo": gin.H{
				"status": mongoStatus,
				"error":  mongoError,
			},
		},
		"proxy_info": gin.H{
			"client_ip":                  clientIP,
			"x_forwarded_for":            forwardedFor,
			"x_real_ip":                  realIP,
			"trusted_proxies_configured": len(hc.config.Server.TrustedProxies) > 0,
			"trusted_proxies":            hc.config.Server.TrustedProxies,
		},
		"config_info": gin.H{
			"mode":            hc.config.Server.Mode,
			"allowed_origins": hc.config.Server.AllowedOrigins,
		},
		"timestamp": time.Now().UTC(),
	}, "Detailed health check completed")
}
