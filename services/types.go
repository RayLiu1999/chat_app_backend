package services

import (
	"chat_app_backend/models"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket 消息結構
type WsMessage[T any] struct {
	Action string `json:"action"`
	Data   T      `json:"data"`
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
	WriteMutex    sync.Mutex // 保護 WebSocket 寫入操作
	Subscribed    map[string]bool
	SubscribedMux sync.RWMutex
	// 新增房間活躍時間追蹤
	RoomActivity  map[string]time.Time // 房間ID -> 最後活躍時間
	ActivityMutex sync.RWMutex
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
