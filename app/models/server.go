package models

import (
	"chat_app_backend/app/providers"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 伺服器
type Server struct {
	providers.BaseModel `bson:",inline"`
	Name                string             `json:"name" bson:"name"`
	ImageID             primitive.ObjectID `json:"image_id,omitempty" bson:"image_id,omitempty"` // 圖片ID
	Description         string             `json:"description" bson:"description"`
	OwnerID             primitive.ObjectID `json:"owner_id" bson:"owner_id"`
	MemberCount         int                `json:"member_count" bson:"member_count"`         // 成員總數（快取）
	IsPublic            bool               `json:"is_public" bson:"is_public"`               // 是否公開（可搜尋）
	Tags                []string           `json:"tags,omitempty" bson:"tags,omitempty"`     // 標籤（用於搜尋）
	Region              string             `json:"region,omitempty" bson:"region,omitempty"` // 地區
	MaxMembers          int                `json:"max_members" bson:"max_members"`           // 最大成員數限制
}

// 伺服器成員（獨立集合）
type ServerMember struct {
	providers.BaseModel `bson:",inline"`
	ServerID            primitive.ObjectID `json:"server_id" bson:"server_id"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	Role                string             `json:"role" bson:"role"` // "owner", "admin", "member"
	JoinedAt            time.Time          `json:"joined_at" bson:"joined_at"`
	LastActiveAt        time.Time          `json:"last_active_at" bson:"last_active_at"`
	Permissions         []string           `json:"permissions,omitempty" bson:"permissions,omitempty"` // 特殊權限
	Nickname            string             `json:"nickname,omitempty" bson:"nickname,omitempty"`       // 伺服器內暱稱
}

func (sm *ServerMember) GetCollectionName() string {
	return "server_members"
}

// 使用者與伺服器關聯
type UserServer struct {
	providers.BaseModel `bson:",inline"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	ServerID            primitive.ObjectID `json:"server_id" bson:"server_id"`
}

func (s *Server) GetCollectionName() string {
	return "servers"
}
