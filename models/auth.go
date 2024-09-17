package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username    string             `json:"username" bson:"username"`
	Email       string             `json:"email" bson:"email"`
	Password    string             `json:"password,omitempty" bson:"password"`
	DisplayName string             `json:"display_name" bson:"display_name"`
	CreatedAt   int64              `json:"created_at" bson:"created_at"`
	UpdateAt    int64              `json:"update_at" bson:"update_at"`
}

type RefreshToken struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	Token     string             `json:"token" bson:"token"`
	ExpiresAt int64              `json:"expires_at" bson:"expires_at"`
	Revoked   bool               `json:"revoked" bson:"revoked"`
	CreatedAt int64              `json:"created_at" bson:"created_at"`
	UpdateAt  int64              `json:"update_at" bson:"update_at"`
	// 可以加入裝置資訊或限制使用者token數量判斷多餘token是否要刪除
}