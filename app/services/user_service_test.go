package services

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestNewUserService 測試創建 UserService
func TestNewUserService(t *testing.T) {
	service := NewUserService(nil, nil, nil, nil)

	assert.NotNil(t, service, "服務應該被成功創建")
	assert.IsType(t, &userService{}, service, "服務應該是 *userService 類型")
}

// TestGetUserPictureURL 測試獲取用戶頭像 URL
func TestGetUserPictureURL(t *testing.T) {
	t.Run("用戶沒有頭像", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NilObjectID,
			},
			PictureID: primitive.NilObjectID,
		}

		url := service.getUserPictureURL(user)
		assert.Empty(t, url, "沒有頭像時應該返回空字串")
	})

	t.Run("FileUploadService 為 nil", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NewObjectID(),
			},
			PictureID: primitive.NewObjectID(),
		}

		url := service.getUserPictureURL(user)
		assert.Empty(t, url, "FileUploadService 為 nil 時應該返回空字串")
	})
}

// TestGetUserBannerURL 測試獲取用戶橫幅 URL
func TestGetUserBannerURL(t *testing.T) {
	t.Run("用戶沒有橫幅", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NilObjectID,
			},
			BannerID: primitive.NilObjectID,
		}

		url := service.getUserBannerURL(user)
		assert.Empty(t, url, "沒有橫幅時應該返回空字串")
	})

	t.Run("FileUploadService 為 nil", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NewObjectID(),
			},
			BannerID: primitive.NewObjectID(),
		}

		url := service.getUserBannerURL(user)
		assert.Empty(t, url, "FileUploadService 為 nil 時應該返回空字串")
	})
}

// TestGetUserPictureURL_WithFileService 測試帶文件服務的頭像獲取
func TestGetUserPictureURL_WithFileService(t *testing.T) {
	t.Run("成功獲取頭像 URL", func(t *testing.T) {
		pictureID := primitive.NewObjectID()
		expectedURL := "https://example.com/avatar.jpg"

		fileService := new(mocks.FileUploadService)
		fileService.On("GetFileURLByID", pictureID.Hex()).Return(expectedURL, (*models.MessageOptions)(nil))

		service := NewUserService(nil, nil, nil, fileService)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NewObjectID(),
			},
			PictureID: pictureID,
		}

		url := service.getUserPictureURL(user)
		assert.Equal(t, expectedURL, url, "應該返回正確的頭像 URL")
	})

	t.Run("獲取頭像 URL 失敗", func(t *testing.T) {
		fileService := new(mocks.FileUploadService)
		fileService.On("GetFileURLByID", mock.Anything).Return("", &models.MessageOptions{Code: models.ErrInternalServer})

		service := NewUserService(nil, nil, nil, fileService)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NewObjectID(),
			},
			PictureID: primitive.NewObjectID(),
		}

		url := service.getUserPictureURL(user)
		assert.Empty(t, url, "獲取失敗時應該返回空字串")
	})
}

// TestGetUserBannerURL_WithFileService 測試帶文件服務的橫幅獲取
func TestGetUserBannerURL_WithFileService(t *testing.T) {
	t.Run("成功獲取橫幅 URL", func(t *testing.T) {
		bannerID := primitive.NewObjectID()
		expectedURL := "https://example.com/banner.jpg"

		fileService := new(mocks.FileUploadService)
		fileService.On("GetFileURLByID", bannerID.Hex()).Return(expectedURL, (*models.MessageOptions)(nil))

		service := NewUserService(nil, nil, nil, fileService)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NewObjectID(),
			},
			BannerID: bannerID,
		}

		url := service.getUserBannerURL(user)
		assert.Equal(t, expectedURL, url, "應該返回正確的橫幅 URL")
	})

	t.Run("獲取橫幅 URL 失敗", func(t *testing.T) {
		fileService := new(mocks.FileUploadService)
		fileService.On("GetFileURLByID", mock.Anything).Return("", &models.MessageOptions{Code: models.ErrInternalServer})

		service := NewUserService(nil, nil, nil, fileService)

		user := &models.User{
			BaseModel: providers.BaseModel{
				ID: primitive.NewObjectID(),
			},
			BannerID: primitive.NewObjectID(),
		}

		url := service.getUserBannerURL(user)
		assert.Empty(t, url, "獲取失敗時應該返回空字串")
	})
}

