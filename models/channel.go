package models

import (
	"chat_app_backend/providers"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Channel 頻道模型
// 注意：頻道權限繼承自伺服器權限，不單獨管理成員
// 只要是伺服器成員就可以訪問該伺服器的所有頻道
type Channel struct {
	providers.BaseModel `bson:",inline"`
	Name                string             `json:"name" bson:"name"`
	ServerID            primitive.ObjectID `json:"server_id" bson:"server_id"` // 所屬伺服器
	CategoryID          primitive.ObjectID `json:"category_id" bson:"category_id"`
	Type                string             `json:"type" bson:"type"` // "text" or "voice"
}

func (c *Channel) GetCollectionName() string {
	return "channels"
}

type ChannelCategory struct {
	providers.BaseModel `bson:",inline"`
	Name                string             `json:"name" bson:"name"`
	ServerID            primitive.ObjectID `json:"server_id" bson:"server_id"`
	CategoryType        string             `json:"category_type" bson:"category_type"` // "text", "voice", "custom"
	Position            int                `json:"position" bson:"position"`           // 排序位置
}

func (cc *ChannelCategory) GetCollectionName() string {
	return "channel_categories"
}
