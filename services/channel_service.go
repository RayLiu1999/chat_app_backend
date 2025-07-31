package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChannelService struct {
	config           *config.Config
	odm              *providers.ODM
	channelRepo      repositories.ChannelRepositoryInterface
	serverRepo       repositories.ServerRepositoryInterface
	serverMemberRepo repositories.ServerMemberRepositoryInterface
	userRepo         repositories.UserRepositoryInterface
	chatRepo         repositories.ChatRepositoryInterface
}

func NewChannelService(cfg *config.Config,
	odm *providers.ODM,
	channelRepo repositories.ChannelRepositoryInterface,
	serverRepo repositories.ServerRepositoryInterface,
	serverMemberRepo repositories.ServerMemberRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	chatRepo repositories.ChatRepositoryInterface) ChannelServiceInterface {
	return &ChannelService{
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
func (cs *ChannelService) GetChannelsByServerID(userID string, serverID string) ([]models.ChannelResponse, error) {
	// 檢查用戶是否有權限訪問該伺服器
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID.Hex() == serverID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("用戶沒有權限訪問該伺服器")
	}

	// 獲取頻道列表
	channels, err := cs.channelRepo.GetChannelsByServerID(serverID)
	if err != nil {
		return nil, err
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
func (cs *ChannelService) GetChannelByID(userID string, channelID string) (*models.ChannelResponse, error) {
	// 獲取頻道信息
	channel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, err
	}

	// 檢查用戶是否有權限訪問該頻道所屬的伺服器
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("用戶沒有權限訪問該頻道")
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
func (cs *ChannelService) CreateChannel(userID string, channel *models.Channel) (*models.ChannelResponse, error) {
	// 檢查用戶是否有權限創建頻道（需要是伺服器成員）
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("用戶沒有權限在該伺服器創建頻道")
	}

	// 設置頻道ID
	if channel.ID.IsZero() {
		channel.ID = primitive.NewObjectID()
	}

	// 創建頻道
	err = cs.channelRepo.CreateChannel(channel)
	if err != nil {
		return nil, err
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
func (cs *ChannelService) UpdateChannel(userID string, channelID string, updates map[string]interface{}) (*models.ChannelResponse, error) {
	// 獲取頻道信息以檢查權限
	channel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, err
	}

	// 檢查用戶是否有權限更新該頻道
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return nil, err
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, errors.New("用戶沒有權限更新該頻道")
	}

	// 更新頻道
	err = cs.channelRepo.UpdateChannel(channelID, updates)
	if err != nil {
		return nil, err
	}

	// 重新獲取更新後的頻道信息
	updatedChannel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return nil, err
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
func (cs *ChannelService) DeleteChannel(userID string, channelID string) error {
	// 獲取頻道信息以檢查權限
	channel, err := cs.channelRepo.GetChannelByID(channelID)
	if err != nil {
		return err
	}

	// 檢查用戶是否有權限刪除該頻道
	serverMembers, err := cs.serverMemberRepo.GetUserServers(userID)
	if err != nil {
		return err
	}

	hasAccess := false
	for _, member := range serverMembers {
		if member.ServerID == channel.ServerID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return errors.New("用戶沒有權限刪除該頻道")
	}

	// 先刪除該頻道的所有訊息
	err = cs.chatRepo.DeleteMessagesByRoomID(channelID)
	if err != nil {
		// 記錄錯誤但不阻止頻道刪除
		log.Printf("刪除頻道 %s 的訊息失敗: %v", channelID, err)
	}

	// 然後刪除頻道本身
	return cs.channelRepo.DeleteChannel(channelID)
}
