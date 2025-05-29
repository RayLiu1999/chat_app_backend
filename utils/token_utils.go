package utils

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TokenResponse struct {
	Token     string
	ExpiresAt int64
}

var jwtSecret = []byte(config.GetConfig().JWT.AccessToken.Secret)

// 生成 access token
func GenAccessToken(userID string) (TokenResponse, error) {
	cfg := config.GetConfig()
	accessTokenJwtSecret := []byte(cfg.JWT.AccessToken.Secret)
	accessTokenExpireHours := cfg.JWT.AccessToken.ExpireHours

	// 將小時轉換為分鐘
	accessTokenExpireDuration := time.Duration(accessTokenExpireHours*60) * time.Minute
	expiresAt := time.Now().Add(accessTokenExpireDuration).Unix()

	// 設置 access token 的聲明
	accessTokenClaims := models.AccessTokenClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	// 生成 access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)

	accessTokenString, err := accessToken.SignedString(accessTokenJwtSecret)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		Token:     accessTokenString,
		ExpiresAt: expiresAt,
	}, nil
}

// 生成 refresh token
func GenRefreshToken(userID string) (TokenResponse, error) {
	cfg := config.GetConfig()
	refreshTokenJwtSecret := []byte(cfg.JWT.RefreshToken.Secret)
	refreshTokenExpireHours := cfg.JWT.RefreshToken.ExpireHours

	// 將小時轉換為分鐘
	refreshTokenExpireDuration := time.Duration(refreshTokenExpireHours*60) * time.Minute
	expiresAt := time.Now().Add(refreshTokenExpireDuration).Unix()

	// 設置 refresh token 的聲明
	refreshTokenClaims := models.RefreshTokenClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}

	// 生成 access token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)

	refreshTokenString, err := refreshToken.SignedString(refreshTokenJwtSecret)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		Token:     refreshTokenString,
		ExpiresAt: expiresAt,
	}, nil
}

// 驗證 JWT access token
func ValidateAccessToken(tokenString string) (bool, error) {
	// 解析和驗證 JWT 簽章
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 檢查簽名方法是否為預期的 HMAC 方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
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

// 從 token 中獲取用戶 ID
func GetUserFromToken(tokenString string) (string, error) {
	// 解析和驗證 JWT token
	token, err := jwt.ParseWithClaims(tokenString, &models.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
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

// 從 HTTP 請求頭中獲取 access token
func GetAccessTokenByHeader(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization header")
	}

	// 驗證 token
	accessToken := authHeader[len("Bearer "):]
	return accessToken, nil
}

// 從 HTTP 請求頭中獲取用戶 ID
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
