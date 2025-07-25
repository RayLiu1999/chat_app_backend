package models

import (
	"chat_app_backend/providers"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	providers.BaseModel `bson:",inline"`
	Username            string               `json:"username" bson:"username"`
	Email               string               `json:"email" bson:"email"`
	Password            string               `json:"password,omitempty" bson:"password"`
	Nickname            string               `json:"nickname" bson:"nickname"`
	Friends             []primitive.ObjectID `json:"friends" bson:"friends"`
	Picture             string               `json:"picture" bson:"picture"`
	IsOnline            bool                 `json:"is_online" bson:"is_online"`           // 在線狀態
	LastActiveAt        int64                `json:"last_active_at" bson:"last_active_at"` // 最後活動時間戳
}

// 好友
type Friend struct {
	providers.BaseModel `bson:",inline"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	FriendID            primitive.ObjectID `json:"friend_id" bson:"friend_id"`
	Status              string             `json:"status" bson:"status"` // e.g., "pending", "accepted", "blocked"
}

type RefreshToken struct {
	providers.BaseModel `bson:",inline"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	Token               string             `json:"token" bson:"token"`
	ExpiresAt           int64              `json:"expires_at" bson:"expires_at"`
	Revoked             bool               `json:"revoked" bson:"revoked"`
	// 可以加入裝置資訊或限制使用者token數量判斷多餘token是否要刪除
}

// 添加到 models/auth.go 文件中
func (u *User) GetCollectionName() string {
	return "users" // 返回集合名稱
}

func (f *Friend) GetCollectionName() string {
	return "friends"
}

func (rt *RefreshToken) GetCollectionName() string {
	return "refresh_tokens"
}
