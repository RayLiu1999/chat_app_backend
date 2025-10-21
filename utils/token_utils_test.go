package utils

import (
	"chat_app_backend/config"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 為測試設置配置
func setupTokenConfig() {
	config.AppConfig = &config.Config{
		JWT: config.JWTConfig{
			AccessSecret:       "test_access_secret",
			RefreshSecret:      "test_refresh_secret",
			AccessExpireHours:  1,
			RefreshExpireHours: 24 * 7,
		},
	}
}

func TestGenAccessToken(t *testing.T) {
	setupTokenConfig()
	userID := primitive.NewObjectID().Hex()

	tokenRes, err := GenAccessToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenRes.Token)
	assert.True(t, tokenRes.ExpiresAt > time.Now().Unix())

	// 解析並驗證令牌
	token, err := jwt.ParseWithClaims(tokenRes.Token, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.AccessSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*AccessTokenClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
}

func TestGenRefreshToken(t *testing.T) {
	setupTokenConfig()
	userID := primitive.NewObjectID().Hex()

	tokenRes, err := GenRefreshToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenRes.Token)
	assert.True(t, tokenRes.ExpiresAt > time.Now().Unix())

	// 解析並驗證令牌
	token, err := jwt.ParseWithClaims(tokenRes.Token, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWT.RefreshSecret), nil
	})

	assert.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*RefreshTokenClaims)
	assert.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
}

func TestValidateAccessToken(t *testing.T) {
	setupTokenConfig()
	userID := primitive.NewObjectID().Hex()
	tokenRes, _ := GenAccessToken(userID)

	t.Run("有效令牌", func(t *testing.T) {
		valid, err := ValidateAccessToken(tokenRes.Token)
		assert.NoError(t, err)
		assert.True(t, valid)
	})

	t.Run("無效令牌", func(t *testing.T) {
		valid, err := ValidateAccessToken("invalid" + tokenRes.Token)
		assert.Error(t, err)
		assert.False(t, valid)
	})

	t.Run("過期令牌", func(t *testing.T) {
		// 建立過期的令牌
		config.AppConfig.JWT.AccessExpireHours = -1
		expiredTokenRes, _ := GenAccessToken(userID)
		config.AppConfig.JWT.AccessExpireHours = 1 // 重置

		valid, err := ValidateAccessToken(expiredTokenRes.Token)
		assert.Error(t, err)
		assert.False(t, valid)
		assert.Contains(t, err.Error(), "token is expired")
	})
}

func TestGetUserIDFromToken(t *testing.T) {
	setupTokenConfig()
	userID := primitive.NewObjectID().Hex()
	tokenRes, _ := GenAccessToken(userID)

	extractedID, objectID, err := GetUserFromToken(tokenRes.Token)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedID)
	assert.NotEqual(t, primitive.NilObjectID, objectID)

	_, _, err = GetUserFromToken("invalidtoken")
	assert.Error(t, err)
}
