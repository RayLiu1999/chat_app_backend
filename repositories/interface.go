package repositories

import (
	"chat_app_backend/models"
)

type ChatRepositoryInterface interface {
	// SaveMessage 將聊天消息保存到數據庫
	SaveMessage(message models.Message) (string, error)

	// GetMessagesByRoomID 根據房間ID獲取消息
	GetMessagesByRoomID(roomID string, limit int64) ([]models.Message, error)

	// GetDMRoomListByUserID 獲取用戶的聊天列表
	GetDMRoomListByUserID(userID string, includeNotVisible bool) ([]models.DMRoom, error)

	// UpdateDMRoom 更新聊天列表的刪除狀態
	UpdateDMRoom(userID string, chatWithUserID string, IsHidden bool) error

	// SaveOrUpdateDMRoom 保存或更新聊天列表
	SaveOrUpdateDMRoom(chat models.DMRoom) (models.DMRoom, error)
}

type ServerRepositoryInterface interface {
	// GetServerListByUserId 獲取用戶的伺服器列表
	GetServerListByUserId(userID string) ([]models.Server, error)

	// CreateServer 新建測試用戶伺服器關聯
	CreateServer(server *models.Server) (models.Server, error)
}

type UserRepositoryInterface interface {
	// GetUserById 根據用戶ID獲取用戶
	GetUserById(userID string) (*models.User, error)

	// GetUserListByIds 根據用戶ID陣列獲取用戶
	GetUserListByIds(userIds []string) ([]models.User, error)

	// CheckUsernameExists 檢查用戶名是否已存在
	CheckUsernameExists(username string) (bool, error)

	// CheckEmailExists 檢查電子郵件是否已存在
	CheckEmailExists(email string) (bool, error)

	// CreateUser 創建用戶
	CreateUser(user models.User) error
}

type FriendRepositoryInterface interface {
	// GetFriendById 根據用戶ID獲取用戶
	GetFriendById(userID string) (*models.Friend, error)
}
