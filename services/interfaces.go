package services

import (
	"chat_app_backend/models"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserServiceInterface 定義了用戶服務的接口
// 所有與用戶相關的業務邏輯方法都應該在這裡聲明
type UserServiceInterface interface {
	// GetUserById 根據ID獲取用戶信息ch
	GetUserById(objectID primitive.ObjectID) (*models.User, error)

	// 未來可能添加的其他方法
	// CreateUser(user *models.User) (primitive.ObjectID, error)
	// UpdateUser(userID primitive.ObjectID, updates map[string]interface{}) error
	// DeleteUser(userID primitive.ObjectID) error
	// VerifyUserCredentials(username, password string) (*models.User, error)
	// ... 其他方法 ...
}

// ChatServiceInterface 定義了聊天服務的接口
// 所有與聊天相關的業務邏輯方法都應該在這裡声明
type ChatServiceInterface interface {
	// HandleConnection 處理 WebSocket 連接
	HandleConnection(userID primitive.ObjectID, ws *websocket.Conn)

	// SaveMessage 儲存聊天訊息
	SaveMessage(message Message)

	// GetChatListByUserID 獲取用戶的聊天列表
	GetChatListByUserID(userID primitive.ObjectID, includeDeleted bool) ([]models.Chat, error)

	// UpdateChatListDeleteStatus 更新聊天列表的刪除狀態
	UpdateChatListDeleteStatus(userID, chatWithUserID primitive.ObjectID, isDeleted bool) error

	// SaveChat 保存聊天列表
	SaveChat(chatList models.Chat) (models.ChatResponse, error)

	// GetChatResponseList 獲取聊天列表response
	GetChatResponseList(userID primitive.ObjectID, includeDeleted bool) ([]models.ChatResponse, error)

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}

// ServerServiceInterface 定義了伺服器服務的接口
// 所有與伺服器相關的業務邏輯方法都應該在這裡声明
type ServerServiceInterface interface {
	// GetServerListByUserId 獲取用戶的伺服器列表
	GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error)

	// CreateServer 新建測試用戶伺服器關聯
	CreateServer(server *models.Server) (models.Server, error)

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}

type FriendServiceInterface interface {
	// GetFriendById 根據ID獲取好友信息
	GetFriendById(objectID primitive.ObjectID) (*models.Friend, error)
}
