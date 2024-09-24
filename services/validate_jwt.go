package services

import (
	"errors"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/models"

	"github.com/dgrijalva/jwt-go"
)

// ValidateJWT 解析和驗證 JWT token
func ValidateAccessToken(tokenString string) (bool, error) {
	cfg := config.GetConfig()
	accessTokenJwtSecret := []byte(cfg.JWT.AccessToken.Secret)

	// 解析和驗證 JWT 簽章
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 檢查簽名方法是否為預期的 HMAC 方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return accessTokenJwtSecret, nil
	})

	if err != nil {
		return false, err
	}

	// 檢查 token 是否有效
	if claims, ok := token.Claims.(*models.AccessTokenClaims); ok && token.Valid {
		// 檢查 token 是否過期
		if claims.ExpiresAt < time.Now().Unix() {
			return false, errors.New("token is expired")
		}
		return true, nil
	}

	return false, errors.New("invalid token")
}
