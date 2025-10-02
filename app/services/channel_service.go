package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/app/repositories"
	"chat_app_backend/config"
	"chat_app_backend/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type channelService struct {
	config           *config.Config
	odm              providers.ODM
	channelRepo      repositories.ChannelRepository
	serverRepo       repositories.ServerRepository
	serverMemberRepo repositories.ServerMemberRepository
	userRepo         repositories.UserRepository
	chatRepo         repositories.ChatRepository
}

func NewChannelService(cfg *config.Config,
	odm providers.ODM,
	channelRepo repositories.ChannelRepository,
	serverRepo repositories.ServerRepository,
	serverMemberRepo repositories.ServerMemberRepository,
	userRepo repositories.UserRepository,
	chatRepo repositories.ChatRepository) *channelService {
	return &channelService{
		config:           cfg,
		odm:              odm,
		channelRepo:      channelRepo,
		serverRepo:       serverRepo,
		serverMemberRepo: serverMemberRepo,
		userRepo:         userRepo,
		chatRepo:         chatRepo,
	}
}

// GetChannelsByServerID 根據伺服器ID獲取頻道列表
func (cs *channelService) GetChannelsByServerID(userID string, serverID string) ([]models.ChannelResponse, *models.MessageOptions) {
	// 檢查用戶是否有權限訪問該伺服器
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取用戶伺服器列表失敗",
			Details: err.Error(),
		}
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID.Hex() == serverID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "用戶沒有權限訪問該伺服器",
		}
	}

	// 獲取頻道列表
	channels, err := cs.channelRepo.GetChannelsByServerID(serverID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取頻道列表失敗",
			Details: err.Error(),
		}
	}

	// 轉換為響應格式
	var channelResponses []models.ChannelResponse
	for _, channel := range channels {
		channelResponses = append(channelResponses, models.ChannelResponse{
			ID:       channel.ID,
			ServerID: channel.ServerID,
			Name:     channel.Name,
			Type:     channel.Type,
		})
	}

	return channelResponses, nil
}

// GetChannelByID 根據頻道ID獲取頻道詳細信息
func (cs *channelService) GetChannelByID(userID string, channelID string) (*models.ChannelResponse, *models.MessageOptions) {
	// 獲取頻道信息
	channel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取頻道信息失敗",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否有權限訪問該頻道所屬的伺服器
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取用戶伺服器列表失敗",
			Details: err.Error(),
		}
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "用戶沒有權限訪問該頻道",
		}
	}

	// 轉換為響應格式
	channelResponse := &models.ChannelResponse{
		ID:       channel.ID,
		ServerID: channel.ServerID,
		Name:     channel.Name,
		Type:     channel.Type,
	}

	return channelResponse, nil
}

// CreateChannel 創建新頻道
func (cs *channelService) CreateChannel(userID string, channel *models.Channel) (*models.ChannelResponse, *models.MessageOptions) {
	// 檢查用戶是否有權限創建頻道（需要是伺服器成員）
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取用戶伺服器列表失敗",
			Details: err.Error(),
		}
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "用戶沒有權限在該伺服器創建頻道",
		}
	}

	// 設置頻道ID
	if channel.ID.IsZero() {
		channel.ID = primitive.NewObjectID()
	}

	// 創建頻道
	err = cs.channelRepo.CreateChannel(channel)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "創建頻道失敗",
			Details: err.Error(),
		}
	}

	// 返回創建的頻道響應
	channelResponse := &models.ChannelResponse{
		ID:       channel.ID,
		ServerID: channel.ServerID,
		Name:     channel.Name,
		Type:     channel.Type,
	}

	return channelResponse, nil
}

// UpdateChannel 更新頻道信息
func (cs *channelService) UpdateChannel(userID string, channelID string, updates map[string]any) (*models.ChannelResponse, *models.MessageOptions) {
	// 獲取頻道信息以檢查權限
	channel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取頻道信息失敗",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否有權限更新該頻道
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取用戶伺服器列表失敗",
			Details: err.Error(),
		}
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "用戶沒有權限更新該頻道",
		}
	}

	// 更新頻道
	err = cs.channelRepo.UpdateChannel(channelID, updates)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "更新頻道失敗",
			Details: err.Error(),
		}
	}

	// 重新獲取更新後的頻道信息
	updatedChannel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取更新後的頻道信息失敗",
			Details: err.Error(),
		}
	}

	// 返回更新後的頻道響應
	channelResponse := &models.ChannelResponse{
		ID:       updatedChannel.ID,
		ServerID: updatedChannel.ServerID,
		Name:     updatedChannel.Name,
		Type:     updatedChannel.Type,
	}

	return channelResponse, nil
}

// DeleteChannel 刪除頻道
func (cs *channelService) DeleteChannel(userID string, channelID string) *models.MessageOptions {
	// 獲取頻道信息以檢查權限
	channel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取頻道信息失敗",
			Details: err.Error(),
		}
	}

	// 檢查用戶是否有權限刪除該頻道
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "獲取用戶伺服器列表失敗",
			Details: err.Error(),
		}
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "用戶沒有權限刪除該頻道",
		}
	}

	// 先刪除該頻道的所有訊息
	err = cs.chatRepo.DeleteMessagesByRoomID(channelID)
	if err != nil {
		// 記錄錯誤但不阻止頻道刪除
		utils.PrettyPrintf("刪除頻道 %s 的訊息失敗: %v", channelID, err)
	}

	// 然後刪除頻道本身
	err = cs.channelRepo.DeleteChannel(channelID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Message: "刪除頻道失敗",
			Details: err.Error(),
		}
	}

	return nil
}
