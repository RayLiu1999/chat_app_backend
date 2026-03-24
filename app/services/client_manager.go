package services

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"chat_app_backend/app/providers"
	"chat_app_backend/utils"

	"github.com/gorilla/websocket"
)

// clientManager 管理客戶端的註冊和註銷
type clientManager struct {
	clients         map[*Client]bool
	clientsByUserID map[string]*Client
	mutex           sync.RWMutex
	cache           providers.CacheProvider // 用於跨實例在線狀態查詢
}

// NewClientManager 創建新的客戶端管理器
func NewClientManager(cache providers.CacheProvider) *clientManager {
	return &clientManager{
		clients:         make(map[*Client]bool, 1000),
		clientsByUserID: make(map[string]*Client, 1000),
		cache:           cache,
	}
}

// NewClient 創建新的客戶端
func (cm *clientManager) NewClient(userID string, ws *websocket.Conn) *Client {
	// #nosec G118
	ctx, cancel := context.WithCancel(context.Background())
	// 這裡不使用 defer cancel() 因為 context 需要隨 Client 存活
	// cancel 會在 Unregister 時被呼叫以釋放資源 (符合 G118 邏輯，但 gosec 可能仍會警告)
	return &Client{
		UserID:       userID,
		Conn:         ws,
		RoomActivity: make(map[string]time.Time),
		// Subscribed:    make(map[string]bool),
		ActivityMutex: sync.RWMutex{},
		ConnectedAt:   time.Now(),
		LastPongTime:  time.Now(),
		IsActive:      true,
		LastError:     nil,
		Send:          make(chan []byte, 256), // 創建發送通道
		Hub:           cm,
		Context:       ctx,
		Cancel:        cancel,
	}
}

// Register 註冊客戶端 (直接加鎖，移除 Channel)
func (cm *clientManager) Register(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.clients[client] = true
	cm.clientsByUserID[client.UserID] = client

	slog.Info("客戶端已註冊", "user_id", client.UserID, "total_connections", len(cm.clients))
}

// GetClient 根據用戶ID獲取客戶端
func (cm *clientManager) GetClient(userID string) (*Client, bool) {
	cm.mutex.RLock()
	client, exists := cm.clientsByUserID[userID]
	cm.mutex.RUnlock()
	return client, exists
}

// GetAllClients 獲取所有客戶端
func (cm *clientManager) GetAllClients() map[*Client]bool {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()
	result := make(map[*Client]bool, len(cm.clients))
	for k, v := range cm.clients {
		result[k] = v
	}
	return result
}

// Unregister 註銷客戶端 (直接加鎖，移除 Channel)
func (cm *clientManager) Unregister(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 標記為非活躍並取消所有相關協程
	client.IsActive = false
	if client.Cancel != nil {
		client.Cancel()
	}

	// ❌ 移除 close(client.Send) 以防止 Panic
	// 依賴 clientWritePump 監聽 Cancel() 訊號後自然退出

	delete(cm.clients, client)
	delete(cm.clientsByUserID, client.UserID)

	// 關閉 WebSocket 連線 (必須由 Hub 負責清理)
	if client.Conn != nil {
		if err := client.Conn.Close(); err != nil {
			slog.Warn("無法關閉客戶端 WebSocket 連線", "user_id", client.UserID, "error", err)
		}
	}

	slog.Info("客戶端已註銷", "user_id", client.UserID, "total_connections", len(cm.clients))
}

// CheckClientsHealth 檢查所有客戶端的健康狀態
func (cm *clientManager) CheckClientsHealth() {
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
		slog.Warn("客戶端健康檢查失敗，強制斷開", "user_id", client.UserID)
		cm.Unregister(client)
	}
}

// StartHealthChecker 啟動健康檢查器
func (cm *clientManager) StartHealthChecker(ctx context.Context) {
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

// IsUserOnline 檢查用戶是否在線
// 先查本機 WebSocket 連線（本實例），找不到再查 Redis 跨實例狀態
func (cm *clientManager) IsUserOnline(userID string) bool {
	if _, exists := cm.GetClient(userID); exists {
		return true
	}
	// 跨實例查詢：從 Redis 讀取在線狀態
	if cm.cache != nil {
		status, _ := cm.cache.Get(utils.UserStatusCacheKey(userID))
		return status == "online"
	}
	return false
}
