package repositories

import (
	"chat_app_backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatRepositoryInterface interface {
	// SaveMessage 將聊天消息保存到數據庫
	SaveMessage(message models.Message) (primitive.ObjectID, error)

	// GetMessagesByRoomID 根據房間ID獲取消息
	GetMessagesByRoomID(roomID primitive.ObjectID, limit int64) ([]models.Message, error)
}

type ServerRepositoryInterface interface {
	// GetServerListByUserId 獲取用戶的伺服器列表
	GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error)

	// AddServer 新建測試用戶伺服器關聯
	AddServer(server *models.Server) (models.Server, error)
}

type UserRepositoryInterface interface {
	// GetUserById 根據用戶ID獲取用戶
	GetUserById(objectID primitive.ObjectID) (*models.User, error)
}
