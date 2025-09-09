package middlewares

import (
	"chat_app_backend/controllers"
	"chat_app_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func VerifyOrigin(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 驗證Origin
		origin := c.GetHeader("Origin")

		// 檢查Origin是否在允許的列表中
		if !isValidOrigin(origin, allowedOrigins) {
			controllers.ErrorResponse(c, http.StatusForbidden, models.MessageOptions{Code: models.ErrInvalidToken})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isValidOrigin(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}
