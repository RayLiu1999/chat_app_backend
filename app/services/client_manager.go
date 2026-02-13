package services

import (
	"context"
	"log/slog"
	"maps"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// clientManager 管理客戶端的註冊和註銷
type clientManager struct {
	clients         map[*Client]bool
	clientsByUserID map[string]*Client
	mutex           sync.RWMutex
	register        chan *Client
	unregister      chan *Client
}

// NewClientManager 創建新的客戶端管理器
func NewClientManager() *clientManager {
	cm := &clientManager{
		clients:         make(map[*Client]bool, 1000),
		clientsByUserID: make(map[string]*Client, 1000),
		register:        make(chan *Client, 1000),
		unregister:      make(chan *Client, 1000),
	}

	go cm.handleRegister()
	go cm.handleUnregister()

	return cm
}

// NewClient 創建新的客戶端
func (cm *clientManager) NewClient(userID string, ws *websocket.Conn) *Client {
	ctx, cancel := context.WithCancel(context.Background())
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

// Register 註冊客戶端
func (cm *clientManager) Register(client *Client) {
	cm.register <- client
}

// Unregister 註銷客戶端
func (cm *clientManager) Unregister(client *Client) {
	cm.unregister <- client
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
	result := make(map[*Client]bool)
	maps.Copy(result, cm.clients)
	return result
}

// handleRegister 處理客戶端註冊
func (cm *clientManager) handleRegister() {
	for client := range cm.register {
		slog.Debug("正在處理用戶註冊事件", "user_id", client.UserID)
		cm.registerClient(client)
		slog.Debug("用戶註冊事件已完成", "user_id", client.UserID)
	}
}

// handleUnregister 處理客戶端註銷
func (cm *clientManager) handleUnregister() {
	for client := range cm.unregister {
		slog.Debug("正在處理用戶註銷事件", "user_id", client.UserID)
		cm.unregisterClient(client)
		slog.Debug("用戶註銷事件已完成", "user_id", client.UserID)
	}
}

// registerClient 註冊客戶端
func (cm *clientManager) registerClient(client *Client) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 由於 Context 已在 NewClient 中初始化，直接加入 Map
	cm.clients[client] = true
	cm.clientsByUserID[client.UserID] = client

	slog.Info("客戶端已註冊", "user_id", client.UserID, "total_connections", len(cm.clients))
}

// unregisterClient 註銷客戶端
func (cm *clientManager) unregisterClient(client *Client) {
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

	// 關閉 WebSocket 連線
	if client.Conn != nil {
		client.Conn.Close()
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

// IsUserOnline 基於 WebSocket 連線檢查用戶是否在線
func (cm *clientManager) IsUserOnline(userID string) bool {
	_, exists := cm.GetClient(userID)
	return exists
}
