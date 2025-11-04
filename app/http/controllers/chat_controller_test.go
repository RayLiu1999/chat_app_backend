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

// TestNewChatController 測試創建 ChatController
func TestNewChatController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}
	mockChatService := new(mocks.ChatService)
	mockUserService := new(mocks.UserService)

	controller := NewChatController(cfg, nil, mockChatService, mockUserService)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
	assert.Equal(t, mockChatService, controller.chatService)
}

// TestChatController_GetDMRoomList 測試獲取聊天列表
func TestChatController_GetDMRoomList(t *testing.T) {
	t.Run("成功獲取聊天列表", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)

		roomID1 := primitive.NewObjectID()
		roomID2 := primitive.NewObjectID()

		expectedRooms := []models.DMRoomResponse{
			{
				RoomID:   roomID1,
				Nickname: "Friend One",
				IsOnline: true,
			},
			{
				RoomID:   roomID2,
				Nickname: "Friend Two",
				IsOnline: false,
			},
		}

		mockChatService.On("GetDMRoomResponseList", "user123", false).Return(expectedRooms, nil)

		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/chat/dm-rooms", controller.GetDMRoomList)

		req, _ := http.NewRequest(http.MethodGet, "/chat/dm-rooms", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "獲取聊天列表成功", response.Message)

		mockChatService.AssertExpectations(t)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)
		mockChatService.On("GetDMRoomResponseList", "user123", false).Return(
			nil,
			&models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取聊天列表失敗",
			},
		)

		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/chat/dm-rooms", controller.GetDMRoomList)

		req, _ := http.NewRequest(http.MethodGet, "/chat/dm-rooms", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInternalServer, response.Code)

		mockChatService.AssertExpectations(t)
	})
}

// TestChatController_CreateDMRoom 測試創建私聊房間
func TestChatController_CreateDMRoom(t *testing.T) {
	t.Run("成功創建私聊房間", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)

		roomID := primitive.NewObjectID()
		expectedRoom := &models.DMRoomResponse{
			RoomID:   roomID,
			Nickname: "New Friend",
		}

		mockChatService.On("CreateDMRoom", "user123", "user456").Return(expectedRoom, nil)

		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/chat/dm-rooms", controller.CreateDMRoom)

		requestBody := map[string]string{
			"chat_with_user_id": "user456",
		}
		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPost, "/chat/dm-rooms", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "聊天列表已創建", response.Message)

		mockChatService.AssertExpectations(t)
	})

	t.Run("無效的請求參數", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)
		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/chat/dm-rooms", controller.CreateDMRoom)

		req, _ := http.NewRequest(http.MethodPost, "/chat/dm-rooms", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
	})
}

// TestChatController_UpdateDMRoom 測試更新聊天房間
func TestChatController_UpdateDMRoom(t *testing.T) {
	t.Run("成功更新聊天房間", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)
		mockChatService.On("UpdateDMRoom", "user123", "room1", true).Return(nil)

		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/chat/dm-rooms", controller.UpdateDMRoom)

		requestBody := map[string]interface{}{
			"room_id":   "room1",
			"is_hidden": true,
		}
		body, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest(http.MethodPut, "/chat/dm-rooms", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "聊天列表保存成功", response.Message)

		mockChatService.AssertExpectations(t)
	})

	t.Run("無效的請求參數", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)
		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.PUT("/chat/dm-rooms", controller.UpdateDMRoom)

		req, _ := http.NewRequest(http.MethodPut, "/chat/dm-rooms", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInvalidParams, response.Code)
	})
}

// TestChatController_GetDMMessages 測試獲取私聊訊息
func TestChatController_GetDMMessages(t *testing.T) {
	t.Run("成功獲取私聊訊息", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)

		msgID1 := primitive.NewObjectID()
		msgID2 := primitive.NewObjectID()

		expectedMessages := []models.MessageResponse{
			{
				ID:      msgID1,
				Content: "Hello",
			},
			{
				ID:      msgID2,
				Content: "World",
			},
		}

		mockChatService.On("GetDMMessages", "user123", "room1", "", "", "").Return(expectedMessages, nil)

		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/chat/dm-rooms/:room_id/messages", controller.GetDMMessages)

		req, _ := http.NewRequest(http.MethodGet, "/chat/dm-rooms/room1/messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "獲取訊息成功", response.Message)

		mockChatService.AssertExpectations(t)
	})
}

// TestChatController_GetChannelMessages 測試獲取頻道訊息
func TestChatController_GetChannelMessages(t *testing.T) {
	t.Run("成功獲取頻道訊息", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)

		msgID1 := primitive.NewObjectID()
		msgID2 := primitive.NewObjectID()

		expectedMessages := []models.MessageResponse{
			{
				ID:      msgID1,
				Content: "Channel message 1",
			},
			{
				ID:      msgID2,
				Content: "Channel message 2",
			},
		}

		mockChatService.On("GetChannelMessages", "user123", "channel1", "", "", "").Return(expectedMessages, nil)

		controller := NewChatController(&config.Config{}, nil, mockChatService, nil)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/channels/:channel_id/messages", controller.GetChannelMessages)

		req, _ := http.NewRequest(http.MethodGet, "/channels/channel1/messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "獲取頻道訊息成功", response.Message)

		mockChatService.AssertExpectations(t)
	})

	t.Run("頻道ID為空", func(t *testing.T) {
		mockChatService := new(mocks.ChatService)
		mockUserService := new(mocks.UserService)
		controller := NewChatController(&config.Config{}, nil, mockChatService, mockUserService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/chat/channels/:channel_id/messages", controller.GetChannelMessages)

		req, _ := http.NewRequest(http.MethodGet, "/chat/channels//messages", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
	})
}
