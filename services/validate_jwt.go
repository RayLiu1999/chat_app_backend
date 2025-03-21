package services

import (
	"errors"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func GetUserFromToken(tokenString string) (string, error) {
	// 解析和驗證 JWT token
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().JWT.AccessToken.Secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*models.AccessTokenClaims)
	if !ok {
		return "", errors.New("invalid token")
	}

	return claims.UserID, nil
}

func GetAccessTokenByHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header")
	}

	// 驗證 token
	accessToken := authHeader[len("Bearer "):]
	return accessToken, nil
}

func GetUserIDFromHeader(c *gin.Context) (string, primitive.ObjectID, error) {
	accessToken, err := GetAccessTokenByHeader(c)
	if err != nil {
		return "", primitive.NilObjectID, err
	}

	userID, err := GetUserFromToken(accessToken)
	if err != nil {
		return "", primitive.NilObjectID, err
	}

	// 將 userID 從字符串轉換為 ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return "", primitive.NilObjectID, errors.New("invalid user ID")
	}

	return userID, objectID, nil
}