// TestUserService_ServiceInitialization 測試服務初始化
func TestUserService_ServiceInitialization(t *testing.T) {
	t.Run("使用 nil 依賴初始化", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)
		assert.NotNil(t, service)
	})

	t.Run("使用完整依賴初始化", func(t *testing.T) {
		fileService := new(mocks.FileUploadService)
		service := NewUserService(nil, nil, nil, fileService)
		assert.NotNil(t, service)
	})
}

// TestUserService_NilSafety 測試 nil 安全性
func TestUserService_NilSafety(t *testing.T) {
	t.Run("getUserPictureURL 處理 nil fileService", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)

		user := &models.User{
			PictureID: primitive.NewObjectID(),
		}

		// 不應該 panic
		assert.NotPanics(t, func() {
			url := service.getUserPictureURL(user)
			assert.Empty(t, url)
		})
	})

	t.Run("getUserBannerURL 處理 nil fileService", func(t *testing.T) {
		service := NewUserService(nil, nil, nil, nil)

		user := &models.User{
			BannerID: primitive.NewObjectID(),
		}

		// 不應該 panic
		assert.NotPanics(t, func() {
			url := service.getUserBannerURL(user)
			assert.Empty(t, url)
		})
	})

	t.Run("getUserPictureURL 處理零值 ObjectID", func(t *testing.T) {
		fileService := new(mocks.FileUploadService)
		service := NewUserService(nil, nil, nil, fileService)

		user := &models.User{
			PictureID: primitive.NilObjectID,
		}

		url := service.getUserPictureURL(user)
		assert.Empty(t, url, "零值 ObjectID 應該返回空字串")
	})

	t.Run("getUserBannerURL 處理零值 ObjectID", func(t *testing.T) {
		fileService := new(mocks.FileUploadService)
		service := NewUserService(nil, nil, nil, fileService)

		user := &models.User{
			BannerID: primitive.NilObjectID,
		}

		url := service.getUserBannerURL(user)
		assert.Empty(t, url, "零值 ObjectID 應該返回空字串")
	})
}

// ===== 新增測試 =====

// testUserRepository 模擬 UserRepository（user service專用）
type testUserRepository struct {
	checkUsernameExistsFunc func(username string) (bool, error)
	checkEmailExistsFunc    func(email string) (bool, error)
	createUserFunc          func(user models.User) error
	getUserByIdFunc         func(userID string) (*models.User, error)
	updateUserFunc          func(userID string, updates map[string]any) error
	deleteUserFunc          func(userID string) error
	updateOnlineStatusFunc  func(userID string, isOnline bool) error
	updateLastActiveFunc    func(userID string, timestamp int64) error
}

func (m *testUserRepository) GetUserById(userID string) (*models.User, error) {
	if m.getUserByIdFunc != nil {
		return m.getUserByIdFunc(userID)
	}
	return nil, fmt.Errorf("user not found")
}

func (m *testUserRepository) GetUserByUsername(username string) (*models.User, error) {
	return nil, fmt.Errorf("user not found")
}

func (m *testUserRepository) GetUserListByIds(userIds []string) ([]models.User, error) {
	return nil, nil
}

func (m *testUserRepository) CheckUsernameExists(username string) (bool, error) {
	if m.checkUsernameExistsFunc != nil {
		return m.checkUsernameExistsFunc(username)
	}
	return false, nil
}

func (m *testUserRepository) CheckEmailExists(email string) (bool, error) {
	if m.checkEmailExistsFunc != nil {
		return m.checkEmailExistsFunc(email)
	}
	return false, nil
}

func (m *testUserRepository) CreateUser(user models.User) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(user)
	}
	return nil
}

func (m *testUserRepository) UpdateUserOnlineStatus(userID string, isOnline bool) error {
	if m.updateOnlineStatusFunc != nil {
		return m.updateOnlineStatusFunc(userID, isOnline)
	}
	return nil
}

