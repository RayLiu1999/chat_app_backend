package services

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Test Helpers ---

// mockCache 模擬 providers.CacheProvider
type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *mockCache) Set(key string, value string, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *mockCache) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// mockClientManager 模擬 ClientManager
type mockClientManager struct {
	mock.Mock
}

func (m *mockClientManager) NewClient(userID string, ws *websocket.Conn) *Client {
	args := m.Called(userID, ws)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*Client)
}

func (m *mockClientManager) Register(client *Client) {
	m.Called(client)
}

func (m *mockClientManager) Unregister(client *Client) {
	m.Called(client)
}

func (m *mockClientManager) GetClient(userID string) (*Client, bool) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*Client), args.Bool(1)
}

func (m *mockClientManager) GetAllClients() map[*Client]bool {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(map[*Client]bool)
}

func (m *mockClientManager) IsUserOnline(userID string) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

// mockRoomManager 模擬 RoomManager
type mockRoomManager struct {
	mock.Mock
}

func (m *mockRoomManager) CheckUserAllowedJoinRoom(userID string, roomID string, roomType models.RoomType) (bool, error) {
	args := m.Called(userID, roomID, roomType)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoomManager) GetRoom(roomType models.RoomType, roomID string) (*Room, bool) {
	args := m.Called(roomType, roomID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*Room), args.Bool(1)
}

func (m *mockRoomManager) AddRoom(room *Room) {
	m.Called(room)
}

func (m *mockRoomManager) InitRoom(roomType models.RoomType, roomID string) *Room {
	args := m.Called(roomType, roomID)
	if args.Get(0) == nil {
		return nil

	}
	return args.Get(0).(*Room)
}

func (m *mockRoomManager) JoinRoom(client *Client, roomType models.RoomType, roomID string) {
	m.Called(client, roomType, roomID)
}

func (m *mockRoomManager) LeaveRoom(client *Client, roomType models.RoomType, roomID string) {
	m.Called(client, roomType, roomID)
}

// mockMessageHandler 模擬 MessageHandler
type mockMessageHandler struct {
	mock.Mock
}

func (m *mockMessageHandler) HandleMessage(message *MessageResponse) {
	m.Called(message)
}

// mockUserService 模擬 UserService
type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) GetUserResponseById(userID string) (*models.UserResponse, error) {
	return nil, nil
}

func (m *mockUserService) GetUserByUsername(username string) (*models.User, *models.MessageOptions) {
	return nil, nil
}

func (m *mockUserService) RegisterUser(user models.User) *models.MessageOptions {
	return nil
}

func (m *mockUserService) Login(loginUser models.User) (*models.LoginResponse, *models.MessageOptions) {
	return nil, nil
}

func (m *mockUserService) Logout(c *gin.Context) *models.MessageOptions {
	return nil
}

func (m *mockUserService) RefreshToken(refreshToken string) (*models.RefreshTokenResponse, *models.MessageOptions) {
	return nil, nil
}

func (m *mockUserService) SetUserOnline(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockUserService) SetUserOffline(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockUserService) UpdateUserActivity(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *mockUserService) CheckAndSetOfflineUsers(offlineThresholdMinutes int) error {
	return nil
}

func (m *mockUserService) GetUserProfile(userID string) (*models.UserProfileResponse, error) {
	return nil, nil
}

func (m *mockUserService) UpdateUserProfile(userID string, updates map[string]any) error {
	return nil
}

func (m *mockUserService) UploadUserImage(userID string, file multipart.File, header *multipart.FileHeader, imageType string) (*models.UserImageResponse, error) {
	return nil, nil
}

func (m *mockUserService) DeleteUserAvatar(userID string) error {
	return nil
}

func (m *mockUserService) DeleteUserBanner(userID string) error {
	return nil
}

func (m *mockUserService) UpdateUserPassword(userID string, newPassword string) error {
	return nil
}

