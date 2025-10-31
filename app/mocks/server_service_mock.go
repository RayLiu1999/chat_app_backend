package mocks

import (
	"chat_app_backend/app/models"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

// ServerService 是 services.ServerService 介面的 mock 實現
type ServerService struct {
	mock.Mock
}

// CreateServer 新建伺服器
func (m *ServerService) CreateServer(userID string, name string, file multipart.File, header *multipart.FileHeader) (*models.ServerResponse, *models.MessageOptions) {
	args := m.Called(userID, name, file, header)
	var resp *models.ServerResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.ServerResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// GetServerListResponse 獲取用戶的伺服器列表回應格式
func (m *ServerService) GetServerListResponse(userID string) ([]models.ServerResponse, *models.MessageOptions) {
	args := m.Called(userID)
	var serverList []models.ServerResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		serverList = args.Get(0).([]models.ServerResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return serverList, msgOpts
}

// GetServerChannels 獲取伺服器的頻道列表
func (m *ServerService) GetServerChannels(serverID string) ([]models.Channel, *models.MessageOptions) {
	args := m.Called(serverID)
	var channels []models.Channel
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		channels = args.Get(0).([]models.Channel)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return channels, msgOpts
}

// SearchPublicServers 搜尋公開伺服器
func (m *ServerService) SearchPublicServers(userID string, request models.ServerSearchRequest) (*models.ServerSearchResults, *models.MessageOptions) {
	args := m.Called(userID, request)
	var results *models.ServerSearchResults
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		results = args.Get(0).(*models.ServerSearchResults)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return results, msgOpts
}

// UpdateServer 更新伺服器信息
func (m *ServerService) UpdateServer(userID string, serverID string, updates map[string]any) (*models.ServerResponse, *models.MessageOptions) {
	args := m.Called(userID, serverID, updates)
	var resp *models.ServerResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.ServerResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// DeleteServer 刪除伺服器
func (m *ServerService) DeleteServer(userID string, serverID string) *models.MessageOptions {
	args := m.Called(userID, serverID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// GetServerByID 根據ID獲取伺服器信息
func (m *ServerService) GetServerByID(userID string, serverID string) (*models.ServerResponse, *models.MessageOptions) {
	args := m.Called(userID, serverID)
	var resp *models.ServerResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.ServerResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// GetServerDetailByID 獲取伺服器詳細信息（包含成員和頻道列表）
func (m *ServerService) GetServerDetailByID(userID string, serverID string) (*models.ServerDetailResponse, *models.MessageOptions) {
	args := m.Called(userID, serverID)
	var resp *models.ServerDetailResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.ServerDetailResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// JoinServer 請求加入伺服器
func (m *ServerService) JoinServer(userID string, serverID string) *models.MessageOptions {
	args := m.Called(userID, serverID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// LeaveServer 離開伺服器
func (m *ServerService) LeaveServer(userID string, serverID string) *models.MessageOptions {
	args := m.Called(userID, serverID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}
