package controllers

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	config        *config.Config
	mongoConnect  *mongo.Database
	userService   services.UserService
	clientManager services.ClientManager
}

func NewUserController(cfg *config.Config, mongodb *mongo.Database, userService services.UserService, clientManager services.ClientManager) *UserController {
	return &UserController{
		config:        cfg,
		mongoConnect:  mongodb,
		userService:   userService,
		clientManager: clientManager,
	}
}

// 註冊
func (uc *UserController) Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "註冊失敗",
			Details: err.Error(),
		})
		return
	}

	appError := uc.userService.RegisterUser(user)
	if appError != nil {
		// 根據錯誤類型決定 HTTP 狀態碼
		statusCode := http.StatusInternalServerError
		if appError.Code == models.ErrUsernameExists || appError.Code == models.ErrEmailExists {
			statusCode = http.StatusBadRequest
		}
		ErrorResponse(c, statusCode, models.MessageOptions{
			Code:    appError.Code,
			Message: appError.Message,
		})
		return
	}

	SuccessResponse(c, nil, "用戶創建成功")
}

// 登入
func (uc *UserController) Login(c *gin.Context) {
	// 綁定請求參數
	var loginUser models.User
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "登入失敗",
			Details: err.Error(),
		})
		return
	}

	// 調用服務層處理登入邏輯
	response, appErr := uc.userService.Login(loginUser)
	if appErr != nil {
		statusCode := http.StatusInternalServerError
		if appErr.Code == models.ErrLoginFailed {
			statusCode = http.StatusUnauthorized
		}
		ErrorResponse(c, statusCode, models.MessageOptions{
			Code:    appErr.Code,
			Message: "登入失敗",
			Details: appErr.Message,
		})
		return
	}

	// 將 refresh token 寫入 cookie
	utils.SetCookie(c, uc.config, "refresh_token", response.RefreshToken, uc.config.JWT.RefreshExpireHours*3600, true)

	// 將 CSRF token 寫入 cookie（不設定 HttpOnly，讓前端可以讀取）
	utils.SetCookie(c, uc.config, "csrf_token", response.CSRFToken, uc.config.JWT.RefreshExpireHours*3600, false)

	// 返回 access token 給客戶端
	SuccessResponse(c, gin.H{"access_token": response.AccessToken}, "登入成功")
}

// 登出
func (uc *UserController) Logout(c *gin.Context) {
	appErr := uc.userService.Logout(c)
	if appErr != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{Code: appErr.Code})
		return
	}

	SuccessResponse(c, nil, "登出成功")
}

// GetUserByUsername 根據用戶名獲取用戶信息（測試用，不需要認證）
func (uc *UserController) GetUserByUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "用戶名不能為空",
		})
		return
	}

	user, appErr := uc.userService.GetUserByUsername(username)
	if appErr != nil {
		statusCode := http.StatusNotFound
		if appErr.Code != models.ErrUserNotFound {
			statusCode = http.StatusInternalServerError
		}
		ErrorResponse(c, statusCode, models.MessageOptions{
			Code:    appErr.Code,
			Message: "獲取用戶信息失敗",
			Details: appErr.Message,
		})
		return
	}

	// 移除敏感信息
	user.Password = ""

	SuccessResponse(c, gin.H{
		"id":             user.ID.Hex(),
		"username":       user.Username,
		"email":          user.Email,
		"nickname":       user.Nickname,
		"picture_id":     user.PictureID.Hex(),
		"banner_id":      user.BannerID.Hex(),
		"status":         user.Status,
		"bio":            user.Bio,
		"is_online":      user.IsOnline,
		"last_active_at": user.LastActiveAt,
		"created_at":     user.CreatedAt,
		"is_active":      user.IsActive,
	}, "獲取用戶信息成功")
}

