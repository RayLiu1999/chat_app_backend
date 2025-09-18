package middlewares

import (
	"chat_app_backend/controllers"
	"chat_app_backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifyCSRFToken 驗證 CSRF token 的中介軟體
func VerifyCSRFToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// GET 請求不需要驗證 CSRF Token
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		// 從 header 中取得 CSRF token
		headerCSRFToken := c.GetHeader("X-CSRF-TOKEN")
		if headerCSRFToken == "" {
			controllers.ErrorResponse(c, http.StatusForbidden, models.MessageOptions{
				Code:    models.ErrInvalidToken,
				Message: "缺少 CSRF token",
			})
			c.Abort()
			return
		}

		// 從 cookie 中取得 CSRF token
		cookieCSRFToken, err := c.Cookie("csrf_token")
		if err != nil {
			controllers.ErrorResponse(c, http.StatusForbidden, models.MessageOptions{
				Code:    models.ErrInvalidToken,
				Message: "CSRF token cookie 不存在",
			})
			c.Abort()
			return
		}

		// 比對兩個 token 是否相同
		if headerCSRFToken != cookieCSRFToken {
			controllers.ErrorResponse(c, http.StatusForbidden, models.MessageOptions{
				Code:    models.ErrInvalidToken,
				Message: "CSRF token 驗證失敗",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
