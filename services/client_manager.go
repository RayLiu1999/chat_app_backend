package services

import (
	"chat_app_backend/utils"
	"context"
	"sync"
	"time"

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
		utils.PrettyPrintf("Handling Register event for user %s", client.UserID)
		cm.registerClient(client)
		cm.updateClientStatus(client, "online")
		utils.PrettyPrintf("Register event completed for user %s", client.UserID)
	}
}

// handleUnregister 處理客戶端註銷
func (cm *ClientManager) handleUnregister() {
	for client := range cm.unregister {
		utils.PrettyPrintf("Handling Unregister event for user %s", client.UserID)
		cm.unregisterClient(client)
		cm.updateClientStatus(client, "offline")
		utils.PrettyPrintf("Unregister event completed for user %s", client.UserID)
	}
}

// registerClient 註冊客戶端
func (cm *ClientManager) registerClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()
	client.Subscribed = make(map[string]bool)
	client.RoomActivity = make(map[string]time.Time)
	cm.clients[client] = true
	cm.clientsByUserID[client.UserID] = client
}

// unregisterClient 註銷客戶端
func (cm *ClientManager) unregisterClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.clients, client)
	delete(cm.clientsByUserID, client.UserID)

	// 清理用戶相關的 Redis 數據
	cm.redisClient.Del(context.Background(), "user:"+client.UserID+":rooms")

	// 使用互斥鎖保護 WebSocket 連接的關閉操作
	client.WriteMutex.Lock()
	client.Conn.Close()
	client.WriteMutex.Unlock()
}

// updateClientStatus 更新客戶端狀態
func (cm *ClientManager) updateClientStatus(client *Client, status string) {
	ctx := context.Background()
	cm.redisClient.Set(ctx, "user:"+client.UserID+":status", status, 24*time.Hour)
	utils.PrettyPrintf("Update status for user %s: %s", client.UserID, status)
}
