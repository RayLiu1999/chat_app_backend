package middlewares

import (
	"chat_app_backend/app/http/controllers"
	"chat_app_backend/app/models"
	"chat_app_backend/utils"
	"net/http"

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
				controllers.ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
				c.Abort()
				return
			}
		}

		res, err := utils.ValidateAccessToken(accessToken)
		if err != nil || !res {
			controllers.ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrInvalidToken})
			c.Abort()
			return
		}

		c.Next()
	}
}
