package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CSRFService struct {
	Secret string
}

func NewCSRFService(secret string) *CSRFService {
	return &CSRFService{Secret: secret}
}

func (s *CSRFService) Middleware() gin.HandlerFunc {
	return csrf.Middleware(csrf.Options{
		Secret: s.Secret,
		ErrorFunc: func(c *gin.Context) {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token mismatch"})
			c.Abort()
		},
	})
}

func (s *CSRFService) GetToken(c *gin.Context) string {
	return csrf.GetToken(c)
}
