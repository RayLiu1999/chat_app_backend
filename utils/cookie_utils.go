package utils

import (
	"chat_app_backend/config"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func SetCookie(c *gin.Context, cfg *config.Config, name string, value string, maxAge int, httpOnly bool) {
	mainDomain := normalizeCookieDomain(cfg.Server.MainDomain)
	secure := shouldUseSecureCookie(c, cfg)

	if cfg.Server.Mode == config.ProductionMode {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
		mainDomain = ""
	}

	c.SetCookie(name, value, maxAge, "/", mainDomain, secure, httpOnly)
}

func ClearCookie(c *gin.Context, cfg *config.Config, name string) {
	SetCookie(c, cfg, name, "", -1, true)
}

func normalizeCookieDomain(raw string) string {
	domain := strings.TrimSpace(raw)
	if domain == "" {
		return ""
	}

	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		if parsedURL, err := url.Parse(domain); err == nil {
			domain = parsedURL.Hostname()
		}
	}

	if host, _, found := strings.Cut(domain, "/"); found {
		domain = host
	}

	if host, _, found := strings.Cut(domain, ":"); found {
		domain = host
	}

	domain = strings.TrimPrefix(domain, ".")
	if domain == "localhost" {
		return ""
	}

	return domain
}

func shouldUseSecureCookie(c *gin.Context, cfg *config.Config) bool {
	if cfg.Server.Mode == config.DevelopmentMode {
		return false
	}

	if c != nil && c.Request != nil {
		if c.Request.TLS != nil {
			return true
		}

		if strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			return true
		}

		if strings.EqualFold(c.GetHeader("X-Forwarded-Ssl"), "on") {
			return true
		}
	}

	return true
}
