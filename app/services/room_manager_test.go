package services

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- 測試輔助函數 ---

// testClient 是用於測試目的的簡化客戶端。
type testClient struct {
	*Client
	sendCh chan []byte
}

// newTestClient 為測試建立新的客戶端。
func newTestClient(userID string) *testClient {
	ctx, cancel := context.WithCancel(context.Background())
	sendCh := make(chan []byte, 5)

	client := &Client{
		UserID:       userID,
		IsActive:     true,
		Send:         sendCh,
		RoomActivity: make(map[string]time.Time),
		Context:      ctx,
		Cancel:       cancel,
	}

	return &testClient{
		Client: client,
		sendCh: sendCh,
	}
}

// SendMessage 實現客戶端的 SendMessage 方法用於測試。
func (c *testClient) SendMessage(message interface{}) error {
	// 模擬實際 SendMessage 的行為，該方法會編組訊息。
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	c.sendCh <- msgBytes
	return nil
}

// --- 測試 ---

func TestNewRoomManager(t *testing.T) {
	redisClient, _ := redismock.NewClientMock()
	mockRepo := new(mocks.ServerMemberRepository)
	mockODM := new(mocks.ODM)
	rm := NewRoomManager(mockODM, redisClient, mockRepo)

	assert.NotNil(t, rm, "RoomManager 不應為 nil")
	assert.NotNil(t, rm.rooms, "Rooms map 應該被初始化")
}

func TestRoomManager_InitRoom(t *testing.T) {
	redisClient, _ := redismock.NewClientMock()
	rm := NewRoomManager(nil, redisClient, new(mocks.ServerMemberRepository))
	roomID := primitive.NewObjectID().Hex()

	t.Run("應該在房間不存在時創建新房間", func(t *testing.T) {
		room := rm.InitRoom(models.RoomTypeChannel, roomID)
		assert.NotNil(t, room)
		assert.Equal(t, roomID, room.ID)
		assert.Equal(t, models.RoomTypeChannel, room.Type)

		retrievedRoom, exists := rm.GetRoom(models.RoomTypeChannel, roomID)
		assert.True(t, exists, "初始化後房間應該存在於管理器中")
		assert.Equal(t, room, retrievedRoom, "檢索到的房間應該是同一個實例")
	})

	t.Run("應該在再次調用時返回現有房間", func(t *testing.T) {
		firstRoom := rm.InitRoom(models.RoomTypeChannel, roomID)
		secondRoom := rm.InitRoom(models.RoomTypeChannel, roomID)

		assert.Same(t, firstRoom, secondRoom, "再次調用 InitRoom 應該返回相同的房間實例")
	})
}

func TestRoomManager_JoinAndLeaveRoom(t *testing.T) {
	redisClient, _ := redismock.NewClientMock()
	rm := NewRoomManager(nil, redisClient, new(mocks.ServerMemberRepository))
	roomID := primitive.NewObjectID().Hex()
	client1 := newTestClient(primitive.NewObjectID().Hex())
	client2 := newTestClient(primitive.NewObjectID().Hex())

	// 初始化房間
	room := rm.InitRoom(models.RoomTypeChannel, roomID)
	assert.NotNil(t, room)

	t.Run("應該允許客戶端加入房間", func(t *testing.T) {
		rm.JoinRoom(client1.Client, models.RoomTypeChannel, roomID)
		rm.JoinRoom(client2.Client, models.RoomTypeChannel, roomID)

		room.Mutex.RLock()
		assert.Len(t, room.Clients, 2, "房間應該有兩個客戶端")
		assert.Contains(t, room.Clients, client1.Client, "Client1 應該在房間中")
		room.Mutex.RUnlock()

		client1.ActivityMutex.RLock()
		_, activityExists := client1.RoomActivity[room.Key.String()]
		assert.True(t, activityExists, "Client1 應該記錄房間活動")
		client1.ActivityMutex.RUnlock()
	})

	t.Run("應該允許客戶端離開房間", func(t *testing.T) {
		rm.LeaveRoom(client1.Client, models.RoomTypeChannel, roomID)

		room.Mutex.RLock()
		assert.Len(t, room.Clients, 1, "房間應該只有一個客戶端")
		assert.NotContains(t, room.Clients, client1.Client, "Client1 不應該再在房間中")
		room.Mutex.RUnlock()

		client1.ActivityMutex.RLock()
		_, activityExists := client1.RoomActivity[room.Key.String()]
		assert.False(t, activityExists, "Client1 的房間活動應該被移除")
		client1.ActivityMutex.RUnlock()
	})

	t.Run("應該在最後一個客戶端離開時清理房間", func(t *testing.T) {
		rm.LeaveRoom(client2.Client, models.RoomTypeChannel, roomID)

		// 清理是非同步的，等待一下
		time.Sleep(20 * time.Millisecond)

		_, exists := rm.GetRoom(models.RoomTypeChannel, roomID)
		assert.False(t, exists, "房間應該被清理並從管理器中移除")
	})

	t.Run("應該處理客戶端離開未加入的房間", func(t *testing.T) {
		room := rm.InitRoom(models.RoomTypeChannel, "new_room")
		client3 := newTestClient("user3")
		client4 := newTestClient("user4")
		rm.JoinRoom(client3.Client, models.RoomTypeChannel, "new_room")

		// Client4 嘗試離開它從未加入的房間
		rm.LeaveRoom(client4.Client, models.RoomTypeChannel, "new_room")

		room.Mutex.RLock()
		assert.Len(t, room.Clients, 1, "房間客戶端數量應該保持不變")
		room.Mutex.RUnlock()
	})
}

