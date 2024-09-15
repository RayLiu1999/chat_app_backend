package middlewares

import (
	"chat_app_backend/models"
	"chat_app_backend/services"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 判斷是否帶bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 驗證 token
		accessToken := authHeader[len("Bearer "):]
		res, err := services.ValidateJWT(accessToken, models.AccessTokenClaims{})
		if err != nil || !res {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
