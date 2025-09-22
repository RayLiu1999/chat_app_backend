package services

import (
	"chat_app_backend/app/models"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// UserServiceInterface 定義了用戶服務的接口
// 所有與用戶相關的業務邏輯方法都應該在這裡聲明
type UserServiceInterface interface {
	// GetUserById 根據ID獲取用戶信息ch
	GetUserResponseById(userID string) (*models.UserResponse, error)

	// RegisterUser 註冊新用戶
	RegisterUser(user models.User) *models.MessageOptions

	// Login 處理用戶登入
	Login(loginUser models.User) (*models.LoginResponse, *models.MessageOptions)

	// Logout 處理用戶登出
	Logout(c *gin.Context) *models.MessageOptions

	// RefreshToken 刷新令牌
	RefreshToken(refreshToken string) (*models.RefreshTokenResponse, *models.MessageOptions)

	// SetUserOnline 設置用戶為在線狀態
	SetUserOnline(userID string) error

	// SetUserOffline 設置用戶為離線狀態
	SetUserOffline(userID string) error

	// UpdateUserActivity 更新用戶活動時間
	UpdateUserActivity(userID string) error

	// CheckAndSetOfflineUsers 檢查並設置離線用戶
	CheckAndSetOfflineUsers(offlineThresholdMinutes int) error

	// IsUserOnlineByWebSocket 基於 WebSocket 連線檢查用戶是否在線
	IsUserOnlineByWebSocket(userID string) bool

	// GetUserProfile 獲取用戶個人資料
	GetUserProfile(userID string) (*models.UserProfileResponse, error)

	// UpdateUserProfile 更新用戶基本資料
	UpdateUserProfile(userID string, updates map[string]interface{}) error

	// UploadUserImage 上傳用戶頭像或橫幅
	UploadUserImage(userID string, file multipart.File, header *multipart.FileHeader, imageType string) (*models.UserImageResponse, error)

	// DeleteUserAvatar 刪除用戶頭像
	DeleteUserAvatar(userID string) error

	// DeleteUserBanner 刪除用戶橫幅
	DeleteUserBanner(userID string) error

	// UpdateUserPassword 更新用戶密碼
	UpdateUserPassword(userID string, newPassword string) error

	// GetTwoFactorStatus 獲取兩步驟驗證狀態
	GetTwoFactorStatus(userID string) (*models.TwoFactorStatusResponse, error)

	// UpdateTwoFactorStatus 啟用/停用兩步驟驗證
	UpdateTwoFactorStatus(userID string, enabled bool) error

	// DeactivateAccount 停用帳號
	DeactivateAccount(userID string) error

	// DeleteAccount 刪除帳號
	DeleteAccount(userID string) error

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

	// GetDMRoomResponseList 獲取聊天列表response
	GetDMRoomResponseList(userID string, includeNotVisible bool) ([]models.DMRoomResponse, *models.MessageOptions)

	// UpdateDMRoom 更新聊天房間狀態
	UpdateDMRoom(userID string, roomID string, isHidden bool) *models.MessageOptions

	// CreateDMRoom 創建私聊房間
	CreateDMRoom(userID string, chatWithUserID string) (*models.DMRoomResponse, *models.MessageOptions)

	// GetDMMessages 獲取私聊訊息
	GetDMMessages(userID string, roomID string, before string, after string, limit string) ([]models.MessageResponse, *models.MessageOptions)

	// GetChannelMessages 獲取頻道訊息
	GetChannelMessages(userID string, channelID string, before string, after string, limit string) ([]models.MessageResponse, *models.MessageOptions)

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}

// ServerServiceInterface 定義了伺服器服務的接口
// 所有與伺服器相關的業務邏輯方法都應該在這裡声明
type ServerServiceInterface interface {
	// CreateServer 新建伺服器
	CreateServer(userID string, name string, file multipart.File, header *multipart.FileHeader) (*models.ServerResponse, *models.MessageOptions)

	// GetServerListResponse 獲取用戶的伺服器列表回應格式
	GetServerListResponse(userID string) ([]models.ServerResponse, *models.MessageOptions)

	// GetServerChannels 獲取伺服器的頻道列表
	GetServerChannels(serverID string) ([]models.Channel, *models.MessageOptions)

	// SearchPublicServers 搜尋公開伺服器
	SearchPublicServers(userID string, request models.ServerSearchRequest) (*models.ServerSearchResults, *models.MessageOptions)

	// UpdateServer 更新伺服器信息
	UpdateServer(userID string, serverID string, updates map[string]interface{}) (*models.ServerResponse, *models.MessageOptions)

	// DeleteServer 刪除伺服器
	DeleteServer(userID string, serverID string) *models.MessageOptions

	// GetServerByID 根據ID獲取伺服器信息
	GetServerByID(userID string, serverID string) (*models.ServerResponse, *models.MessageOptions)

	// GetServerDetailByID 獲取伺服器詳細信息（包含成員和頻道列表）
	GetServerDetailByID(userID string, serverID string) (*models.ServerDetailResponse, *models.MessageOptions)

	// JoinServer 請求加入伺服器
	JoinServer(userID string, serverID string) *models.MessageOptions

	// LeaveServer 離開伺服器
	LeaveServer(userID string, serverID string) *models.MessageOptions

	// 其他已實現的方法應該添加到這裡
	// ... 其他方法 ...
}

type FriendServiceInterface interface {
	// GetFriendList 獲取好友列表
	GetFriendList(userID string) ([]models.FriendResponse, *models.MessageOptions)

	// AddFriendRequest 發送好友請求
	AddFriendRequest(userID string, username string) *models.MessageOptions

	// GetPendingRequests 獲取待處理好友請求
	GetPendingRequests(userID string) (*models.PendingRequestsResponse, *models.MessageOptions)

	// GetBlockedUsers 獲取封鎖用戶列表
	GetBlockedUsers(userID string) ([]models.BlockedUserResponse, *models.MessageOptions)

	// AcceptFriendRequest 接受好友請求
	AcceptFriendRequest(userID string, requestID string) *models.MessageOptions

	// DeclineFriendRequest 拒絕好友請求
	DeclineFriendRequest(userID string, requestID string) *models.MessageOptions

	// CancelFriendRequest 取消好友請求
	CancelFriendRequest(userID string, requestID string) *models.MessageOptions

	// BlockUser 封鎖用戶
	BlockUser(userID string, targetUserID string) *models.MessageOptions

	// UnblockUser 解除封鎖用戶
	UnblockUser(userID string, targetUserID string) *models.MessageOptions

	// RemoveFriend 刪除好友
	RemoveFriend(userID string, friendID string) *models.MessageOptions
}

type ChannelServiceInterface interface {
	// GetChannelsByServerID 根據伺服器ID獲取頻道列表
	GetChannelsByServerID(userID string, serverID string) ([]models.ChannelResponse, *models.MessageOptions)

	// GetChannelByID 根據頻道ID獲取頻道詳細信息
	GetChannelByID(userID string, channelID string) (*models.ChannelResponse, *models.MessageOptions)

	// CreateChannel 創建新頻道
	CreateChannel(userID string, channel *models.Channel) (*models.ChannelResponse, *models.MessageOptions)

	// UpdateChannel 更新頻道信息
	UpdateChannel(userID string, channelID string, updates map[string]interface{}) (*models.ChannelResponse, *models.MessageOptions)

	// DeleteChannel 刪除頻道
	DeleteChannel(userID string, channelID string) *models.MessageOptions
}

// FileUploadService - 負責業務邏輯和安全驗證
type FileUploadServiceInterface interface {
	// 業務方法
	UploadFile(file multipart.File, header *multipart.FileHeader, userID string) (*models.FileResult, *models.MessageOptions)
	UploadAvatar(file multipart.File, header *multipart.FileHeader, userID string) (*models.FileResult, *models.MessageOptions)
	UploadDocument(file multipart.File, header *multipart.FileHeader, userID string) (*models.FileResult, *models.MessageOptions)
	UploadFileWithConfig(file multipart.File, header *multipart.FileHeader, userID string, config *models.FileUploadConfig) (*models.FileResult, *models.MessageOptions)

	// 驗證方法
	ValidateFile(header *multipart.FileHeader) *models.MessageOptions
	ValidateImage(header *multipart.FileHeader) *models.MessageOptions
	ValidateDocument(header *multipart.FileHeader) *models.MessageOptions

	// 安全檢查方法
	ScanFileForMalware(filePath string) *models.MessageOptions
	CheckFileContent(file multipart.File, header *multipart.FileHeader) *models.MessageOptions

	// 檔案管理方法
	DeleteFile(filePath string) *models.MessageOptions
	DeleteFileByID(fileID string, userID string) *models.MessageOptions
	GetFileInfo(filePath string) (*models.FileInfo, *models.MessageOptions)
	GetFileURLByID(fileID string) (string, *models.MessageOptions)
	GetFileInfoByID(fileID string) (*models.UploadedFile, *models.MessageOptions)
	GetUserFiles(userID string) ([]*models.UploadedFile, *models.MessageOptions)
	CleanupExpiredFiles() *models.MessageOptions
}
