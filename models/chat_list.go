package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatList struct {
	UserID         primitive.ObjectID `json:"user_id" bson:"user_id"`
	ChatWithUserID primitive.ObjectID `json:"chat_with_user_id" bson:"chat_with_user_id"`
	LastMessageID  primitive.ObjectID `json:"last_message_id" bson:"last_message_id"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdateAt       time.Time          `json:"update_at" bson:"update_at"` // 最後聊天時間(用來排序)
}
