package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 成員結構
type Member struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	UserName string             `json:"user_name" bson:"user_name"`
	Nickname string             `json:"nickname" bson:"nickname"`
	JoinedAt time.Time          `json:"joined_at" bson:"joined_at"`
	// RoleID       primitive.ObjectID `json:"role_id,omitempty" bson:"role_id,omitempty"`
	LastActiveAt time.Time `json:"last_active_at" bson:"last_active_at"`
}

// 聊天室(頻道或私聊)
type Room struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name,omitempty" bson:"name,omitempty"` // 頻道名稱或對話名稱（私聊可選）
	Type        string             `json:"type" bson:"type"`                     // "channel" 或 "dm"
	ChannelType string             `json:"channel_type" bson:"channel_type"`     // "text" 或 "voice"
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// 聊天室與使用者關聯
type RoomParticipants struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	RoomID    primitive.ObjectID `json:"room_id" bson:"room_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// 訊息
type Message struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Type       string             `json:"type" bson:"type"` // "channel" or "dm"
	Text       string             `json:"text" bson:"text"`
	SenderID   primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	ReceiverID primitive.ObjectID `json:"receiver_id" bson:"receiver_id"`
	RoomID     primitive.ObjectID `json:"room_id" bson:"room_id"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// 聊天紀錄
type Chat struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	ChatWithUserID primitive.ObjectID `json:"chat_with_user_id" bson:"chat_with_user_id"`
	IsDeleted      bool               `json:"is_deleted" bson:"is_deleted"` // 標記記錄是否被「刪除」（實際上只是隱藏）
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"` // 最後聊天時間(用來排序)
}
