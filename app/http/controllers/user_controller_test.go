package controllers

import (
	"bytes"
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupTestConfig 設置測試用的 config
func setupTestConfig() {
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			Server: config.ServerConfig{
				MainDomain:          "localhost",
				Port:                "8080",
				BaseURL:             "http://localhost:8080",
				Mode:                config.TestMode,
				AccessExpireMinutes: 15,
				RefreshExpireHours:  24,
			},
		}
	}
}

// setupTestRouter 設置測試用的 Gin 路由器
func setupTestRouter() *gin.Engine {
	setupTestConfig()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("config", &config.ServerConfig{
			MainDomain:          "localhost",
			Port:                "8080",
			BaseURL:             "http://localhost:8080",
			AccessExpireMinutes: 15,
			RefreshExpireHours:  24,
		})
		c.Next()
	})
	return router
}

// TestNewUserController 測試創建 UserController
func TestNewUserController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}
	mockUserService := new(mocks.UserService)

	controller := NewUserController(cfg, nil, mockUserService, nil)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
	assert.Equal(t, mockUserService, controller.userService)
	// 注意: clientManager 和 mongoConnect 未使用,傳入 nil 即可
}

// TestUserController_Register 測試用戶註冊
func TestUserController_Register(t *testing.T) {
	t.Run("成功註冊", func(t *testing.T) {
		// 準備測試數據
		user := models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		// 創建 mock service
		mockUserService := new(mocks.UserService)
		mockUserService.On("RegisterUser", mock.AnythingOfType("models.User")).Return(nil)

		// 創建 controller
		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		// 設置路由
		router := setupTestRouter()
		router.POST("/register", controller.Register)

		// 創建請求
		body, _ := json.Marshal(user)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// 執行請求
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 驗證結果
		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "用戶創建成功", response.Message)

		mockUserService.AssertExpectations(t)
	})

	t.Run("無效的請求參數", func(t *testing.T) {
		mockUserService := new(mocks.UserService)
		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/register", controller.Register)

		// 發送無效的 JSON
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		user := models.User{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		mockUserService := new(mocks.UserService)
		mockUserService.On("RegisterUser", mock.AnythingOfType("models.User")).Return(&models.MessageOptions{
			Code:    models.ErrUsernameExists,
			Message: "用戶已存在",
		})

		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/register", controller.Register)

		body, _ := json.Marshal(user)
		req, _ := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrUsernameExists, response.Code)

		mockUserService.AssertExpectations(t)
	})
}

// TestUserController_Login 測試用戶登入
func TestUserController_Login(t *testing.T) {
	t.Run("成功登入", func(t *testing.T) {
		loginUser := models.User{
			Username: "testuser",
			Password: "password123",
		}

		loginResponse := &models.LoginResponse{
			AccessToken:  "access_token_123",
			RefreshToken: "refresh_token_123",
			CSRFToken:    "csrf_token_123",
		}

		mockUserService := new(mocks.UserService)
		mockUserService.On("Login", mock.AnythingOfType("models.User")).Return(loginResponse, nil)

		cfg := &config.Config{
			Server: config.ServerConfig{
				RefreshExpireHours: 24,
			},
		}
		controller := NewUserController(cfg, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/login", controller.Login)

		body, _ := json.Marshal(loginUser)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "登入成功", response.Message)

		// 驗證返回的數據包含 access_token
		dataMap, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "access_token_123", dataMap["access_token"])

		// 驗證 cookies 被設置
		cookies := w.Result().Cookies()
		assert.True(t, len(cookies) > 0, "應該設置 cookies")

		mockUserService.AssertExpectations(t)
	})

	t.Run("無效的請求參數", func(t *testing.T) {
		mockUserService := new(mocks.UserService)
		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/login", controller.Login)

		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
	})

	t.Run("登入失敗 - 錯誤的憑證", func(t *testing.T) {
		loginUser := models.User{
			Username: "testuser",
			Password: "wrongpassword",
		}

		mockUserService := new(mocks.UserService)
		mockUserService.On("Login", mock.AnythingOfType("models.User")).Return(
			(*models.LoginResponse)(nil),
			&models.MessageOptions{
				Code:    models.ErrLoginFailed,
				Message: "用戶名或密碼錯誤",
			},
		)

		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/login", controller.Login)

		body, _ := json.Marshal(loginUser)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrLoginFailed, response.Code)

		mockUserService.AssertExpectations(t)
	})

	t.Run("服務層內部錯誤", func(t *testing.T) {
		loginUser := models.User{
			Username: "testuser",
			Password: "password123",
		}

		mockUserService := new(mocks.UserService)
		mockUserService.On("Login", mock.AnythingOfType("models.User")).Return(
			(*models.LoginResponse)(nil),
			&models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "內部服務器錯誤",
			},
		)

		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/login", controller.Login)

		body, _ := json.Marshal(loginUser)
		req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInternalServer, response.Code)

		mockUserService.AssertExpectations(t)
	})
}

// TestUserController_Logout 測試用戶登出
func TestUserController_Logout(t *testing.T) {
	t.Run("成功登出", func(t *testing.T) {
		mockUserService := new(mocks.UserService)
		mockUserService.On("Logout", mock.AnythingOfType("*gin.Context")).Return(nil)

		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/logout", controller.Logout)

		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "登出成功", response.Message)

		mockUserService.AssertExpectations(t)
	})

	t.Run("登出失敗", func(t *testing.T) {
		mockUserService := new(mocks.UserService)
		mockUserService.On("Logout", mock.AnythingOfType("*gin.Context")).Return(&models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "登出失敗",
		})

		controller := NewUserController(&config.Config{}, nil, mockUserService, nil)

		router := setupTestRouter()
		router.POST("/logout", controller.Logout)

		req, _ := http.NewRequest(http.MethodPost, "/logout", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInternalServer, response.Code)

		mockUserService.AssertExpectations(t)
	})
}
