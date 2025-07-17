package models

import (
	"chat_app_backend/providers"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 伺服器
type Server struct {
	providers.BaseModel `bson:",inline"`
	Name                string             `json:"name" bson:"name"`
	Picture             string             `json:"picture" bson:"picture"`
	Description         string             `json:"description" bson:"description"`
	OwnerID             primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	Members             []Member           `json:"members" bson:"members"` // 伺服器成員
}

// 使用者與伺服器關聯
type UserServer struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	ServerID  primitive.ObjectID `json:"server_id" bson:"server_id"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

func (s *Server) GetCollectionName() string {
	return "servers"
}

func (s *Server) GetID() primitive.ObjectID {
	return s.ID
}

func (s *Server) SetID(id primitive.ObjectID) {
	s.ID = id
}
