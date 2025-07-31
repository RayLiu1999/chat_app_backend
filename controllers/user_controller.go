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

// GetUserProfile 獲取用戶個人資料
func (uc *UserController) GetUserProfile(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	profile, err := uc.userService.GetUserProfile(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, profile, utils.MessageOptions{Message: "獲取個人資料成功"})
}

// UpdateUserProfile 更新用戶基本資料
func (uc *UserController) UpdateUserProfile(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	err = uc.userService.UpdateUserProfile(userID, updates)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "更新個人資料成功"})
}

// UploadUserImage 上傳用戶頭像或橫幅
func (uc *UserController) UploadUserImage(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: "無法獲取上傳的檔案"})
		return
	}
	defer file.Close()

	// 獲取圖片類型
	imageType := c.PostForm("type")
	if imageType != "avatar" && imageType != "banner" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: "圖片類型必須是 avatar 或 banner"})
		return
	}

	result, err := uc.userService.UploadUserImage(userID, file, header, imageType)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, result, utils.MessageOptions{Message: "圖片上傳成功"})
}

// DeleteUserAvatar 刪除用戶頭像
func (uc *UserController) DeleteUserAvatar(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	err = uc.userService.DeleteUserAvatar(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "頭像刪除成功"})
}

// DeleteUserBanner 刪除用戶橫幅
func (uc *UserController) DeleteUserBanner(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	err = uc.userService.DeleteUserBanner(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "橫幅刪除成功"})
}

// UpdateUserPassword 更新用戶密碼
func (uc *UserController) UpdateUserPassword(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	err = uc.userService.UpdateUserPassword(userID, req.NewPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "密碼更新成功"})
}

// GetTwoFactorStatus 獲取兩步驟驗證狀態
func (uc *UserController) GetTwoFactorStatus(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	status, err := uc.userService.GetTwoFactorStatus(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, status, utils.MessageOptions{Message: "獲取兩步驟驗證狀態成功"})
}

// UpdateTwoFactorStatus 啟用/停用兩步驟驗證
func (uc *UserController) UpdateTwoFactorStatus(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	err = uc.userService.UpdateTwoFactorStatus(userID, req.Enabled)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "兩步驟驗證狀態更新成功"})
}

// DeactivateAccount 停用帳號
func (uc *UserController) DeactivateAccount(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	err = uc.userService.DeactivateAccount(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "帳號已停用"})
}

// DeleteAccount 刪除帳號
func (uc *UserController) DeleteAccount(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	err = uc.userService.DeleteAccount(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "帳號已刪除"})
}
