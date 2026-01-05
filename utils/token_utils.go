package utils

import (
	"chat_app_backend/config"
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AccessTokenClaims 定義了 access token 中的聲明
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims 定義了 refresh token 中的聲明
type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	Token     string
	ExpiresAt int64
}

// 生成 access token
func GenAccessToken(userID string) (TokenResponse, error) {
	if config.AppConfig == nil {
		return TokenResponse{}, errors.New("config not loaded")
	}

	accessTokenJwtSecret := []byte(config.AppConfig.JWT.AccessSecret)
	accessTokenExpireHours := config.AppConfig.JWT.AccessExpireHours

	// 將小時轉換為分鐘
	accessTokenExpireDuration := time.Duration(accessTokenExpireHours*60) * time.Minute
	expireTime := time.Now().Add(accessTokenExpireDuration)
	expiresAt := jwt.NewNumericDate(expireTime)

	// 設置 access token 的聲明
	accessTokenClaims := &AccessTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
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
		ExpiresAt: expireTime.Unix(),
	}, nil
}

// 生成 refresh token
func GenRefreshToken(userID string) (TokenResponse, error) {
	if config.AppConfig == nil {
		return TokenResponse{}, errors.New("config not loaded")
	}

	refreshTokenJwtSecret := []byte(config.AppConfig.JWT.RefreshSecret)
	refreshTokenExpireHours := config.AppConfig.JWT.RefreshExpireHours

	// 將小時轉換為分鐘
	refreshTokenExpireDuration := time.Duration(refreshTokenExpireHours*60) * time.Minute
	expireTime := time.Now().Add(refreshTokenExpireDuration)
	expiresAt := jwt.NewNumericDate(expireTime)

	// 設置 refresh token 的聲明
	refreshTokenClaims := &RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
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
		ExpiresAt: expireTime.Unix(),
	}, nil
}

// 驗證 JWT access token
func ValidateAccessToken(tokenString string) (bool, error) {
	if config.AppConfig == nil {
		return false, errors.New("config not loaded")
	}

	jwtSecret := []byte(config.AppConfig.JWT.AccessSecret)

	// 解析和驗證 JWT 簽章
	// v5 ParseWithClaims 會自動驗證過期時間 (exp)
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
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
	if _, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return true, nil
	}

	return false, errors.New("invalid token")
}

// 從 token 中獲取用戶 ID
func GetUserFromToken(tokenString string) (string, primitive.ObjectID, error) {
	if config.AppConfig == nil {
		return "", primitive.NilObjectID, errors.New("config not loaded")
	}

	jwtSecret := []byte(config.AppConfig.JWT.AccessSecret)

	// 解析和驗證 JWT token
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return "", primitive.NilObjectID, err
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return "", primitive.NilObjectID, errors.New("invalid token")
	}

	userObjectID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return "", primitive.NilObjectID, errors.New("invalid user ID")
	}

	return claims.UserID, userObjectID, nil
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
	// 優先檢查上下文中的用戶資訊（用於測試環境）
	if userID, exists := c.Get("user_id"); exists {
		if userObjectID, exists := c.Get("user_object_id"); exists {
			return userID.(string), userObjectID.(primitive.ObjectID), nil
		}
	}

	// 正常流程：從 Authorization header 解析 JWT token
	accessToken, err := GetAccessTokenByHeader(c)
	if err != nil {
		return "", primitive.NilObjectID, err
	}

	userID, userObjectID, err := GetUserFromToken(accessToken)
	if err != nil {
		return "", primitive.NilObjectID, err
	}

	return userID, userObjectID, nil
}
