package models

import (
	"chat_app_backend/app/providers"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoomType string

const (
	RoomTypeChannel RoomType = "channel"
	RoomTypeDM      RoomType = "dm"
)

// 成員結構
type Member struct {
	UserID   primitive.ObjectID `json:"user_id" bson:"user_id"`
	JoinedAt time.Time          `json:"joined_at" bson:"joined_at"`
	// RoleID       primitive.ObjectID `json:"role_id,omitempty" bson:"role_id,omitempty"`
	LastActiveAt time.Time `json:"last_active_at" bson:"last_active_at"`
}

// 聊天室(頻道或私聊)
// type Room struct {
// 	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
// 	Name      string             `json:"name,omitempty" bson:"name,omitempty"` // 頻道名稱或對話名稱（私聊可選）
// 	Type      string             `json:"type" bson:"type"`                     // "channel" 或 "dm"
// 	DMRoomID  primitive.ObjectID `json:"dm_room_id" bson:"dm_room_id"`
// 	ChannelID primitive.ObjectID `json:"channel_id" bson:"channel_id"`
// 	// Members   []Member           `json:"members" bson:"members"`
// 	CreatedAt time.Time `json:"created_at" bson:"created_at"`
// 	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
// }

// 聊天室與使用者關聯
// type RoomParticipants struct {
// 	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
// 	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
// 	RoomID    primitive.ObjectID `json:"room_id" bson:"room_id"`
// 	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
// 	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
// }

// 訊息
type Message struct {
	providers.BaseModel `bson:",inline"`
	RoomType            RoomType           `json:"room_type" bson:"room_type"` // "channel" or "dm"
	Content             string             `json:"content" bson:"content"`
	SenderID            primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	RoomID              primitive.ObjectID `json:"room_id" bson:"room_id"`
}

// GetCollectionName 返回Message的集合名稱
func (m *Message) GetCollectionName() string {
	return "messages"
}

// 私聊房間
type DMRoom struct {
	providers.BaseModel `bson:",inline"`
	RoomID              primitive.ObjectID `json:"room_id" bson:"room_id"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	ChatWithUserID      primitive.ObjectID `json:"chat_with_user_id" bson:"chat_with_user_id"`
	IsHidden            bool               `json:"is_hidden" bson:"is_hidden"` // 是否隱藏
}

// GetCollectionName 返回DMRoom的集合名稱
func (dm *DMRoom) GetCollectionName() string {
	return "dm_rooms"
}

// 房間已讀時間
type RoomReads struct {
	providers.BaseModel `bson:",inline"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	RoomID              primitive.ObjectID `json:"room_id" bson:"room_id"`
	LastReadAt          time.Time          `json:"last_read_at" bson:"last_read_at"`
}

// GetCollectionName 返回RoomReads的集合名稱
func (rr *RoomReads) GetCollectionName() string {
	return "room_reads"
}
