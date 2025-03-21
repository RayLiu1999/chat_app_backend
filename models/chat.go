package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 伺服器
type Server struct {
	ID          primitive.ObjectID   `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string               `json:"name" bson:"name"`
	Picture     string               `json:"picture" bson:"picture"`
	Description string               `json:"description" bson:"description"`
	OwnerID     primitive.ObjectID   `json:"owner_id" bson:"owner_id"`
	Rooms       []primitive.ObjectID `json:"rooms" bson:"rooms"`
	Members     []Member             `json:"members" bson:"members"` // 伺服器成員
	CreatedAt   time.Time            `json:"created_at" bson:"created_at"`
	UpdateAt    time.Time            `json:"update_at" bson:"update_at"`
}

type Member struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	JoinedAt time.Time          `json:"joined_at" bson:"joined_at"`
	// RoleID       primitive.ObjectID `json:"role_id,omitempty" bson:"role_id,omitempty"`
	Nickname     string    `json:"nickname" bson:"nickname"`
	LastActiveAt time.Time `json:"last_active_at" bson:"last_active_at"`
}

// 聊天室(頻道或私聊)
type Room struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"` // 頻道名稱或對話名稱（私聊可選）
	Type        string             `json:"type" bson:"type"`                     // "channel" 或 "dm"
	channelType string             `json:"channel_type" bson:"channel_type"`     // "text" 或 "voice"
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// 訊息
type Message struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Type       string             `json:"type" bson:"type"` // "channel" or "dm"
	Content    string             `json:"content" bson:"content"`
	SenderID   primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	ReceiverID primitive.ObjectID `json:"receiver_id" bson:"receiver_id"`
	RoomID     primitive.ObjectID `json:"room_id" bson:"room_id"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt   time.Time          `json:"update_at" bson:"update_at"`
}

// 使用者與伺服器關聯
type UserServer struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	ServerID  primitive.ObjectID `json:"server_id" bson:"server_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
}

// 聊天室與使用者關聯
type RoomParticipants struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	RoomID    primitive.ObjectID `json:"room_id" bson:"room_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
}
