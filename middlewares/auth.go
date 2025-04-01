package middlewares

import (
	"chat_app_backend/utils"

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
			accessToken, err = utils.GetAccessTokenByHeader(c)
			if err != nil {
				utils.ErrorResponse(c, 401, utils.ErrUnauthorized, "未授權的請求")
				c.Abort()
				return
			}
		}

		res, err := utils.ValidateAccessToken(accessToken)
		if err != nil || !res {
			utils.ErrorResponse(c, 401, utils.ErrInvalidToken, "無效的 Token")
			c.Abort()
			return
		}

		c.Next()
	}
}