func (m *mockUserService) GetTwoFactorStatus(userID string) (*models.TwoFactorStatusResponse, error) {
	return nil, nil
}

func (m *mockUserService) UpdateTwoFactorStatus(userID string, enabled bool) error {
	return nil
}

func (m *mockUserService) DeactivateAccount(userID string) error {
	return nil
}

func (m *mockUserService) DeleteAccount(userID string) error {
	return nil
}

// mockCacheProvider 模擬 CacheProvider
type mockCacheProvider struct {
	mock.Mock
}

func (m *mockCacheProvider) Set(key string, value string, expiration time.Duration) error {
	args := m.Called(key, value, expiration)
	return args.Error(0)
}

func (m *mockCacheProvider) Get(key string) (string, error) {
	args := m.Called(key)
	return args.String(0), args.Error(1)
}

func (m *mockCacheProvider) Delete(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

func (m *mockCacheProvider) Exists(key string) (bool, error) {
	args := m.Called(key)
	return args.Bool(0), args.Error(1)
}

func (m *mockCacheProvider) SetWithContext(ctx context.Context, key string, value string, expiration time.Duration) error {
	return nil
}

func (m *mockCacheProvider) GetWithContext(ctx context.Context, key string) (string, error) {
	return "", nil
}

func (m *mockCacheProvider) GetRedisClient() interface{} {
	return nil
}

// --- Tests ---

func TestNewWebSocketHandler(t *testing.T) {
	mockODM := new(mocks.ODM)
	mockCM := new(mockClientManager)
	mockRM := new(mockRoomManager)
	mockMH := new(mockMessageHandler)
	mockUS := new(mockUserService)
	mockCache := new(mockCacheProvider)

	handler := NewWebSocketHandler(mockODM, mockCM, mockRM, mockMH, mockUS, mockCache)

	assert.NotNil(t, handler)
	assert.Equal(t, mockODM, handler.odm)
	assert.Equal(t, mockCM, handler.clientManager)
	assert.Equal(t, mockRM, handler.roomManager)
	assert.Equal(t, mockMH, handler.messageHandler)
	assert.Equal(t, mockUS, handler.userService)
	assert.Equal(t, mockCache, handler.cache)
}

func TestHandleJoinRoom(t *testing.T) {
	roomID := primitive.NewObjectID().Hex()
	userID := primitive.NewObjectID().Hex()

	t.Run("成功加入房間", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		handler := &webSocketHandler{
			roomManager: mockRM,
		}

		// 創建測試客戶端
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		// 準備請求數據
		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeChannel,
		}
		data, _ := json.Marshal(requestData)

		// 設定 mock 期望
		mockRM.On("CheckUserAllowedJoinRoom", userID, roomID, models.RoomTypeChannel).Return(true, nil).Once()
		mockRM.On("InitRoom", models.RoomTypeChannel, roomID).Return(&Room{}).Once()
		mockRM.On("JoinRoom", client, models.RoomTypeChannel, roomID).Once()

		// 執行
		handler.handleJoinRoom(client, data)

		// 驗證發送的訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[WsStatusResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "room_joined", response.Action)
			assert.Equal(t, "success", response.Data.Status)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到訊息")
		}

		mockRM.AssertExpectations(t)
	})

	t.Run("無效的請求數據", func(t *testing.T) {
		handler := &webSocketHandler{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		// 無效的 JSON
		invalidData := json.RawMessage(`{invalid json`)

		handler.handleJoinRoom(client, invalidData)

		// 驗證發送了錯誤訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[ErrorResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Action)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到錯誤訊息")
		}
	})

	t.Run("用戶無權限加入房間", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		handler := &webSocketHandler{
			roomManager: mockRM,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeChannel,
		}
		data, _ := json.Marshal(requestData)

		// 設定 mock：用戶無權限
		mockRM.On("CheckUserAllowedJoinRoom", userID, roomID, models.RoomTypeChannel).Return(false, nil).Once()

		handler.handleJoinRoom(client, data)

		// 驗證發送了錯誤訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[ErrorResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Action)
			assert.Contains(t, response.Data.Message, "沒有權限")
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到錯誤訊息")
		}

		mockRM.AssertExpectations(t)
	})

	t.Run("檢查權限時發生錯誤", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		handler := &webSocketHandler{
			roomManager: mockRM,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeChannel,
		}
		data, _ := json.Marshal(requestData)

		// 設定 mock：檢查權限時發生錯誤
		mockRM.On("CheckUserAllowedJoinRoom", userID, roomID, models.RoomTypeChannel).Return(false, errors.New("database error")).Once()

		handler.handleJoinRoom(client, data)

		// 驗證發送了錯誤訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[ErrorResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Action)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到錯誤訊息")
		}

		mockRM.AssertExpectations(t)
	})
}

