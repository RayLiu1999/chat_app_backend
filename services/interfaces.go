package services

import (
	"chat_app_backend/models"
	"chat_app_backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// UserServiceInterface 定義了用戶服務的接口
// 所有與用戶相關的業務邏輯方法都應該在這裡聲明
type UserServiceInterface interface {
	// GetUserById 根據ID獲取用戶信息ch
	GetUserResponseById(userID string) (*models.UserResponse, error)

	// RegisterUser 註冊新用戶
	RegisterUser(user models.User) *utils.AppError

	// Login 處理用戶登入
	Login(loginUser models.User) (*models.LoginResponse, *utils.AppError)

	// Logout 處理用戶登出
	Logout(c *gin.Context) *utils.AppError

	// RefreshToken 刷新令牌
	RefreshToken(refreshToken string) (string, *utils.AppError)

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
	// HandleWebSocket 處理 WebSocket 連接
	HandleWebSocket(ws *websocket.Conn, userID string)

	// UpdateDMRoom 更新聊天列表
	// UpdateDMRoom(userID, chatWithUserID primitive.ObjectID, IsHidden bool) error

	// CreateDMRoom 保存聊天列表
	// CreateDMRoom(chatList models.DMRoom) (models.DMRoomResponse, error)

	// GetDMRoomResponseList 獲取聊天列表response
	GetDMRoomResponseList(userID string, includeNotVisible bool) ([]models.DMRoomResponse, error)

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}

// ServerServiceInterface 定義了伺服器服務的接口
// 所有與伺服器相關的業務邏輯方法都應該在這裡声明
type ServerServiceInterface interface {
	// GetServerListByUserId 獲取用戶的伺服器列表
	GetServerListByUserId(userID string) ([]models.Server, error)

	// CreateServer 新建測試用戶伺服器關聯
	CreateServer(server *models.Server) (models.Server, error)

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}

type FriendServiceInterface interface {
	// GetFriendById 根據ID獲取好友信息
	GetFriendById(userID string) (*models.Friend, error)
}
