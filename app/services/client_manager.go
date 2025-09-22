package services

import (
	"chat_app_backend/utils"
	"context"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// ClientManager 管理客戶端的註冊和註銷
type ClientManager struct {
	clients         map[*Client]bool
	clientsByUserID map[string]*Client
	redisClient     *redis.Client
	mutex           sync.RWMutex
	register        chan *Client
	unregister      chan *Client
}

// NewClientManager 創建新的客戶端管理器
func NewClientManager(redisClient *redis.Client) *ClientManager {
	cm := &ClientManager{
		clients:         make(map[*Client]bool, 1000),
		clientsByUserID: make(map[string]*Client, 1000),
		redisClient:     redisClient,
		register:        make(chan *Client, 1000),
		unregister:      make(chan *Client, 1000),
	}

	go cm.handleRegister()
	go cm.handleUnregister()

	return cm
}

// NewClient 創建新的客戶端
func (cm *ClientManager) NewClient(userID string, ws *websocket.Conn) *Client {
	return &Client{
		UserID:        userID,
		Conn:          ws,
		Subscribed:    make(map[string]bool),
		RoomActivity:  make(map[string]time.Time),
		ActivityMutex: sync.RWMutex{},
		ConnectedAt:   time.Now(),
		LastPongTime:  time.Now(),
		IsActive:      true,
		LastError:     nil,
		Send:          make(chan []byte, 256), // 創建發送通道
		Hub:           cm,
	}
}

// Register 註冊客戶端
func (cm *ClientManager) Register(client *Client) {
	cm.register <- client
}

// Unregister 註銷客戶端
func (cm *ClientManager) Unregister(client *Client) {
	cm.unregister <- client
}

// GetClient 根據用戶ID獲取客戶端
func (cm *ClientManager) GetClient(userID string) (*Client, bool) {
	cm.mutex.RLock()
	client, exists := cm.clientsByUserID[userID]
	cm.mutex.RUnlock()
	return client, exists
}

// GetAllClients 獲取所有客戶端
func (cm *ClientManager) GetAllClients() map[*Client]bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	result := make(map[*Client]bool)
	for client, active := range cm.clients {
		result[client] = active
	}
	return result
}

// handleRegister 處理客戶端註冊
func (cm *ClientManager) handleRegister() {
	for client := range cm.register {
		utils.PrettyPrintf("正在處理用戶 %s 的註冊事件", client.UserID)
		cm.registerClient(client)
		cm.updateClientStatus(client, "online")
		utils.PrettyPrintf("用戶 %s 的註冊事件已完成", client.UserID)
	}
}

// handleUnregister 處理客戶端註銷
func (cm *ClientManager) handleUnregister() {
	for client := range cm.unregister {
		utils.PrettyPrintf("正在處理用戶 %s 的註銷事件", client.UserID)
		cm.unregisterClient(client)
		cm.updateClientStatus(client, "offline")
		utils.PrettyPrintf("用戶 %s 的註銷事件已完成", client.UserID)
	}
}

// registerClient 註冊客戶端
func (cm *ClientManager) registerClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 創建 context 來管理協程
	client.Context, client.Cancel = context.WithCancel(context.Background())

	cm.clients[client] = true
	cm.clientsByUserID[client.UserID] = client

	utils.PrettyPrintf("客戶端 %s 已註冊，當前連線數: %d", client.UserID, len(cm.clients))
}

// unregisterClient 註銷客戶端
func (cm *ClientManager) unregisterClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 標記為非活躍並取消所有相關協程
	client.IsActive = false
	if client.Cancel != nil {
		client.Cancel()
	}

	// 關閉發送通道
	if client.Send != nil {
		close(client.Send)
	}

	delete(cm.clients, client)
	delete(cm.clientsByUserID, client.UserID)

	// 清理用戶相關的 Redis 數據
	cm.redisClient.Del(context.Background(), "user:"+client.UserID+":rooms")

	// 關閉 WebSocket 連線
	if client.Conn != nil {
		client.Conn.Close()
	}

	utils.PrettyPrintf("客戶端 %s 已註銷，當前連線數: %d", client.UserID, len(cm.clients))
}

// updateClientStatus 更新客戶端狀態
func (cm *ClientManager) updateClientStatus(client *Client, status string) {
	ctx := context.Background()
	cm.redisClient.Set(ctx, "user:"+client.UserID+":status", status, 24*time.Hour)
	utils.PrettyPrintf("更新用戶 %s 的狀態：%s", client.UserID, status)
}

// CheckClientsHealth 檢查所有客戶端的健康狀態
func (cm *ClientManager) CheckClientsHealth() {
	cm.mutex.RLock()
	var unhealthyClients []*Client

	for client := range cm.clients {
		if !client.IsHealthy() {
			unhealthyClients = append(unhealthyClients, client)
		}
	}
	cm.mutex.RUnlock()

	// 移除不健康的客戶端
	for _, client := range unhealthyClients {
		utils.PrettyPrintf("客戶端 %s 健康檢查失敗，強制斷開", client.UserID)
		cm.Unregister(client)
	}
}

// StartHealthChecker 啟動健康檢查器
func (cm *ClientManager) StartHealthChecker(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // 每30秒檢查一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.CheckClientsHealth()
		}
	}
}
