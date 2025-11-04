package mocks

import (
	"chat_app_backend/app/models"

	"github.com/stretchr/testify/mock"
)

// ChannelService 是 services.ChannelService 介面的 mock 實作
type ChannelService struct {
	mock.Mock
}

// GetChannelsByServerID 根據伺服器ID獲取頻道列表
func (m *ChannelService) GetChannelsByServerID(userID string, serverID string) ([]models.ChannelResponse, *models.MessageOptions) {
	args := m.Called(userID, serverID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).([]models.ChannelResponse), args.Get(1).(*models.MessageOptions)
}

// GetChannelByID 根據頻道ID獲取頻道詳細信息
func (m *ChannelService) GetChannelByID(userID string, channelID string) (*models.ChannelResponse, *models.MessageOptions) {
	args := m.Called(userID, channelID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.ChannelResponse), args.Get(1).(*models.MessageOptions)
}

// CreateChannel 創建新頻道
func (m *ChannelService) CreateChannel(userID string, channel *models.Channel) (*models.ChannelResponse, *models.MessageOptions) {
	args := m.Called(userID, channel)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.ChannelResponse), args.Get(1).(*models.MessageOptions)
}

// UpdateChannel 更新頻道信息
func (m *ChannelService) UpdateChannel(userID string, channelID string, updates map[string]any) (*models.ChannelResponse, *models.MessageOptions) {
	args := m.Called(userID, channelID, updates)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.ChannelResponse), args.Get(1).(*models.MessageOptions)
}

// DeleteChannel 刪除頻道
func (m *ChannelService) DeleteChannel(userID string, channelID string) *models.MessageOptions {
	args := m.Called(userID, channelID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}
