package mocks

import (
	"mime/multipart"

	"chat_app_backend/app/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// UserService 是 services.UserService 介面的 mock 實現
type UserService struct {
	mock.Mock
}

// GetUserResponseById 根據ID獲取用戶信息
func (m *UserService) GetUserResponseById(userID string) (*models.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserResponse), args.Error(1)
}

// GetUserByUsername 根據用戶名獲取用戶信息(測試用)
func (m *UserService) GetUserByUsername(username string) (*models.User, *models.MessageOptions) {
	args := m.Called(username)
	var user *models.User
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return user, msgOpts
}

// RegisterUser 註冊新用戶
func (m *UserService) RegisterUser(user models.User) *models.MessageOptions {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// Login 處理用戶登入
func (m *UserService) Login(loginUser models.User) (*models.LoginResponse, *models.MessageOptions) {
	args := m.Called(loginUser)
	var loginResp *models.LoginResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		loginResp = args.Get(0).(*models.LoginResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return loginResp, msgOpts
}

// Logout 處理用戶登出
func (m *UserService) Logout(c *gin.Context) *models.MessageOptions {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// RefreshToken 刷新令牌
func (m *UserService) RefreshToken(refreshToken string) (*models.RefreshTokenResponse, *models.MessageOptions) {
	args := m.Called(refreshToken)
	var resp *models.RefreshTokenResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.RefreshTokenResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// SetUserOnline 設置用戶為在線狀態
func (m *UserService) SetUserOnline(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// SetUserOffline 設置用戶為離線狀態
func (m *UserService) SetUserOffline(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// UpdateUserActivity 更新用戶活動時間
func (m *UserService) UpdateUserActivity(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// CheckAndSetOfflineUsers 檢查並設置離線用戶
func (m *UserService) CheckAndSetOfflineUsers(offlineThresholdMinutes int) error {
	args := m.Called(offlineThresholdMinutes)
	return args.Error(0)
}

// GetUserProfile 獲取用戶個人資料
func (m *UserService) GetUserProfile(userID string) (*models.UserProfileResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserProfileResponse), args.Error(1)
}

// UpdateUserProfile 更新用戶基本資料
func (m *UserService) UpdateUserProfile(userID string, updates map[string]any) error {
	args := m.Called(userID, updates)
	return args.Error(0)
}

// UploadUserImage 上傳用戶頭像或橫幅
func (m *UserService) UploadUserImage(userID string, file multipart.File, header *multipart.FileHeader, imageType string) (*models.UserImageResponse, error) {
	args := m.Called(userID, file, header, imageType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserImageResponse), args.Error(1)
}

// DeleteUserAvatar 刪除用戶頭像
func (m *UserService) DeleteUserAvatar(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// DeleteUserBanner 刪除用戶橫幅
func (m *UserService) DeleteUserBanner(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// UpdateUserPassword 更新用戶密碼
func (m *UserService) UpdateUserPassword(userID string, newPassword string) error {
	args := m.Called(userID, newPassword)
	return args.Error(0)
}

// GetTwoFactorStatus 獲取兩步驟驗證狀態
func (m *UserService) GetTwoFactorStatus(userID string) (*models.TwoFactorStatusResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.TwoFactorStatusResponse), args.Error(1)
}

// UpdateTwoFactorStatus 啟用/停用兩步驟驗證
func (m *UserService) UpdateTwoFactorStatus(userID string, enabled bool) error {
	args := m.Called(userID, enabled)
	return args.Error(0)
}

// DeactivateAccount 停用帳號
func (m *UserService) DeactivateAccount(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// DeleteAccount 刪除帳號
func (m *UserService) DeleteAccount(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}
