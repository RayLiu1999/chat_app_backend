package controllers

import (
	"context"
	"net/http"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	config       *config.Config
	mongoConnect *mongo.Database
	userService  services.UserServiceInterface
}

func NewUserController(cfg *config.Config, mongodb *mongo.Database, userService services.UserServiceInterface) *UserController {
	return &UserController{
		config:       cfg,
		mongoConnect: mongodb,
		userService:  userService,
	}
}

type TokenResponse struct {
	Token     string
	ExpiresAt int64
}

type APIUser struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
	NickName string `json:"nick_name" bson:"nick_name"`
	Picture  string `json:"picture" bson:"picture"`
}

// 註冊
func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrInvalidParams, "請求參數錯誤")
		return
	}

	// 檢查用戶名是否已存在
	collection := uc.mongoConnect.Collection("users")
	var existingUser models.User
	err := collection.FindOne(context.Background(), bson.M{"username": user.Username, "email": user.Email}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, utils.ErrUsernameExists, "該用戶名已被使用")
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	utils.SuccessResponse(c, nil, "用戶創建成功", 0)
}

// 登入
func (uc *UserController) Login(c *gin.Context) {
	var loginUser models.User
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrInvalidParams, "請求參數錯誤")
		return
	}

	collection := uc.mongoConnect.Collection("users")
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": loginUser.Email}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrLoginFailed, "登入失敗，請檢查用戶名和密碼")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrLoginFailed, "登入失敗，請檢查用戶名和密碼")
		return
	}

	// Generate JWT tokens
	refreshTokenResponse, err := GenRefreshToken(user.ID.Hex())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	// 將 refresh token 寫入 資料庫
	refreshTokenCollection := uc.mongoConnect.Collection("refresh_tokens")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	// 將 refresh token 寫入 cookie
	c.SetCookie("refresh_token", refreshTokenResponse.Token, 3600*72, "/", "localhost", false, true)

	// Generate JWT tokens
	accessTokenResponse, err := GenAccessToken(user.ID.Hex())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	// 返回 access token 給客戶端
	utils.SuccessResponse(c, gin.H{"access_token": accessTokenResponse.Token}, "登入成功", 0)
}

// 登出
func (uc *UserController) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	utils.SuccessResponse(c, nil, "登出成功", 0)
}

// 刷新 access token
func (uc *UserController) Refresh(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "未提供刷新令牌")
		return
	}

	// 資料庫檢查 refresh token 是否有效
	refreshTokenCollection := uc.mongoConnect.Collection("refresh_tokens")
	var refreshTokenDoc models.RefreshToken
	err = refreshTokenCollection.FindOne(context.Background(), bson.M{"token": token, "expires_at": bson.M{"$gt": time.Now().Unix()}}).Decode(&refreshTokenDoc)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrInvalidToken, "無效的刷新令牌")
		return
	}

	// Generate new access
	accessTokenResponse, err := GenAccessToken(refreshTokenDoc.UserID.Hex())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	utils.SuccessResponse(c, gin.H{"access_token": accessTokenResponse.Token}, "令牌刷新成功", 0)
}

// 取得用戶資訊
func (uc *UserController) GetUser(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "未授權的請求")
		return
	}

	collection := uc.mongoConnect.Collection("users")

	var apiUser APIUser
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&apiUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.ErrUserNotFound, "使用者不存在")
		return
	}

	utils.SuccessResponse(c, apiUser, "使用者資訊獲取成功", 0)
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
