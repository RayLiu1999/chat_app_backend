package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 伺服器
type Server struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Picture     string             `json:"picture" bson:"picture"`
	Description string             `json:"description" bson:"description"`
	OwnerID     primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt    time.Time          `json:"update_at" bson:"update_at"`
}

// 頻道
type Channel struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
}

// 訊息
type Message struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Content    string             `json:"content" bson:"content"`
	SenderID   primitive.ObjectID `json:"sender_id" bson:"sender_id"`
	ReceiverID primitive.ObjectID `json:"receiver_id" bson:"receiver_id"`
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

// 使用者與頻道關聯
// type UserChannel struct {
// 	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
// 	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
// 	ChannelID primitive.ObjectID `json:"channel_id" bson:"channel_id"`
// 	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
// 	UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
// }

// 好友
type Friend struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	FriendID  primitive.ObjectID `json:"friend_id" bson:"friend_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
}

// 好友請求
type FriendRequest struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	InviterID primitive.ObjectID `json:"inviter_id" bson:"inviter_id"`
	InviteeID primitive.ObjectID `json:"invitee_id" bson:"invitee_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt  time.Time          `json:"update_at" bson:"update_at"`
}
