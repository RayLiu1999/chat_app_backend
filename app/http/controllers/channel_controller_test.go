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

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestNewChannelController 測試創建 ChannelController
func TestNewChannelController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}
	mockChannelService := new(mocks.ChannelService)

	controller := NewChannelController(cfg, nil, mockChannelService)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
	assert.Equal(t, mockChannelService, controller.channelService)
}

// TestChannelController_GetChannelsByServerID 測試獲取伺服器的頻道列表
func TestChannelController_GetChannelsByServerID(t *testing.T) {
	t.Run("成功獲取頻道列表", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)

		channelID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		expectedChannels := []models.ChannelResponse{
			{
				ID:       channelID,
				ServerID: serverID,
				Name:     "一般",
				Type:     "text",
			},
		}

		mockChannelService.On("GetChannelsByServerID", "user123", "server123").
			Return(expectedChannels, (*models.MessageOptions)(nil))

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/servers/:server_id/channels", controller.GetChannelsByServerID)

		req, _ := http.NewRequest(http.MethodGet, "/servers/server123/channels", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "獲取頻道列表成功", response.Message)

		mockChannelService.AssertExpectations(t)
	})

	t.Run("伺服器ID為空", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/servers/:server_id/channels", controller.GetChannelsByServerID)

		req, _ := http.NewRequest(http.MethodGet, "/servers//channels", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		mockChannelService.On("GetChannelsByServerID", "user123", "server123").
			Return(nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取頻道列表失敗",
			})

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/servers/:server_id/channels", controller.GetChannelsByServerID)

		req, _ := http.NewRequest(http.MethodGet, "/servers/server123/channels", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)

		mockChannelService.AssertExpectations(t)
	})
}

// TestChannelController_GetChannelByID 測試獲取單個頻道信息
func TestChannelController_GetChannelByID(t *testing.T) {
	t.Run("成功獲取頻道信息", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)

		channelID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		expectedChannel := &models.ChannelResponse{
			ID:       channelID,
			ServerID: serverID,
			Name:     "一般",
			Type:     "text",
		}

		mockChannelService.On("GetChannelByID", "user123", channelID.Hex()).
			Return(expectedChannel, (*models.MessageOptions)(nil))

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/channels/:channel_id", controller.GetChannelByID)

		req, _ := http.NewRequest(http.MethodGet, "/channels/"+channelID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "獲取頻道信息成功", response.Message)

		mockChannelService.AssertExpectations(t)
	})

	t.Run("頻道ID為空", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/channels/:channel_id", controller.GetChannelByID)

		req, _ := http.NewRequest(http.MethodGet, "/channels/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("頻道ID格式無效", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/channels/:channel_id", controller.GetChannelByID)

		req, _ := http.NewRequest(http.MethodGet, "/channels/invalid-id", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
		assert.Contains(t, response.Message, "無效的頻道ID格式")
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		channelID := primitive.NewObjectID()

		mockChannelService.On("GetChannelByID", "user123", channelID.Hex()).
			Return(nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取頻道信息失敗",
			})

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/channels/:channel_id", controller.GetChannelByID)

		req, _ := http.NewRequest(http.MethodGet, "/channels/"+channelID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockChannelService.AssertExpectations(t)
	})
}

