package controllers

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestNewServerController 測試創建 ServerController
func TestNewServerController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}
	mockServerService := new(mocks.ServerService)

	controller := NewServerController(cfg, nil, mockServerService)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
	assert.Equal(t, mockServerService, controller.serverService)
}

// TestServerController_GetServerList 測試獲取伺服器列表
func TestServerController_GetServerList(t *testing.T) {
	t.Run("成功獲取伺服器列表", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)

		serverID := primitive.NewObjectID()
		expectedServers := []models.ServerResponse{
			{
				ID:   serverID,
				Name: "Test Server",
			},
		}

		mockServerService.On("GetServerListResponse", "user123").Return(expectedServers, nil)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/servers", controller.GetServerList)

		req, _ := http.NewRequest(http.MethodGet, "/servers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "伺服器列表獲取成功", response.Message)

		mockServerService.AssertExpectations(t)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)
		mockServerService.On("GetServerListResponse", "user123").Return(
			nil,
			&models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取伺服器列表失敗",
			},
		)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/servers", controller.GetServerList)

		req, _ := http.NewRequest(http.MethodGet, "/servers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInternalServer, response.Code)

		mockServerService.AssertExpectations(t)
	})
}

// TestServerController_CreateServer 測試創建伺服器
func TestServerController_CreateServer(t *testing.T) {
	t.Run("成功創建伺服器 - 無圖片", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)

		serverID := primitive.NewObjectID()
		expectedServer := &models.ServerResponse{
			ID:   serverID,
			Name: "New Server",
		}

		mockServerService.On("CreateServer", "user123", "New Server",
			multipart.File(nil), (*multipart.FileHeader)(nil)).Return(expectedServer, nil)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers", controller.CreateServer)

		// 創建 multipart form 請求
		body := &strings.Builder{}
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "New Server")
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/servers", strings.NewReader(body.String()))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "伺服器創建成功", response.Message)

		mockServerService.AssertExpectations(t)
	})

	t.Run("創建伺服器失敗 - 缺少名稱", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)
		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers", controller.CreateServer)

		// 創建空的 form 請求
		body := &strings.Builder{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/servers", strings.NewReader(body.String()))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)
		mockServerService.On("CreateServer", "user123", "New Server",
			multipart.File(nil), (*multipart.FileHeader)(nil)).Return(
			nil,
			&models.MessageOptions{
				Code:    models.ErrCreateServerFailed,
				Message: "創建伺服器失敗",
			},
		)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers", controller.CreateServer)

		body := &strings.Builder{}
		writer := multipart.NewWriter(body)
		writer.WriteField("name", "New Server")
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/servers", strings.NewReader(body.String()))
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrCreateServerFailed, response.Code)

		mockServerService.AssertExpectations(t)
	})
}

// TestServerController_SearchPublicServers 測試搜尋公開伺服器
func TestServerController_SearchPublicServers(t *testing.T) {
	t.Run("成功搜尋公開伺服器", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)

		serverID := primitive.NewObjectID()
		expectedResults := &models.ServerSearchResults{
			Servers: []models.ServerSearchResponse{
				{
					ID:          serverID,
					Name:        "Public Server",
					MemberCount: 10,
				},
			},
			TotalCount: 1,
			Page:       1,
			Limit:      20,
			TotalPages: 1,
		}

		searchRequest := models.ServerSearchRequest{
			Query:     "Public",
			Page:      1,
			Limit:     20,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		mockServerService.On("SearchPublicServers", "user123", searchRequest).Return(expectedResults, nil)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/servers/search", controller.SearchPublicServers)

		req, _ := http.NewRequest(http.MethodGet, "/servers/search?q=Public&page=1&limit=20&sort_by=created_at&sort_order=desc", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "搜尋完成", response.Message)

		mockServerService.AssertExpectations(t)
	})
}

// TestServerController_JoinServer 測試加入伺服器
func TestServerController_JoinServer(t *testing.T) {
	t.Run("成功加入伺服器", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)
		mockServerService.On("JoinServer", "user123", "server1").Return(nil)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/join", controller.JoinServer)

		req, _ := http.NewRequest(http.MethodPost, "/servers/server1/join", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockServerService.AssertExpectations(t)
	})

	t.Run("加入伺服器失敗", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)
		mockServerService.On("JoinServer", "user123", "server1").Return(
			&models.MessageOptions{
				Code:    models.ErrServerNotFound,
				Message: "伺服器不存在",
			},
		)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/join", controller.JoinServer)

		req, _ := http.NewRequest(http.MethodPost, "/servers/server1/join", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrServerNotFound, response.Code)

		mockServerService.AssertExpectations(t)
	})
}

// TestServerController_LeaveServer 測試離開伺服器
func TestServerController_LeaveServer(t *testing.T) {
	t.Run("成功離開伺服器", func(t *testing.T) {
		mockServerService := new(mocks.ServerService)
		mockServerService.On("LeaveServer", "user123", "server1").Return(nil)

		controller := NewServerController(&config.Config{}, nil, mockServerService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/leave", controller.LeaveServer)

		req, _ := http.NewRequest(http.MethodPost, "/servers/server1/leave", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockServerService.AssertExpectations(t)
	})
}
