package repositories

import (
	"chat_app_backend/app/models"
)

type ChatRepository interface {
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

	// DeleteMessagesByRoomID 根據房間ID刪除所有訊息
	DeleteMessagesByRoomID(roomID string) error
}

type ServerRepository interface {
	// CreateServer 新建測試用戶伺服器關聯
	CreateServer(server *models.Server) (models.Server, error)

	// SearchPublicServers 搜尋公開伺服器
	SearchPublicServers(request models.ServerSearchRequest) ([]models.Server, int64, error)

	// GetServerWithOwnerInfo 獲取包含擁有者信息的伺服器
	GetServerWithOwnerInfo(serverID string) (*models.Server, *models.User, error)

	// CheckUserInServer 檢查用戶是否在伺服器中
	CheckUserInServer(userID, serverID string) (bool, error)

	// UpdateServer 更新伺服器信息
	UpdateServer(serverID string, updates map[string]any) error

	// DeleteServer 刪除伺服器
	DeleteServer(serverID string) error

	// GetServerByID 根據ID獲取伺服器
	GetServerByID(serverID string) (*models.Server, error)

	// UpdateMemberCount 更新成員數量快取
	UpdateMemberCount(serverID string, count int) error
}

type ServerMemberRepository interface {
	// AddMemberToServer 將用戶添加到伺服器
	AddMemberToServer(serverID, userID string, role string) error

	// RemoveMemberFromServer 從伺服器移除用戶
	RemoveMemberFromServer(serverID, userID string) error

	// GetServerMembers 獲取伺服器所有成員
	GetServerMembers(serverID string, page, limit int) ([]models.ServerMember, int64, error)

	// GetUserServers 獲取用戶加入的所有伺服器
	GetUserServers(userID string) ([]models.ServerMember, error)

	// IsMemberOfServer 檢查用戶是否為伺服器成員
	IsMemberOfServer(serverID, userID string) (bool, error)

	// UpdateMemberRole 更新成員角色
	UpdateMemberRole(serverID, userID, newRole string) error

	// GetMemberCount 獲取伺服器成員數量
	GetMemberCount(serverID string) (int64, error)
}

type UserRepository interface {
	// GetUserById 根據用戶ID獲取用戶
	GetUserById(userID string) (*models.User, error)

	// GetUserByUsername 根據用戶名獲取用戶（測試用）
	GetUserByUsername(username string) (*models.User, error)

	// GetUserListByIds 根據用戶ID陣列獲取用戶
	GetUserListByIds(userIds []string) ([]models.User, error)

	// CheckUsernameExists 檢查用戶名是否已存在
	CheckUsernameExists(username string) (bool, error)

	// CheckEmailExists 檢查電子郵件是否已存在
	CheckEmailExists(email string) (bool, error)

	// CreateUser 創建用戶
	CreateUser(user models.User) error

	// UpdateUserOnlineStatus 更新用戶在線狀態
	UpdateUserOnlineStatus(userID string, isOnline bool) error

	// UpdateUserLastActiveTime 更新用戶最後活動時間
	UpdateUserLastActiveTime(userID string, timestamp int64) error

	// UpdateUser 更新用戶信息
	UpdateUser(userID string, updates map[string]any) error

	// DeleteUser 刪除用戶
	DeleteUser(userID string) error
}

type FriendRepository interface {
	// GetFriendById 根據用戶ID獲取用戶
	GetFriendById(userID string) (*models.Friend, error)
}

type ChannelRepository interface {
	// GetChannelsByServerID 根據伺服器ID獲取頻道列表
	GetChannelsByServerID(serverID string) ([]models.Channel, error)

	// GetChannelByID 根據頻道ID獲取頻道
	GetChannelByID(channelID string) (*models.Channel, error)

	// CreateChannel 創建新頻道
	CreateChannel(channel *models.Channel) error

	// UpdateChannel 更新頻道
	UpdateChannel(channelID string, updates map[string]any) error

	// DeleteChannel 刪除頻道
	DeleteChannel(channelID string) error

	// CheckChannelExists 檢查頻道是否存在
	CheckChannelExists(channelID string) (bool, error)
}

type FileRepository interface {
	// CreateFile 創建檔案記錄
	CreateFile(file *models.UploadedFile) error

	// GetFileByID 根據檔案ID獲取檔案
	GetFileByID(fileID string) (*models.UploadedFile, error)

	// GetFileByPath 根據檔案路徑獲取檔案
	GetFileByPath(filePath string) (*models.UploadedFile, error)

	// GetFilesByUserID 根據用戶ID獲取檔案列表
	GetFilesByUserID(userID string) ([]models.UploadedFile, error)

	// UpdateFileStatus 更新檔案狀態
	UpdateFileStatus(fileID string, status string) error

	// DeleteFileByID 根據檔案ID刪除檔案記錄
	DeleteFileByID(fileID string) error

	// DeleteFileByPath 根據檔案路徑刪除檔案記錄
	DeleteFileByPath(filePath string) error

	// GetExpiredFiles 獲取過期檔案列表
	GetExpiredFiles() ([]models.UploadedFile, error)

	// CleanupExpiredFiles 清理過期檔案記錄
	CleanupExpiredFiles() error
}

type ChannelCategoryRepository interface {
	// CreateChannelCategory 創建頻道類別
	CreateChannelCategory(category *models.ChannelCategory) error

	// GetChannelCategoriesByServerID 根據伺服器ID獲取頻道類別列表
	GetChannelCategoriesByServerID(serverID string) ([]models.ChannelCategory, error)

	// GetChannelCategoryByID 根據類別ID獲取頻道類別
	GetChannelCategoryByID(categoryID string) (*models.ChannelCategory, error)

	// UpdateChannelCategory 更新頻道類別
	UpdateChannelCategory(categoryID string, updates map[string]any) error

	// DeleteChannelCategory 刪除頻道類別
	DeleteChannelCategory(categoryID string) error

	// CheckChannelCategoryExists 檢查頻道類別是否存在
	CheckChannelCategoryExists(categoryID string) (bool, error)
}
