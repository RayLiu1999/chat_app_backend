package middlewares

import (
	"chat_app_backend/services"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var accessToken string
		if websocket.IsWebSocketUpgrade(c.Request) {
			accessToken = c.Query("token")
		} else {
			// 判斷是否帶bearer token
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(401, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}

			// 驗證 token
			accessToken = authHeader[len("Bearer "):]
		}
		log.Printf("accessToken: %s", accessToken)

		res, err := services.ValidateAccessToken(accessToken)
		if err != nil || !res {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}