func TestHandleLeaveRoom(t *testing.T) {
	roomID := primitive.NewObjectID().Hex()
	userID := primitive.NewObjectID().Hex()

	t.Run("成功離開房間", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		handler := &webSocketHandler{
			roomManager: mockRM,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeChannel,
		}
		data, _ := json.Marshal(requestData)

		// 設定 mock
		mockRM.On("LeaveRoom", client, models.RoomTypeChannel, roomID).Once()

		handler.handleLeaveRoom(client, data)

		// 驗證發送的訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[WsStatusResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "room_left", response.Action)
			assert.Equal(t, "success", response.Data.Status)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到訊息")
		}

		mockRM.AssertExpectations(t)
	})

	t.Run("無效的請求數據", func(t *testing.T) {
		handler := &webSocketHandler{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		invalidData := json.RawMessage(`{invalid`)

		handler.handleLeaveRoom(client, invalidData)

		// 驗證發送了錯誤訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[ErrorResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Action)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到錯誤訊息")
		}
	})
}

func TestHandleSendMessage(t *testing.T) {
	roomID := primitive.NewObjectID().Hex()
	userID := primitive.NewObjectID().Hex()

	t.Run("成功發送訊息", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		mockMH := new(mockMessageHandler)
		mockODM := new(mocks.ODM)

		handler := &webSocketHandler{
			roomManager:    mockRM,
			messageHandler: mockMH,
			odm:            mockODM,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
			Content  string          `json:"content"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeChannel,
			Content:  "Hello, World!",
		}
		data, _ := json.Marshal(requestData)

		// 設定 mock
		mockRM.On("InitRoom", models.RoomTypeChannel, roomID).Return(&Room{}).Once()
		mockMH.On("HandleMessage", mock.AnythingOfType("*services.MessageResponse")).Once()

		handler.handleSendMessage(client, data)

		mockRM.AssertExpectations(t)
		mockMH.AssertExpectations(t)
	})

	t.Run("DM 房間自動創建", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		mockMH := new(mockMessageHandler)
		mockODM := new(mocks.ODM)

		handler := &webSocketHandler{
			roomManager:    mockRM,
			messageHandler: mockMH,
			odm:            mockODM,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
			Content  string          `json:"content"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeDM,
			Content:  "Private message",
		}
		data, _ := json.Marshal(requestData)

		// 設定 mock - 模擬找到一個 DM 房間（需要創建對方的房間）
		partnerID := primitive.NewObjectID()
		dmRooms := []models.DMRoom{
			{
				RoomID:         primitive.ObjectID{},
				UserID:         partnerID,
				ChatWithUserID: primitive.ObjectID{},
			},
		}
		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.DMRoom)
			*arg = dmRooms
		}).Return(nil).Once()

		// 需要創建新的 DM 房間
		mockODM.On("Create", mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil).Once()

		mockRM.On("InitRoom", models.RoomTypeDM, roomID).Return(&Room{}).Once()
		mockMH.On("HandleMessage", mock.AnythingOfType("*services.MessageResponse")).Once()

		handler.handleSendMessage(client, data)

		mockRM.AssertExpectations(t)
		mockMH.AssertExpectations(t)
		mockODM.AssertExpectations(t)
	})

	t.Run("無效的請求數據", func(t *testing.T) {
		handler := &webSocketHandler{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		invalidData := json.RawMessage(`{invalid`)

		handler.handleSendMessage(client, invalidData)

		// 驗證發送了錯誤訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[ErrorResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Action)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到錯誤訊息")
		}
	})
}

