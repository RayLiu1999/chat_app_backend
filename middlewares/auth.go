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
		var err error

		if websocket.IsWebSocketUpgrade(c.Request) {
			accessToken = c.Query("token")
		} else {
			// 判斷是否帶bearer token
			accessToken, err = services.GetAccessTokenByHeader(c)
			if err != nil {
				c.JSON(401, gin.H{"error": "Unauthorized"})
				c.Abort()
				return
			}
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
