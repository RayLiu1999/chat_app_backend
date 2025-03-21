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

	// SendMessage 發送消息到聊天室
	// HandleNewMessage(message models.Message)

	// GetServerListByUserId 獲取用戶的伺服器列表
	GetServerListByUserId(userID primitive.ObjectID) ([]models.Server, error)

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}
