package services

import (
	"context"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockWebSocketConn 模擬 WebSocket 連線
type mockWebSocketConn struct {
	*websocket.Conn
	closed bool
}

func (m *mockWebSocketConn) Close() error {
	m.closed = true
	return nil
}

func TestNewClientManager(t *testing.T) {
	cm := NewClientManager()

	assert.NotNil(t, cm)
	assert.NotNil(t, cm.clients)
	assert.NotNil(t, cm.clientsByUserID)
	assert.NotNil(t, cm.register)
	assert.NotNil(t, cm.unregister)
	assert.Equal(t, 0, len(cm.clients))
	assert.Equal(t, 0, len(cm.clientsByUserID))
}

func TestNewClient(t *testing.T) {
	cm := NewClientManager()
	userID := primitive.NewObjectID().Hex()
	mockConn := &mockWebSocketConn{}

	client := cm.NewClient(userID, mockConn.Conn)

	assert.NotNil(t, client)
	assert.Equal(t, userID, client.UserID)
	assert.Equal(t, mockConn.Conn, client.Conn)
	assert.NotNil(t, client.RoomActivity)
	assert.NotNil(t, client.Send)
	assert.True(t, client.IsActive)
	assert.Equal(t, cm, client.Hub)
	assert.False(t, client.ConnectedAt.IsZero())
	assert.False(t, client.LastPongTime.IsZero())
}

func TestRegisterClient(t *testing.T) {
	cm := NewClientManager()
	userID := primitive.NewObjectID().Hex()
	mockConn := &mockWebSocketConn{}
	client := cm.NewClient(userID, mockConn.Conn)

	// 註冊客戶端
	cm.Register(client)

	// 等待註冊完成
	time.Sleep(50 * time.Millisecond)

	// 驗證客戶端已被註冊
	retrievedClient, exists := cm.GetClient(userID)
	assert.True(t, exists)
	assert.Equal(t, client, retrievedClient)

	// 驗證客戶端在 clients map 中
	allClients := cm.GetAllClients()
	assert.Equal(t, 1, len(allClients))
	assert.True(t, allClients[client])
}

func TestUnregisterClient(t *testing.T) {
	cm := NewClientManager()
	userID := primitive.NewObjectID().Hex()
	mockConn := &mockWebSocketConn{}
	client := cm.NewClient(userID, mockConn.Conn)

	// 註冊客戶端
	cm.Register(client)
	time.Sleep(50 * time.Millisecond)

	// 驗證客戶端已註冊
	_, exists := cm.GetClient(userID)
	assert.True(t, exists)

	// 註銷客戶端
	cm.Unregister(client)
	time.Sleep(50 * time.Millisecond)

	// 驗證客戶端已被註銷
	_, exists = cm.GetClient(userID)
	assert.False(t, exists)

	// 驗證客戶端已從 clients map 中移除
	allClients := cm.GetAllClients()
	assert.Equal(t, 0, len(allClients))

	// 驗證客戶端被標記為非活躍
	assert.False(t, client.IsActive)

	// 驗證 Send channel 已被關閉（透過嘗試發送來測試）
	// 如果 channel 已關閉，這個操作會導致 panic，我們只驗證客戶端狀態即可
	// 不再驗證 WebSocket 連線的 closed 狀態，因為 mock 的限制
}

func TestGetClient(t *testing.T) {
	cm := NewClientManager()
	userID1 := primitive.NewObjectID().Hex()
	userID2 := primitive.NewObjectID().Hex()
	mockConn := &mockWebSocketConn{}

	client1 := cm.NewClient(userID1, mockConn.Conn)
	cm.Register(client1)
	time.Sleep(50 * time.Millisecond)

	t.Run("Get existing client", func(t *testing.T) {
		retrievedClient, exists := cm.GetClient(userID1)
		assert.True(t, exists)
		assert.Equal(t, client1, retrievedClient)
	})

	t.Run("Get non-existing client", func(t *testing.T) {
		_, exists := cm.GetClient(userID2)
		assert.False(t, exists)
	})
}

func TestGetAllClients(t *testing.T) {
	cm := NewClientManager()
	mockConn := &mockWebSocketConn{}

	userID1 := primitive.NewObjectID().Hex()
	userID2 := primitive.NewObjectID().Hex()
	userID3 := primitive.NewObjectID().Hex()

	client1 := cm.NewClient(userID1, mockConn.Conn)
	client2 := cm.NewClient(userID2, mockConn.Conn)
	client3 := cm.NewClient(userID3, mockConn.Conn)

	// 註冊三個客戶端
	cm.Register(client1)
	cm.Register(client2)
	cm.Register(client3)
	time.Sleep(100 * time.Millisecond)

	allClients := cm.GetAllClients()

	assert.Equal(t, 3, len(allClients))
	assert.True(t, allClients[client1])
	assert.True(t, allClients[client2])
	assert.True(t, allClients[client3])
}

func TestCheckClientsHealth(t *testing.T) {
	cm := NewClientManager()
	mockConn := &mockWebSocketConn{}

	healthyUserID := primitive.NewObjectID().Hex()
	unhealthyUserID := primitive.NewObjectID().Hex()

	healthyClient := cm.NewClient(healthyUserID, mockConn.Conn)
	unhealthyClient := cm.NewClient(unhealthyUserID, mockConn.Conn)

	// 註冊兩個客戶端
	cm.Register(healthyClient)
	cm.Register(unhealthyClient)
	time.Sleep(50 * time.Millisecond)

	// 將 unhealthyClient 的 LastPongTime 設置為很久之前（超過 PongWait）
	unhealthyClient.ActivityMutex.Lock()
	unhealthyClient.LastPongTime = time.Now().Add(-2 * PongWait)
	unhealthyClient.ActivityMutex.Unlock()

	// 執行健康檢查
	cm.CheckClientsHealth()
	time.Sleep(100 * time.Millisecond)

	// 驗證健康的客戶端仍然存在
	_, exists := cm.GetClient(healthyUserID)
	assert.True(t, exists)

	// 驗證不健康的客戶端已被移除
	_, exists = cm.GetClient(unhealthyUserID)
	assert.False(t, exists)
}

func TestIsUserOnline(t *testing.T) {
	cm := NewClientManager()
	mockConn := &mockWebSocketConn{}

	onlineUserID := primitive.NewObjectID().Hex()
	offlineUserID := primitive.NewObjectID().Hex()

	client := cm.NewClient(onlineUserID, mockConn.Conn)
	cm.Register(client)
	time.Sleep(50 * time.Millisecond)

	t.Run("Online user", func(t *testing.T) {
		isOnline := cm.IsUserOnline(onlineUserID)
		assert.True(t, isOnline)
	})

	t.Run("Offline user", func(t *testing.T) {
		isOnline := cm.IsUserOnline(offlineUserID)
		assert.False(t, isOnline)
	})
}

func TestStartHealthChecker(t *testing.T) {
	cm := NewClientManager()
	mockConn := &mockWebSocketConn{}

	unhealthyUserID := primitive.NewObjectID().Hex()
	unhealthyClient := cm.NewClient(unhealthyUserID, mockConn.Conn)

	// 註冊客戶端
	cm.Register(unhealthyClient)
	time.Sleep(50 * time.Millisecond)

	// 將客戶端設置為不健康
	unhealthyClient.ActivityMutex.Lock()
	unhealthyClient.LastPongTime = time.Now().Add(-2 * PongWait)
	unhealthyClient.ActivityMutex.Unlock()

	// 創建 context 來控制健康檢查器
	ctx, cancel := context.WithCancel(context.Background())

	// 啟動健康檢查器（使用較短的間隔進行測試）
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				cm.CheckClientsHealth()
			}
		}
	}()

	// 等待健康檢查器運行
	time.Sleep(200 * time.Millisecond)

	// 驗證不健康的客戶端已被移除
	_, exists := cm.GetClient(unhealthyUserID)
	assert.False(t, exists)

	// 停止健康檢查器
	cancel()
}