func TestHandlePing(t *testing.T) {
	userID := primitive.NewObjectID().Hex()

	t.Run("成功處理 Ping", func(t *testing.T) {
		handler := &webSocketHandler{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		handler.handlePing(client)

		// 驗證發送的 pong 訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[PingResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "pong", response.Action)
			assert.Greater(t, response.Data.Timestamp, int64(0))
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到 pong 訊息")
		}
	})
}

func TestHandleClientMessage(t *testing.T) {
	userID := primitive.NewObjectID().Hex()

	t.Run("處理 join_room 動作", func(t *testing.T) {
		mockRM := new(mockRoomManager)
		handler := &webSocketHandler{
			roomManager: mockRM,
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		roomID := primitive.NewObjectID().Hex()
		requestData := struct {
			RoomID   string          `json:"room_id"`
			RoomType models.RoomType `json:"room_type"`
		}{
			RoomID:   roomID,
			RoomType: models.RoomTypeChannel,
		}
		dataBytes, _ := json.Marshal(requestData)

		msg := WsMessage[json.RawMessage]{
			Action: "join_room",
			Data:   dataBytes,
		}

		mockRM.On("CheckUserAllowedJoinRoom", userID, roomID, models.RoomTypeChannel).Return(true, nil).Once()
		mockRM.On("InitRoom", models.RoomTypeChannel, roomID).Return(&Room{}).Once()
		mockRM.On("JoinRoom", client, models.RoomTypeChannel, roomID).Once()

		handler.handleClientMessage(client, msg)

		mockRM.AssertExpectations(t)
	})

	t.Run("處理 ping 動作", func(t *testing.T) {
		handler := &webSocketHandler{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		msg := WsMessage[json.RawMessage]{
			Action: "ping",
			Data:   json.RawMessage(`{}`),
		}

		handler.handleClientMessage(client, msg)

		// 驗證收到 pong
		select {
		case <-sendCh:
			// 成功
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到 pong 訊息")
		}
	})

	t.Run("處理未知動作", func(t *testing.T) {
		handler := &webSocketHandler{}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		sendCh := make(chan []byte, 5)
		client := &Client{
			UserID:       userID,
			IsActive:     true,
			Send:         sendCh,
			RoomActivity: make(map[string]time.Time),
			Context:      ctx,
			Cancel:       cancel,
		}

		msg := WsMessage[json.RawMessage]{
			Action: "unknown_action",
			Data:   json.RawMessage(`{}`),
		}

		handler.handleClientMessage(client, msg)

		// 驗證收到錯誤訊息
		select {
		case msg := <-sendCh:
			var response WsMessage[ErrorResponse]
			err := json.Unmarshal(msg, &response)
			assert.NoError(t, err)
			assert.Equal(t, "error", response.Action)
			assert.Equal(t, "unknown_action", response.Data.OriginalAction)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("未收到錯誤訊息")
		}
	})
}

func TestHandleDMRoomCreation(t *testing.T) {
	roomObjectID := primitive.NewObjectID()
	userObjectID := primitive.NewObjectID()
	partnerObjectID := primitive.NewObjectID()

	t.Run("為當前用戶創建 DM 房間記錄", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		handler := &webSocketHandler{
			odm: mockODM,
		}

		// 模擬只找到對方的房間記錄
		dmRooms := []models.DMRoom{
			{
				RoomID:         roomObjectID,
				UserID:         partnerObjectID,
				ChatWithUserID: userObjectID,
			},
		}

		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.DMRoom)
			*arg = dmRooms
		}).Return(nil).Once()

		mockODM.On("Create", mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil).Once()

		handler.handleDMRoomCreation(roomObjectID.Hex(), userObjectID.Hex())

		mockODM.AssertExpectations(t)
	})

	t.Run("為對方創建 DM 房間記錄", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		handler := &webSocketHandler{
			odm: mockODM,
		}

		// 模擬只找到當前用戶的房間記錄
		dmRooms := []models.DMRoom{
			{
				RoomID:         roomObjectID,
				UserID:         userObjectID,
				ChatWithUserID: partnerObjectID,
			},
		}

		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.DMRoom)
			*arg = dmRooms
		}).Return(nil).Once()

		mockODM.On("Create", mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil).Once()

		handler.handleDMRoomCreation(roomObjectID.Hex(), userObjectID.Hex())

		mockODM.AssertExpectations(t)
	})

	t.Run("雙方都有房間記錄時不創建", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		handler := &webSocketHandler{
			odm: mockODM,
		}

		// 模擬雙方都有房間記錄
		dmRooms := []models.DMRoom{
			{
				RoomID:         roomObjectID,
				UserID:         userObjectID,
				ChatWithUserID: partnerObjectID,
			},
			{
				RoomID:         roomObjectID,
				UserID:         partnerObjectID,
				ChatWithUserID: userObjectID,
			},
		}

		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.DMRoom)
			*arg = dmRooms
		}).Return(nil).Once()

		// 不應該調用 Create

		handler.handleDMRoomCreation(roomObjectID.Hex(), userObjectID.Hex())

		mockODM.AssertExpectations(t)
	})

	t.Run("處理無效的房間 ID", func(t *testing.T) {
		handler := &webSocketHandler{}

		// 無效的 ObjectID
		handler.handleDMRoomCreation("invalid_id", userObjectID.Hex())

		// 測試不應該 panic，並且沒有操作資料庫
	})

	t.Run("處理資料庫查詢錯誤", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		handler := &webSocketHandler{
			odm: mockODM,
		}

		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Return(errors.New("database error")).Once()

		handler.handleDMRoomCreation(roomObjectID.Hex(), userObjectID.Hex())

		mockODM.AssertExpectations(t)
	})
}

