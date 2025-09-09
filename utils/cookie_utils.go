package utils

import (
	"chat_app_backend/config"

	"github.com/gin-gonic/gin"
)

func SetCookie(c *gin.Context, cfg *config.Config, name string, value string, maxAge int) {
	mainDomain := cfg.Server.MainDomain
	secure := true
	if cfg.Server.Mode == config.DevelopmentMode {
		mainDomain = "localhost"
		secure = false
	}

	c.SetCookie(name, value, maxAge, "/", mainDomain, secure, true)
}

func ClearCookie(c *gin.Context, cfg *config.Config, name string) {
	SetCookie(c, cfg, name, "", -1)
}
