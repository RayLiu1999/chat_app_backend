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

// TestNewFriendController 測試創建 FriendController
func TestNewFriendController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}
	mockFriendService := new(mocks.FriendService)

	controller := NewFriendController(cfg, nil, mockFriendService)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
	assert.Equal(t, mockFriendService, controller.friendService)
}

// TestFriendController_GetFriendList 測試獲取好友列表
func TestFriendController_GetFriendList(t *testing.T) {
	t.Run("成功獲取好友列表", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)

		friendID := primitive.NewObjectID()
		expectedFriends := []models.FriendResponse{
			{
				ID:       friendID.Hex(),
				Name:     "Friend One",
				Status:   "accepted",
				IsOnline: true,
			},
		}

		mockFriendService.On("GetFriendList", "user123").Return(expectedFriends, nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/friends", controller.GetFriendList)

		req, _ := http.NewRequest(http.MethodGet, "/friends", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "好友列表獲取成功", response.Message)

		mockFriendService.AssertExpectations(t)
	})

	t.Run("服務層錯誤", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("GetFriendList", "user123").Return(
			nil,
			&models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "獲取好友列表失敗",
			},
		)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/friends", controller.GetFriendList)

		req, _ := http.NewRequest(http.MethodGet, "/friends", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, models.ErrInternalServer, response.Code)

		mockFriendService.AssertExpectations(t)
	})
}

// TestFriendController_GetPendingRequests 測試獲取待處理好友請求
func TestFriendController_GetPendingRequests(t *testing.T) {
	t.Run("成功獲取待處理好友請求", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)

		expectedRequests := &models.PendingRequestsResponse{
			Sent: []models.PendingFriendRequest{
				{
					RequestID: "req1",
					Username:  "user1",
				},
			},
			Received: []models.PendingFriendRequest{
				{
					RequestID: "req2",
					Username:  "user2",
				},
			},
		}

		mockFriendService.On("GetPendingRequests", "user123").Return(expectedRequests, nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.GET("/friends/requests/pending", controller.GetPendingRequests)

		req, _ := http.NewRequest(http.MethodGet, "/friends/requests/pending", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})
}

// TestFriendController_SendFriendRequest 測試發送好友請求
func TestFriendController_SendFriendRequest(t *testing.T) {
	t.Run("成功發送好友請求", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("AddFriendRequest", "user123", "targetUser").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/request", controller.SendFriendRequest)

		requestBody := map[string]string{
			"username": "targetUser",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/friends/request", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})

	t.Run("缺少用戶名", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/request", controller.SendFriendRequest)

		requestBody := map[string]string{}
		jsonBody, _ := json.Marshal(requestBody)

		req, _ := http.NewRequest(http.MethodPost, "/friends/request", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)
	})
}

// TestFriendController_AcceptFriendRequest 測試接受好友請求
func TestFriendController_AcceptFriendRequest(t *testing.T) {
	t.Run("成功接受好友請求", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("AcceptFriendRequest", "user123", "request123").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/request/:request_id/accept", controller.AcceptFriendRequest)

		req, _ := http.NewRequest(http.MethodPost, "/friends/request/request123/accept", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})

	t.Run("請求ID為空", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/request/:request_id/accept", controller.AcceptFriendRequest)

		req, _ := http.NewRequest(http.MethodPost, "/friends/request//accept", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestFriendController_DeclineFriendRequest 測試拒絕好友請求
func TestFriendController_DeclineFriendRequest(t *testing.T) {
	t.Run("成功拒絕好友請求", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("DeclineFriendRequest", "user123", "request123").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/request/:request_id/decline", controller.DeclineFriendRequest)

		req, _ := http.NewRequest(http.MethodPost, "/friends/request/request123/decline", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})
}

// TestFriendController_CancelFriendRequest 測試取消好友請求
func TestFriendController_CancelFriendRequest(t *testing.T) {
	t.Run("成功取消好友請求", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("CancelFriendRequest", "user123", "request123").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/request/:request_id/cancel", controller.CancelFriendRequest)

		req, _ := http.NewRequest(http.MethodPost, "/friends/request/request123/cancel", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})
}

// TestFriendController_RemoveFriend 測試移除好友
func TestFriendController_RemoveFriend(t *testing.T) {
	t.Run("成功移除好友", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("RemoveFriend", "user123", "friend123").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.DELETE("/friends/:friend_id", controller.RemoveFriend)

		req, _ := http.NewRequest(http.MethodDelete, "/friends/friend123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})

	t.Run("移除好友失敗", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("RemoveFriend", "user123", "friend123").Return(
			&models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Message: "好友不存在",
			},
		)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.DELETE("/friends/:friend_id", controller.RemoveFriend)

		req, _ := http.NewRequest(http.MethodDelete, "/friends/friend123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response.Status)

		mockFriendService.AssertExpectations(t)
	})
}

// TestFriendController_BlockUser 測試封鎖用戶
func TestFriendController_BlockUser(t *testing.T) {
	t.Run("成功封鎖用戶", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("BlockUser", "user123", "target123").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/block/:user_id", controller.BlockUser)

		req, _ := http.NewRequest(http.MethodPost, "/friends/block/target123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})
}

// TestFriendController_UnblockUser 測試解除封鎖
func TestFriendController_UnblockUser(t *testing.T) {
	t.Run("成功解除封鎖", func(t *testing.T) {
		mockFriendService := new(mocks.FriendService)
		mockFriendService.On("UnblockUser", "user123", "target123").Return(nil)

		controller := NewFriendController(&config.Config{}, nil, mockFriendService)

		router := setupTestRouter()
		router.Use(mocks.MockAuthMiddleware("user123"))
		router.POST("/friends/unblock/:user_id", controller.UnblockUser)

		req, _ := http.NewRequest(http.MethodPost, "/friends/unblock/target123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFriendService.AssertExpectations(t)
	})
}
