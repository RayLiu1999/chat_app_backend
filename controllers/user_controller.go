package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
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

// 註冊
func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	appError := uc.userService.RegisterUser(user)
	if appError != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: appError.Code, Message: appError.Message, Displayable: appError.Displayable})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "用戶創建成功"})
}

// 登入
func (uc *UserController) Login(c *gin.Context) {
	// 綁定請求參數
	var loginUser models.User
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 調用服務層處理登入邏輯
	response, appErr := uc.userService.Login(loginUser)
	if appErr != nil {
		statusCode := http.StatusInternalServerError
		if appErr.Code == utils.ErrLoginFailed {
			statusCode = http.StatusUnauthorized
		}
		utils.ErrorResponse(c, statusCode, utils.MessageOptions{Code: appErr.Code, Displayable: appErr.Displayable})
		return
	}

	// 將 refresh token 寫入 cookie
	c.SetCookie("refresh_token", response.RefreshToken, 3600*72, "/", "localhost", false, true)

	// 返回 access token 給客戶端
	utils.SuccessResponse(c, gin.H{"access_token": response.AccessToken}, utils.MessageOptions{Message: "登入成功"})
}

// 登出
func (uc *UserController) Logout(c *gin.Context) {
	appErr := uc.userService.Logout(c)
	if appErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: appErr.Code, Displayable: appErr.Displayable})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "登出成功"})
}

// 刷新 access token
func (uc *UserController) RefreshToken(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 調用服務層的 RefreshToken 方法
	accessToken, appErr := uc.userService.RefreshToken(token)
	if appErr != nil {
		statusCode := http.StatusInternalServerError
		if appErr.Code == utils.ErrInvalidToken {
			statusCode = http.StatusUnauthorized
		}
		utils.ErrorResponse(c, statusCode, utils.MessageOptions{Code: appErr.Code, Displayable: appErr.Displayable})
		return
	}

	utils.SuccessResponse(c, gin.H{"access_token": accessToken}, utils.MessageOptions{Message: "令牌刷新成功"})
}

// 取得用戶資訊
func (uc *UserController) GetUser(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	userResponse, err := uc.userService.GetUserResponseById(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound})
		return
	}

	utils.SuccessResponse(c, userResponse, utils.MessageOptions{Message: "使用者資訊獲取成功"})
}

// CheckUserOnlineStatus 檢查特定用戶是否在線
func (uc *UserController) CheckUserOnlineStatus(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Message: "用戶 ID 不能為空"})
		return
	}

	isOnline := uc.userService.IsUserOnlineByWebSocket(userID)
	utils.SuccessResponse(c, map[string]interface{}{
		"user_id":   userID,
		"is_online": isOnline,
	}, utils.MessageOptions{Message: "用戶在線狀態檢查完成"})
}
