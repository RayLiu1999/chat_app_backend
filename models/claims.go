package models

import "github.com/dgrijalva/jwt-go"

// AccessTokenClaims 定義了 access token 中的聲明
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// RefreshTokenClaims 定義了 refresh token 中的聲明
type RefreshTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}