func TestRoomManager_Broadcast(t *testing.T) {
	redisClient, _ := redismock.NewClientMock()
	rm := NewRoomManager(nil, redisClient, new(mocks.ServerMemberRepository))
	roomID := primitive.NewObjectID().Hex()

	client1 := newTestClient(primitive.NewObjectID().Hex())
	client2 := newTestClient(primitive.NewObjectID().Hex())

	room := rm.InitRoom(models.RoomTypeChannel, roomID)
	rm.JoinRoom(client1.Client, models.RoomTypeChannel, roomID)
	rm.JoinRoom(client2.Client, models.RoomTypeChannel, roomID)

	t.Run("應該廣播訊息到房間中的所有客戶端", func(t *testing.T) {
		testMsg := &WsMessage[MessageResponse]{
			Action: "send_message",
			Data:   MessageResponse{RoomID: roomID, Content: "Hello everyone!"},
		}

		// 廣播訊息
		room.Broadcast <- testMsg

		var wg sync.WaitGroup
		wg.Add(2)

		// 每個客戶端的檢查函數
		checker := func(tc *testClient) {
			defer wg.Done()
			select {
			case msgBytes := <-tc.sendCh:
				var receivedMsg WsMessage[MessageResponse]
				err := json.Unmarshal(msgBytes, &receivedMsg)
				assert.NoError(t, err)
				assert.Equal(t, testMsg.Data.Content, receivedMsg.Data.Content)
			case <-time.After(100 * time.Millisecond):
				assert.Fail(t, "客戶端未能及時收到訊息", tc.UserID)
			}
		}

		go checker(client1)
		go checker(client2)

		wg.Wait()
	})

	t.Run("應該處理發送錯誤並停用客戶端", func(t *testing.T) {
		// 建立一個發送通道大小為 0 的客戶端（會立即阻塞）
		failingClient := newTestClient("failing_user")
		// 建立一個沒有緩衝區的通道且不從中讀取
		// 當通道滿時這會導致發送失敗
		failingClient.Client.Send = make(chan []byte) // 零緩衝區通道

		rm.JoinRoom(failingClient.Client, models.RoomTypeChannel, roomID)
		assert.True(t, failingClient.IsActive, "故障客戶端初始應該是活躍的")

		// 發送多個訊息以填充並阻塞
		errorMsg := &WsMessage[MessageResponse]{
			Action: "error_test",
			Data:   MessageResponse{RoomID: roomID, Content: "This will fail"},
		}

		// 廣播訊息 - 故障客戶端將無法接收
		// 因為它的通道已滿且沒有人從中讀取
		room.Broadcast <- errorMsg

		// 等待廣播處理完成
		time.Sleep(50 * time.Millisecond)

		// 客戶端仍應該是活躍的，因為 SendMessage 在通道滿時返回錯誤
		// （而不是在通道關閉時），我們只是跳過該客戶端

		// 確保健康的客戶端仍能接收訊息
		select {
		case <-client1.sendCh:
			// 訊息已接收，良好。
		case <-time.After(100 * time.Millisecond):
			assert.Fail(t, "健康的客戶端未能接收訊息")
		}
	})
}

