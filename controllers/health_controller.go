package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"context"
	"net/http"
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
	// 檢查請求是否來自 localhost
	clientIP := c.ClientIP()
	if clientIP != "127.0.0.1" && clientIP != "::1" && clientIP != "localhost" {
		ErrorResponse(c, http.StatusForbidden, models.MessageOptions{})
		return
	}

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
