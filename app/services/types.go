package services

import (
	"chat_app_backend/app/models"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket 配置常數
const (
	MaxMessageSize   = 1024 * 1024         // 1MB
	WriteWait        = 10 * time.Second    // 寫入超時
	PongWait         = 60 * time.Second    // Pong 等待時間
	PingPeriod       = (PongWait * 9) / 10 // Ping 週期
	CloseGracePeriod = 10 * time.Second    // 優雅關閉等待時間
)

// WebSocket 消息結構
type WsMessage[T any] struct {
	Action string `json:"action"`
	Data   T      `json:"data"`
}

// WsStatusResponse 定義狀態回應結構
type WsStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// MessageResponse 定義聊天室消息
type MessageResponse struct {
	RoomType  models.RoomType `json:"room_type"`
	RoomID    string          `json:"room_id"`
	SenderID  string          `json:"sender_id"`
	Content   string          `json:"content"`
	Timestamp int64           `json:"timestamp"`
}

// Notification 定義通知結構
type Notification struct {
	Action      string          `json:"action"`
	RoomType    models.RoomType `json:"room_type"`
	RoomID      string          `json:"room_id"`
	Message     string          `json:"message"`
	UnreadCount int             `json:"unread_count"`
}

// Client 定義 WebSocket 客戶端
type Client struct {
	UserID        string
	Conn          *websocket.Conn
	Send          chan []byte    // 發送訊息通道
	Hub           *ClientManager // 所屬的客戶端管理器
	Subscribed    map[string]bool
	SubscribedMux sync.RWMutex
	// 房間活躍時間追蹤
	RoomActivity  map[string]time.Time // 房間ID -> 最後活躍時間
	ActivityMutex sync.RWMutex
	// 連線狀態管理
	LastPongTime time.Time // 最後收到 pong 的時間
	ConnectedAt  time.Time // 連線建立時間
	IsActive     bool      // 連線是否活躍
	LastError    error     // 最後的錯誤
	// 協程管理
	Context context.Context    // 用於控制協程生命週期
	Cancel  context.CancelFunc // 取消函數
}

// Room 定義房間結構
type Room struct {
	Key       RoomKey                          `json:"key"`  // 複合ID
	ID        string                           `json:"id"`   // channel_id or dm_room_id
	Type      models.RoomType                  `json:"type"` // channel, dm
	Clients   map[*Client]bool                 // 房間中的客戶端
	Broadcast chan *WsMessage[MessageResponse] // 房間廣播通道
	Mutex     sync.RWMutex                     // 保護 Clients
}

// RoomKey 定義房間的複合ID
type RoomKey struct {
	Type   models.RoomType
	RoomID string
}

// String 實現 RoomKey 的 String 方法
func (rk RoomKey) String() string {
	return fmt.Sprintf("%s:%s", rk.Type, rk.RoomID)
}

// Client 方法

// SendMessage 發送 JSON 訊息
func (c *Client) SendMessage(message interface{}) error {
	select {
	case c.Send <- encodeMessage(message):
		return nil
	default:
		return fmt.Errorf("客戶端發送通道已滿")
	}
}

// SendText 發送文字訊息
func (c *Client) SendText(text string) error {
	select {
	case c.Send <- []byte(text):
		return nil
	default:
		return fmt.Errorf("客戶端發送通道已滿")
	}
}

// SendError 發送錯誤訊息
func (c *Client) SendError(errorType, message string) {
	errorMsg := WsMessage[map[string]interface{}]{
		Action: "error",
		Data: map[string]interface{}{
			"error_type": errorType,
			"message":    message,
			"timestamp":  time.Now().UnixMilli(),
		},
	}
	c.SendMessage(errorMsg)
}

// Close 優雅關閉客戶端連線
func (c *Client) Close() {
	c.IsActive = false
	if c.Cancel != nil {
		c.Cancel()
	}

	// 設置寫入超時並發送關閉訊息
	c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
	c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// 等待優雅關閉
	time.AfterFunc(CloseGracePeriod, func() {
		c.Conn.Close()
	})
}

// UpdateLastSeen 更新最後活動時間
func (c *Client) UpdateLastSeen() {
	c.ActivityMutex.Lock()
	c.LastPongTime = time.Now()
	c.ActivityMutex.Unlock()
}

// IsHealthy 檢查客戶端是否健康
func (c *Client) IsHealthy() bool {
	c.ActivityMutex.RLock()
	defer c.ActivityMutex.RUnlock()

	if !c.IsActive {
		return false
	}

	// 檢查最後 pong 時間
	if time.Since(c.LastPongTime) > PongWait {
		return false
	}

	return true
}

// 輔助函數
func encodeMessage(message interface{}) []byte {
	if data, err := json.Marshal(message); err == nil {
		return data
	}
	return []byte("{}")
}