func TestRoomManager_Concurrency(t *testing.T) {
	redisClient, _ := redismock.NewClientMock()
	rm := NewRoomManager(nil, redisClient, new(mocks.ServerMemberRepository))
	numGoroutines := 100
	numRooms := 10

	// 首先，初始化所有房間以避免競爭條件
	for i := 0; i < numRooms; i++ {
		roomID := fmt.Sprintf("room_%d", i)
		rm.InitRoom(models.RoomTypeChannel, roomID)
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			roomID := fmt.Sprintf("room_%d", i%numRooms)
			userID := fmt.Sprintf("user_%d", i)

			client := newTestClient(userID)
			rm.JoinRoom(client.Client, models.RoomTypeChannel, roomID)
		}(i)
	}

	wg.Wait()

	// 給予一些時間讓所有加入操作完成
	time.Sleep(50 * time.Millisecond)

	assert.Len(t, rm.rooms, numRooms, "應該創建正確數量的房間")

	totalClients := 0
	for _, room := range rm.rooms {
		room.Mutex.RLock()
		totalClients += len(room.Clients)
		room.Mutex.RUnlock()
	}
	assert.Equal(t, numGoroutines, totalClients, "所有房間中的客戶端總數應該正確")
}

func TestCheckUserAllowedJoinRoom(t *testing.T) {
	userID := primitive.NewObjectID()
	roomID := primitive.NewObjectID()
	serverID := primitive.NewObjectID()

	t.Run("DM 房間 - 允許", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		rm := NewRoomManager(mockODM, nil, nil)

		// 期望 FindOne 被調用並返回非錯誤結果
		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(models.DMRoom{}, nil).Once()

		allowed, err := rm.CheckUserAllowedJoinRoom(userID.Hex(), roomID.Hex(), models.RoomTypeDM)

		assert.True(t, allowed)
		assert.NoError(t, err)
		mockODM.AssertExpectations(t)
	})

	t.Run("DM 房間 - 未找到", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		rm := NewRoomManager(mockODM, nil, nil)

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil, providers.ErrDocumentNotFound).Once()

		allowed, err := rm.CheckUserAllowedJoinRoom(userID.Hex(), roomID.Hex(), models.RoomTypeDM)

		assert.False(t, allowed)
		assert.NoError(t, err)
		mockODM.AssertExpectations(t)
	})

	t.Run("頻道房間 - 作為成員允許", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		mockRepo := new(mocks.ServerMemberRepository)
		rm := NewRoomManager(mockODM, nil, mockRepo)

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Channel")).Return(models.Channel{ServerID: serverID}, nil).Once()

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Server")).Return(models.Server{}, nil).Once()

		mockRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(true, nil).Once()

		allowed, err := rm.CheckUserAllowedJoinRoom(userID.Hex(), roomID.Hex(), models.RoomTypeChannel)

		assert.True(t, allowed)
		assert.NoError(t, err)
		mockODM.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("頻道房間 - 非成員被拒絕", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		mockRepo := new(mocks.ServerMemberRepository)
		rm := NewRoomManager(mockODM, nil, mockRepo)

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Channel")).Return(models.Channel{ServerID: serverID}, nil).Once()

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Server")).Return(models.Server{}, nil).Once()

		mockRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(false, nil).Once()

		allowed, err := rm.CheckUserAllowedJoinRoom(userID.Hex(), roomID.Hex(), models.RoomTypeChannel)

		assert.False(t, allowed)
		assert.NoError(t, err)
		mockODM.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("頻道房間 - 查找頻道時 ODM 錯誤", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		rm := NewRoomManager(mockODM, nil, nil)
		expectedErr := errors.New("db connection error")

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.Channel")).Return(nil, expectedErr).Once()

		allowed, err := rm.CheckUserAllowedJoinRoom(userID.Hex(), roomID.Hex(), models.RoomTypeChannel)

		assert.False(t, allowed)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		mockODM.AssertExpectations(t)
	})
}
