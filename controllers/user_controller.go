package controllers

import (
	"context"
	"net/http"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"

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

type APIUser struct {
	ID       string `json:"id" bson:"_id"`
	Username string `json:"username" bson:"username"`
	Email    string `json:"email" bson:"email"`
	Nickname string `json:"nickname" bson:"nickname"`
	Picture  string `json:"picture" bson:"picture"`
}

// 註冊
func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 檢查用戶名是否已存在
	collection := uc.mongoConnect.Collection("users")
	existingUser := models.User{}
	err := collection.FindOne(context.Background(), bson.M{"username": user.Username}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, utils.MessageOptions{Code: utils.ErrUsernameExists, Message: "使用者名稱已被使用", Displayable: true})
		return
	}

	existingUser = models.User{}
	err = collection.FindOne(context.Background(), bson.M{"email": user.Email}).Decode(&existingUser)
	if err == nil {
		utils.ErrorResponse(c, http.StatusConflict, utils.MessageOptions{Code: utils.ErrEmailExists, Message: "電子郵件已被使用", Displayable: true})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "用戶創建成功"})
}

// 登入
func (uc *UserController) Login(c *gin.Context) {
	var loginUser models.User
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	collection := uc.mongoConnect.Collection("users")
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": loginUser.Email}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrLoginFailed, Displayable: true})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginUser.Password))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrLoginFailed, Displayable: true})
		return
	}

	// Generate JWT tokens
	refreshTokenResponse, err := utils.GenRefreshToken(user.ID.Hex())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
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
		UpdatedAt: time.Now().Unix(),
	}
	_, err = refreshTokenCollection.InsertOne(context.Background(), refreshTokenDoc)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	// 將 refresh token 寫入 cookie
	c.SetCookie("refresh_token", refreshTokenResponse.Token, 3600*72, "/", "localhost", false, true)

	// Generate JWT tokens
	accessTokenResponse, err := utils.GenAccessToken(user.ID.Hex())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	// 返回 access token 給客戶端
	utils.SuccessResponse(c, gin.H{"access_token": accessTokenResponse.Token}, utils.MessageOptions{Message: "登入成功"})
}

// 登出
func (uc *UserController) Logout(c *gin.Context) {
	c.SetCookie("refresh_token", "", -1, "/", "localhost", false, true)
	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "登出成功"})
}

// 刷新 access token
func (uc *UserController) Refresh(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 資料庫檢查 refresh token 是否有效
	refreshTokenCollection := uc.mongoConnect.Collection("refresh_tokens")
	var refreshTokenDoc models.RefreshToken
	err = refreshTokenCollection.FindOne(context.Background(), bson.M{"token": token, "expires_at": bson.M{"$gt": time.Now().Unix()}}).Decode(&refreshTokenDoc)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrInvalidToken})
		return
	}

	// Generate new access
	accessTokenResponse, err := utils.GenAccessToken(refreshTokenDoc.UserID.Hex())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	utils.SuccessResponse(c, gin.H{"access_token": accessTokenResponse.Token}, utils.MessageOptions{Message: "令牌刷新成功"})
}

// 取得用戶資訊
func (uc *UserController) GetUser(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	collection := uc.mongoConnect.Collection("users")

	var apiUser APIUser
	err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&apiUser)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound})
		return
	}

	utils.SuccessResponse(c, apiUser, utils.MessageOptions{Message: "使用者資訊獲取成功"})
}