func TestUpdateActivityWithThrottle(t *testing.T) {
	userID := primitive.NewObjectID().Hex()

	t.Run("節流閥不存在時更新活動", func(t *testing.T) {
		mockCache := new(mockCacheProvider)
		mockUS := new(mockUserService)

		handler := &webSocketHandler{
			cache:       mockCache,
			userService: mockUS,
		}

		// 模擬節流閥不存在
		mockCache.On("Get", mock.AnythingOfType("string")).Return("", errors.New("key not found")).Once()
		mockUS.On("UpdateUserActivity", userID).Return(nil).Once()
		mockCache.On("Set", mock.AnythingOfType("string"), "1", 3*time.Minute).Return(nil).Once()

		handler.updateActivityWithThrottle(userID)

		// 等待 goroutine 完成
		time.Sleep(50 * time.Millisecond)

		mockUS.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("節流閥存在時跳過更新", func(t *testing.T) {
		mockCache := new(mockCacheProvider)
		mockUS := new(mockUserService)

		handler := &webSocketHandler{
			cache:       mockCache,
			userService: mockUS,
		}

		// 模擬節流閥存在
		mockCache.On("Get", mock.AnythingOfType("string")).Return("1", nil).Once()
		// 不應該調用 UpdateUserActivity

		handler.updateActivityWithThrottle(userID)

		// 等待確認沒有異步調用
		time.Sleep(50 * time.Millisecond)

		mockCache.AssertExpectations(t)
		// mockUS 不應該被調用
	})
}