func (m *testUserRepository) UpdateUserLastActiveTime(userID string, timestamp int64) error {
	if m.updateLastActiveFunc != nil {
		return m.updateLastActiveFunc(userID, timestamp)
	}
	return nil
}

func (m *testUserRepository) UpdateUser(userID string, updates map[string]any) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(userID, updates)
	}
	return nil
}

func (m *testUserRepository) DeleteUser(userID string) error {
	if m.deleteUserFunc != nil {
		return m.deleteUserFunc(userID)
	}
	return nil
}

// TestRegisterUser 測試用戶註冊
func TestRegisterUser(t *testing.T) {
	t.Run("成功註冊新用戶", func(t *testing.T) {
		mockRepo := &testUserRepository{
			checkUsernameExistsFunc: func(username string) (bool, error) {
				return false, nil
			},
			checkEmailExistsFunc: func(email string) (bool, error) {
				return false, nil
			},
			createUserFunc: func(user models.User) error {
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		user := models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		msgOpt := service.RegisterUser(user)

		assert.Nil(t, msgOpt)
	})

	t.Run("用戶名已存在", func(t *testing.T) {
		mockRepo := &testUserRepository{
			checkUsernameExistsFunc: func(username string) (bool, error) {
				return true, nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		user := models.User{
			Username: "existinguser",
			Email:    "test@example.com",
			Password: "password123",
		}

		msgOpt := service.RegisterUser(user)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUsernameExists, msgOpt.Code)
	})

	t.Run("電子郵件已存在", func(t *testing.T) {
		mockRepo := &testUserRepository{
			checkUsernameExistsFunc: func(username string) (bool, error) {
				return false, nil
			},
			checkEmailExistsFunc: func(email string) (bool, error) {
				return true, nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		user := models.User{
			Username: "testuser",
			Email:    "existing@example.com",
			Password: "password123",
		}

		msgOpt := service.RegisterUser(user)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrEmailExists, msgOpt.Code)
	})

	t.Run("檢查用戶名時發生錯誤", func(t *testing.T) {
		mockRepo := &testUserRepository{
			checkUsernameExistsFunc: func(username string) (bool, error) {
				return false, fmt.Errorf("database error")
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		user := models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		msgOpt := service.RegisterUser(user)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
	})
}

// TestGetUserResponseById 測試獲取用戶信息
func TestGetUserResponseById(t *testing.T) {
	t.Run("成功獲取用戶信息", func(t *testing.T) {
		userID := primitive.NewObjectID()
		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return &models.User{
					BaseModel: providers.BaseModel{ID: userID},
					Username:  "testuser",
					Email:     "test@example.com",
					Nickname:  "Test User",
				}, nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		response, err := service.GetUserResponseById(userID.Hex())

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "testuser", response.Username)
		assert.Equal(t, "Test User", response.Nickname)
	})

	t.Run("用戶不存在", func(t *testing.T) {
		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return nil, fmt.Errorf("user not found")
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		response, err := service.GetUserResponseById(primitive.NewObjectID().Hex())

		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

// TestSetUserOnline 測試設置用戶在線狀態
func TestSetUserOnline(t *testing.T) {
	t.Run("成功設置用戶在線", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			updateOnlineStatusFunc: func(id string, isOnline bool) error {
				called = true
				assert.Equal(t, userID, id)
				assert.True(t, isOnline)
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.SetUserOnline(userID)

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("設置在線狀態失敗", func(t *testing.T) {
		mockRepo := &testUserRepository{
			updateOnlineStatusFunc: func(id string, isOnline bool) error {
				return fmt.Errorf("database error")
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.SetUserOnline(primitive.NewObjectID().Hex())

		assert.Error(t, err)
	})
}

// TestSetUserOffline 測試設置用戶離線狀態
func TestSetUserOffline(t *testing.T) {
	t.Run("成功設置用戶離線", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			updateOnlineStatusFunc: func(id string, isOnline bool) error {
				called = true
				assert.Equal(t, userID, id)
				assert.False(t, isOnline)
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.SetUserOffline(userID)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

// TestUpdateUserActivity 測試更新用戶活動時間
func TestUpdateUserActivity(t *testing.T) {
	t.Run("成功更新用戶活動時間", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			updateLastActiveFunc: func(id string, timestamp int64) error {
				called = true
				assert.Equal(t, userID, id)
				assert.Greater(t, timestamp, int64(0))
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.UpdateUserActivity(userID)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

// TestGetUserProfile 測試獲取用戶個人資料
func TestGetUserProfile(t *testing.T) {
	t.Run("成功獲取用戶個人資料", func(t *testing.T) {
		userID := primitive.NewObjectID()
		pictureID := primitive.NewObjectID()

		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return &models.User{
					BaseModel: providers.BaseModel{ID: userID},
					Username:  "testuser",
					Email:     "test@example.com",
					Nickname:  "Test User",
					PictureID: pictureID,
					Status:    "online",
					Bio:       "Test bio",
				}, nil
			},
		}

		fileService := new(mocks.FileUploadService)
		fileService.On("GetFileURLByID", pictureID.Hex()).Return("https://example.com/avatar.jpg", (*models.MessageOptions)(nil))

		service := NewUserService(nil, nil, mockRepo, fileService)

		profile, err := service.GetUserProfile(userID.Hex())

		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "testuser", profile.Username)
		assert.Equal(t, "Test bio", profile.Bio)
		assert.Equal(t, "https://example.com/avatar.jpg", profile.PictureURL)
	})

	t.Run("用戶不存在", func(t *testing.T) {
		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return nil, fmt.Errorf("user not found")
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		profile, err := service.GetUserProfile(primitive.NewObjectID().Hex())

		assert.Error(t, err)
		assert.Nil(t, profile)
	})
}

// TestUpdateUserProfile 測試更新用戶個人資料
func TestUpdateUserProfile(t *testing.T) {
	t.Run("成功更新用戶資料", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			updateUserFunc: func(id string, updates map[string]any) error {
				called = true
				assert.Equal(t, userID, id)
				assert.Contains(t, updates, "nickname")
				assert.Equal(t, "New Nickname", updates["nickname"])
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		updates := map[string]any{
			"nickname": "New Nickname",
			"bio":      "New bio",
		}

		err := service.UpdateUserProfile(userID, updates)

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("過濾不允許的欄位", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()

		mockRepo := &testUserRepository{
			updateUserFunc: func(id string, updates map[string]any) error {
				// 不應該包含 password 欄位
				assert.NotContains(t, updates, "password")
				assert.NotContains(t, updates, "email")
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		updates := map[string]any{
			"nickname": "New Nickname",
			"password": "should_be_filtered",
			"email":    "should_be_filtered",
		}

		err := service.UpdateUserProfile(userID, updates)

		assert.NoError(t, err)
	})

	t.Run("沒有需要更新的欄位", func(t *testing.T) {
		mockRepo := &testUserRepository{
			updateUserFunc: func(id string, updates map[string]any) error {
				t.Fatal("不應該調用 UpdateUser")
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		updates := map[string]any{
			"invalid_field": "value",
		}

		err := service.UpdateUserProfile(primitive.NewObjectID().Hex(), updates)

		assert.NoError(t, err)
	})
}

// TestDeleteUserAvatar 測試刪除用戶頭像
func TestDeleteUserAvatar(t *testing.T) {
	t.Run("成功刪除頭像", func(t *testing.T) {
		userID := primitive.NewObjectID()
		pictureID := primitive.NewObjectID()

		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return &models.User{
					BaseModel: providers.BaseModel{ID: userID},
					PictureID: pictureID,
				}, nil
			},
			updateUserFunc: func(id string, updates map[string]any) error {
				assert.Contains(t, updates, "picture_id")
				return nil
			},
		}

		fileService := new(mocks.FileUploadService)
		fileService.On("DeleteFileByID", pictureID.Hex(), userID.Hex()).Return((*models.MessageOptions)(nil))

		service := NewUserService(nil, nil, mockRepo, fileService)

		err := service.DeleteUserAvatar(userID.Hex())

		assert.NoError(t, err)
	})

	t.Run("用戶沒有頭像", func(t *testing.T) {
		userID := primitive.NewObjectID()

		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return &models.User{
					BaseModel: providers.BaseModel{ID: userID},
					PictureID: primitive.NilObjectID,
				}, nil
			},
			updateUserFunc: func(id string, updates map[string]any) error {
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.DeleteUserAvatar(userID.Hex())

		assert.NoError(t, err)
	})
}

// TestDeleteUserBanner 測試刪除用戶橫幅
func TestDeleteUserBanner(t *testing.T) {
	t.Run("成功刪除橫幅", func(t *testing.T) {
		userID := primitive.NewObjectID()
		bannerID := primitive.NewObjectID()

		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return &models.User{
					BaseModel: providers.BaseModel{ID: userID},
					BannerID:  bannerID,
				}, nil
			},
			updateUserFunc: func(id string, updates map[string]any) error {
				assert.Contains(t, updates, "banner_id")
				return nil
			},
		}

		fileService := new(mocks.FileUploadService)
		fileService.On("DeleteFileByID", bannerID.Hex(), userID.Hex()).Return((*models.MessageOptions)(nil))

		service := NewUserService(nil, nil, mockRepo, fileService)

		err := service.DeleteUserBanner(userID.Hex())

		assert.NoError(t, err)
	})
}

// TestUpdateUserPassword 測試更新用戶密碼
func TestUpdateUserPassword(t *testing.T) {
	t.Run("成功更新密碼", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		newPassword := "newpassword123"
		called := false

		mockRepo := &testUserRepository{
			updateUserFunc: func(id string, updates map[string]any) error {
				called = true
				assert.Equal(t, userID, id)
				assert.Contains(t, updates, "password")
				// 密碼應該被雜湊
				hashedPassword := updates["password"].(string)
				assert.NotEqual(t, newPassword, hashedPassword)
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.UpdateUserPassword(userID, newPassword)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

// TestGetTwoFactorStatus 測試獲取兩步驟驗證狀態
func TestGetTwoFactorStatus(t *testing.T) {
	t.Run("成功獲取兩步驟驗證狀態", func(t *testing.T) {
		userID := primitive.NewObjectID()

		mockRepo := &testUserRepository{
			getUserByIdFunc: func(id string) (*models.User, error) {
				return &models.User{
					BaseModel:        providers.BaseModel{ID: userID},
					TwoFactorEnabled: true,
				}, nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		status, err := service.GetTwoFactorStatus(userID.Hex())

		assert.NoError(t, err)
		assert.NotNil(t, status)
		assert.True(t, status.Enabled)
	})
}

// TestUpdateTwoFactorStatus 測試更新兩步驟驗證狀態
func TestUpdateTwoFactorStatus(t *testing.T) {
	t.Run("成功啟用兩步驟驗證", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			updateUserFunc: func(id string, updates map[string]any) error {
				called = true
				assert.Equal(t, userID, id)
				assert.Contains(t, updates, "two_factor_enabled")
				assert.Equal(t, true, updates["two_factor_enabled"])
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.UpdateTwoFactorStatus(userID, true)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

// TestDeactivateAccount 測試停用帳號
func TestDeactivateAccount(t *testing.T) {
	t.Run("成功停用帳號", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			updateUserFunc: func(id string, updates map[string]any) error {
				called = true
				assert.Equal(t, userID, id)
				assert.Contains(t, updates, "is_active")
				assert.Equal(t, false, updates["is_active"])
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.DeactivateAccount(userID)

		assert.NoError(t, err)
		assert.True(t, called)
	})
}

// TestDeleteAccount 測試刪除帳號
func TestDeleteAccount(t *testing.T) {
	t.Run("成功刪除帳號", func(t *testing.T) {
		userID := primitive.NewObjectID().Hex()
		called := false

		mockRepo := &testUserRepository{
			deleteUserFunc: func(id string) error {
				called = true
				assert.Equal(t, userID, id)
				return nil
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.DeleteAccount(userID)

		assert.NoError(t, err)
		assert.True(t, called)
	})

	t.Run("刪除帳號失敗", func(t *testing.T) {
		mockRepo := &testUserRepository{
			deleteUserFunc: func(id string) error {
				return fmt.Errorf("database error")
			},
		}

		service := NewUserService(nil, nil, mockRepo, nil)

		err := service.DeleteAccount(primitive.NewObjectID().Hex())

		assert.Error(t, err)
	})
}
