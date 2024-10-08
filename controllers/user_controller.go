package controllers

import (
	"context"
	"net/http"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	*BaseController
}

type TokenResponse struct {
	Token     string
	ExpiresAt int64
}

// 註冊
func (bc *BaseController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 檢查用戶名是否已存在
	collection := bc.MongoConnect.Collection("users")
	var existingUser models.User
	err := collection.FindOne(context.Background(), bson.M{"username": user.Username, "email": user.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists!"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user!"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

// 登入
func (bc *BaseController) Login(c *gin.Context) {
	var loginUser models.User
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := bc.MongoConnect.Collection("users")
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": loginUser.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT tokens
	refreshTokenResponse, err := GenRefreshToken(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// 將 refresh token 寫入 資料庫
	refreshTokenCollection := bc.MongoConnect.Collection("refresh_tokens")
	var refreshTokenDoc = models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshTokenResponse.Token,
		ExpiresAt: refreshTokenResponse.ExpiresAt,
		Revoked:   false,
		CreatedAt: time.Now().Unix(),
		UpdateAt:  time.Now().Unix(),
	}
	_, err = refreshTokenCollection.InsertOne(context.Background(), refreshTokenDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// 將 refresh token 寫入 cookie
	c.SetCookie("refresh_token", refreshTokenResponse.Token, 3600*72, "/", "localhost", false, true)

	// Generate JWT tokens
	accessTokenResponse, err := GenAccessToken(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// 返回 access token 給客戶端
	c.JSON(http.StatusOK, gin.H{"access_token": accessTokenResponse.Token})
}

// 登出
func (bc *BaseController) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// 刷新 access token
func (bc *BaseController) Refresh(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token provided"})
		return
	}

	// 資料庫檢查 refresh token 是否有效
	refreshTokenCollection := bc.MongoConnect.Collection("refresh_tokens")
	var refreshTokenDoc models.RefreshToken
	err = refreshTokenCollection.FindOne(context.Background(), bson.M{"token": token, "expires_at": bson.M{"$gt": time.Now().Unix()}}).Decode(&refreshTokenDoc)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Generate new access
	accessTokenResponse, err := GenAccessToken(refreshTokenDoc.UserID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessTokenResponse.Token})
}

// 取得用戶資訊
func (bc *BaseController) GetUser(c *gin.Context) {
	var user models.User
	var userID string
	var err error

	accessToken, _ := services.GetAccessTokenByHeader(c)
	userID, err = services.GetUserFromToken(accessToken)
	collection := bc.MongoConnect.Collection("users")

	err = collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

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