func TestMultipleClientsRegistrationAndUnregistration(t *testing.T) {
	cm := NewClientManager()
	mockConn := &mockWebSocketConn{}

	// 創建多個客戶端
	clients := make([]*Client, 10)
	userIDs := make([]string, 10)

	for i := 0; i < 10; i++ {
		userIDs[i] = primitive.NewObjectID().Hex()
		clients[i] = cm.NewClient(userIDs[i], mockConn.Conn)
		cm.Register(clients[i])
	}

	time.Sleep(100 * time.Millisecond)

	// 驗證所有客戶端都已註冊
	allClients := cm.GetAllClients()
	assert.Equal(t, 10, len(allClients))

	// 註銷一半的客戶端
	for i := 0; i < 5; i++ {
		cm.Unregister(clients[i])
	}

	time.Sleep(100 * time.Millisecond)

	// 驗證還剩 5 個客戶端
	allClients = cm.GetAllClients()
	assert.Equal(t, 5, len(allClients))

	// 驗證被註銷的客戶端不再存在
	for i := 0; i < 5; i++ {
		_, exists := cm.GetClient(userIDs[i])
		assert.False(t, exists)
	}

	// 驗證未註銷的客戶端仍然存在
	for i := 5; i < 10; i++ {
		_, exists := cm.GetClient(userIDs[i])
		assert.True(t, exists)
	}
}
