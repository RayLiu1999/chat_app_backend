package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func VerifyCsrfToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 取得 POST 內容
		csrfName := c.GetHeader("X-CSRF-Name")
		csrfToken := c.GetHeader("X-CSRF-Token")

		// 比對 CSRF Token 與 Cookie 中的 Token 是否相同
		cookieCsrfToken, err := c.Cookie(csrfName)
		if err != nil || csrfToken != cookieCsrfToken {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF Token"})
			c.Abort()
			return
		}

		c.Next()
	}
}