// TestChannelController_CreateChannel 測試創建新頻道
func TestChannelController_CreateChannel(t *testing.T) {
	t.Run("成功創建頻道", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)

		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		expectedChannel := &models.ChannelResponse{
			ID:       channelID,
			ServerID: serverID,
			Name:     "一般",
			Type:     "text",
		}

		mockChannelService.On("CreateChannel", "user123", &models.Channel{
			Name:     "一般",
			ServerID: serverID,
			Type:     "text",
		}).Return(expectedChannel, (*models.MessageOptions)(nil))

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/channels", controller.CreateChannel)

		requestBody := CreateChannelRequest{
			Name: "一般",
			Type: "text",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/servers/"+serverID.Hex()+"/channels", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "創建頻道成功", response.Message)

		mockChannelService.AssertExpectations(t)
	})

	t.Run("伺服器ID為空", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/channels", controller.CreateChannel)

		requestBody := CreateChannelRequest{
			Name: "一般",
			Type: "text",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/servers//channels", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("伺服器ID格式無效", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/channels", controller.CreateChannel)

		requestBody := CreateChannelRequest{
			Name: "一般",
			Type: "text",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/servers/invalid-id/channels", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "無效的伺服器ID格式")
	})

	t.Run("請求體無效", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/channels", controller.CreateChannel)

		serverID := primitive.NewObjectID()

		req, _ := http.NewRequest(http.MethodPost, "/servers/"+serverID.Hex()+"/channels", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("頻道名稱為空", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/channels", controller.CreateChannel)

		serverID := primitive.NewObjectID()
		requestBody := CreateChannelRequest{
			Name: "",
			Type: "text",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/servers/"+serverID.Hex()+"/channels", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		// 可能是 binding required 或 頻道名稱不能為空
		assert.Contains(t, []string{"無效的請求格式", "頻道名稱不能為空"}, response.Message)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		serverID := primitive.NewObjectID()

		mockChannelService.On("CreateChannel", "user123", &models.Channel{
			Name:     "一般",
			ServerID: serverID,
			Type:     "text",
		}).Return(nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "創建頻道失敗",
		})

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/servers/:server_id/channels", controller.CreateChannel)

		requestBody := CreateChannelRequest{
			Name: "一般",
			Type: "text",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/servers/"+serverID.Hex()+"/channels", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockChannelService.AssertExpectations(t)
	})
}

// TestChannelController_UpdateChannel 測試更新頻道信息
func TestChannelController_UpdateChannel(t *testing.T) {
	t.Run("成功更新頻道", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		channelID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		updates := map[string]any{
			"name": "新名稱",
		}

		expectedChannel := &models.ChannelResponse{
			ID:       channelID,
			ServerID: serverID,
			Name:     "新名稱",
			Type:     "text",
		}

		mockChannelService.On("UpdateChannel", "user123", channelID.Hex(), updates).
			Return(expectedChannel, (*models.MessageOptions)(nil))

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/channels/:channel_id", controller.UpdateChannel)

		requestBody := UpdateChannelRequest{
			Name: "新名稱",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPut, "/channels/"+channelID.Hex(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "更新頻道成功", response.Message)

		mockChannelService.AssertExpectations(t)
	})

	t.Run("頻道ID為空", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/channels/:channel_id", controller.UpdateChannel)

		requestBody := UpdateChannelRequest{
			Name: "新名稱",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPut, "/channels/", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("頻道ID格式無效", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/channels/:channel_id", controller.UpdateChannel)

		requestBody := UpdateChannelRequest{
			Name: "新名稱",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPut, "/channels/invalid-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "無效的頻道ID格式")
	})

	t.Run("請求體無效", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/channels/:channel_id", controller.UpdateChannel)

		channelID := primitive.NewObjectID()

		req, _ := http.NewRequest(http.MethodPut, "/channels/"+channelID.Hex(), bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("沒有提供要更新的字段", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/channels/:channel_id", controller.UpdateChannel)

		channelID := primitive.NewObjectID()
		requestBody := UpdateChannelRequest{}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPut, "/channels/"+channelID.Hex(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "沒有提供要更新的字段")
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		channelID := primitive.NewObjectID()

		updates := map[string]any{
			"name": "新名稱",
		}

		mockChannelService.On("UpdateChannel", "user123", channelID.Hex(), updates).
			Return(nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "更新頻道失敗",
			})

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/channels/:channel_id", controller.UpdateChannel)

		requestBody := UpdateChannelRequest{
			Name: "新名稱",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPut, "/channels/"+channelID.Hex(), bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockChannelService.AssertExpectations(t)
	})
}

// TestChannelController_DeleteChannel 測試刪除頻道
func TestChannelController_DeleteChannel(t *testing.T) {
	t.Run("成功刪除頻道", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		channelID := primitive.NewObjectID()

		mockChannelService.On("DeleteChannel", "user123", channelID.Hex()).
			Return((*models.MessageOptions)(nil))

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.DELETE("/channels/:channel_id", controller.DeleteChannel)

		req, _ := http.NewRequest(http.MethodDelete, "/channels/"+channelID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "刪除頻道成功", response.Message)

		mockChannelService.AssertExpectations(t)
	})

	t.Run("頻道ID為空", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.DELETE("/channels/:channel_id", controller.DeleteChannel)

		req, _ := http.NewRequest(http.MethodDelete, "/channels/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("頻道ID格式無效", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.DELETE("/channels/:channel_id", controller.DeleteChannel)

		req, _ := http.NewRequest(http.MethodDelete, "/channels/invalid-id", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Contains(t, response.Message, "無效的頻道ID格式")
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockChannelService := new(mocks.ChannelService)
		channelID := primitive.NewObjectID()

		mockChannelService.On("DeleteChannel", "user123", channelID.Hex()).
			Return(&models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "刪除頻道失敗",
			})

		controller := NewChannelController(&config.Config{}, nil, mockChannelService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.DELETE("/channels/:channel_id", controller.DeleteChannel)

		req, _ := http.NewRequest(http.MethodDelete, "/channels/"+channelID.Hex(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)

		mockChannelService.AssertExpectations(t)
	})
}
