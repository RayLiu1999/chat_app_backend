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

	// GetChatListByUserID 獲取用戶的聊天列表
	GetChatListByUserID(userID primitive.ObjectID, includeDeleted bool) ([]models.Chat, error)

	// UpdateChatListDeleteStatus 更新聊天列表的刪除狀態
	UpdateChatListDeleteStatus(userID, chatWithUserID primitive.ObjectID, isDeleted bool) error

	// SaveOrUpdateChat 保存或更新聊天列表
	SaveOrUpdateChat(chat models.Chat) (models.Chat, error)
}

type ServerRepositoryInterface interface {
	// GetServerListByUserId 獲取用戶的伺服器列表
	GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error)

	// CreateServer 新建測試用戶伺服器關聯
	CreateServer(server *models.Server) (models.Server, error)
}

type UserRepositoryInterface interface {
	// GetUserById 根據用戶ID獲取用戶
	GetUserById(objectID primitive.ObjectID) (*models.User, error)

	// GetUserListByIds 根據用戶ID陣列獲取用戶
	GetUserListByIds(objectIds []primitive.ObjectID) ([]models.User, error)

	// CheckUsernameExists 檢查用戶名是否已存在
	CheckUsernameExists(username string) (bool, error)

	// CheckEmailExists 檢查電子郵件是否已存在
	CheckEmailExists(email string) (bool, error)

	// CreateUser 創建用戶
	CreateUser(user models.User) error
}

type FriendRepositoryInterface interface {
	// GetFriendById 根據用戶ID獲取用戶
	GetFriendById(objectID primitive.ObjectID) (*models.Friend, error)
}
