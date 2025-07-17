package models

import (
	"chat_app_backend/providers"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Channel struct {
	providers.BaseModel `bson:",inline"`
	Name                string             `json:"name" bson:"name"`
	ServerID            primitive.ObjectID `json:"server_id" bson:"server_id"`
	CategoryID          primitive.ObjectID `json:"category_id" bson:"category_id"`
	Type                string             `json:"type" bson:"type"` // "text" or "voice"
	// OwnerID             primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	// Members             []Member           `json:"members" bson:"members"` // 頻道成員
}

func (c *Channel) GetCollectionName() string {
	return "channels"
}

func (c *Channel) GetID() primitive.ObjectID {
	return c.ID
}

func (c *Channel) SetID(id primitive.ObjectID) {
	c.ID = id
}

type ChannelCategory struct {
	providers.BaseModel `bson:",inline"`
	Name                string             `json:"name" bson:"name"`
	ServerID            primitive.ObjectID `json:"server_id" bson:"server_id"`
}