// 刷新 access token
func (uc *UserController) RefreshToken(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 調用服務層的 RefreshToken 方法
	response, appErr := uc.userService.RefreshToken(token)
	if appErr != nil {
		statusCode := http.StatusInternalServerError
		if appErr.Code == models.ErrInvalidToken {
			statusCode = http.StatusUnauthorized
		}
		ErrorResponse(c, statusCode, models.MessageOptions{
			Code:    appErr.Code,
			Message: "令牌刷新失敗",
			Details: appErr.Message,
		})
		return
	}

	// 將新的 csrf token 寫入 cookie
	utils.SetCookie(c, uc.config, "csrf_token", response.CSRFToken, uc.config.JWT.RefreshExpireHours*3600, false)

	SuccessResponse(c, gin.H{"access_token": response.AccessToken}, "令牌刷新成功")
}

// 取得用戶資訊
func (uc *UserController) GetUser(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	userResponse, err := uc.userService.GetUserResponseById(userID)
	if err != nil {
		ErrorResponse(c, http.StatusNotFound, models.MessageOptions{
			Code:    models.ErrUserNotFound,
			Message: "使用者資訊獲取失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, userResponse, "使用者資訊獲取成功")
}

// CheckUserOnlineStatus 檢查特定用戶是否在線
func (uc *UserController) CheckUserOnlineStatus(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "用戶 ID 不能為空",
		})
		return
	}

	isOnline := uc.clientManager.IsUserOnline(userID)
	SuccessResponse(c, map[string]any{
		"user_id":   userID,
		"is_online": isOnline,
	}, "用戶在線狀態檢查完成")
}

// GetUserProfile 獲取用戶個人資料
func (uc *UserController) GetUserProfile(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	profile, err := uc.userService.GetUserProfile(userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取個人資料失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, profile, "獲取個人資料成功")
}

// UpdateUserProfile 更新用戶基本資料
func (uc *UserController) UpdateUserProfile(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	var updates map[string]any
	if err := c.ShouldBindJSON(&updates); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "更新個人資料失敗",
			Details: err.Error(),
		})
		return
	}

	err = uc.userService.UpdateUserProfile(userID, updates)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "更新個人資料失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "更新個人資料成功")
}

// UploadUserImage 上傳用戶頭像或橫幅
func (uc *UserController) UploadUserImage(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無法獲取上傳的檔案",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// 獲取圖片類型
	imageType := c.PostForm("type")
	if imageType != "avatar" && imageType != "banner" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "圖片類型必須是 avatar 或 banner",
		})
		return
	}

	result, err := uc.userService.UploadUserImage(userID, file, header, imageType)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "圖片上傳失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, result, "圖片上傳成功")
}

// DeleteUserAvatar 刪除用戶頭像
func (uc *UserController) DeleteUserAvatar(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	err = uc.userService.DeleteUserAvatar(userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "頭像刪除失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "頭像刪除成功")
}

// DeleteUserBanner 刪除用戶橫幅
func (uc *UserController) DeleteUserBanner(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	err = uc.userService.DeleteUserBanner(userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "橫幅刪除失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "橫幅刪除成功")
}

// UpdateUserPassword 更新用戶密碼
func (uc *UserController) UpdateUserPassword(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "密碼更新失敗",
			Details: err.Error(),
		})
		return
	}

	err = uc.userService.UpdateUserPassword(userID, req.NewPassword)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "密碼更新失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "密碼更新成功")
}

// GetTwoFactorStatus 獲取兩步驟驗證狀態
func (uc *UserController) GetTwoFactorStatus(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	status, err := uc.userService.GetTwoFactorStatus(userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取兩步驟驗證狀態失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, status, "獲取兩步驟驗證狀態成功")
}

// UpdateTwoFactorStatus 啟用/停用兩步驟驗證
func (uc *UserController) UpdateTwoFactorStatus(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "兩步驟驗證狀態更新失敗",
			Details: err.Error(),
		})
		return
	}

	err = uc.userService.UpdateTwoFactorStatus(userID, req.Enabled)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "兩步驟驗證狀態更新失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "兩步驟驗證狀態更新成功")
}

// DeactivateAccount 停用帳號
func (uc *UserController) DeactivateAccount(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	err = uc.userService.DeactivateAccount(userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "帳號停用失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "帳號已停用")
}

// DeleteAccount 刪除帳號
func (uc *UserController) DeleteAccount(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	err = uc.userService.DeleteAccount(userID)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "帳號刪除失敗",
			Details: err.Error(),
		})
		return
	}

	SuccessResponse(c, nil, "帳號已刪除")
}
