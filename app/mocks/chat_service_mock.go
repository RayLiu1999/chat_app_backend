package mocks

import (
	"chat_app_backend/app/models"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
)

// ChatService 是 services.ChatService 介面的 mock 實現
type ChatService struct {
	mock.Mock
}

// HandleWebSocket 處理 WebSocket 連接
func (m *ChatService) HandleWebSocket(ws *websocket.Conn, userID string) {
	m.Called(ws, userID)
}

// GetDMRoomResponseList 獲取聊天列表response
func (m *ChatService) GetDMRoomResponseList(userID string, includeNotVisible bool) ([]models.DMRoomResponse, *models.MessageOptions) {
	args := m.Called(userID, includeNotVisible)
	var roomList []models.DMRoomResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		roomList = args.Get(0).([]models.DMRoomResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return roomList, msgOpts
}

// UpdateDMRoom 更新聊天房間狀態
func (m *ChatService) UpdateDMRoom(userID string, roomID string, isHidden bool) *models.MessageOptions {
	args := m.Called(userID, roomID, isHidden)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// CreateDMRoom 創建私聊房間
func (m *ChatService) CreateDMRoom(userID string, chatWithUserID string) (*models.DMRoomResponse, *models.MessageOptions) {
	args := m.Called(userID, chatWithUserID)
	var resp *models.DMRoomResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.DMRoomResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// GetDMMessages 獲取私聊訊息
func (m *ChatService) GetDMMessages(userID string, roomID string, before string, after string, limit string) ([]models.MessageResponse, *models.MessageOptions) {
	args := m.Called(userID, roomID, before, after, limit)
	var messages []models.MessageResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		messages = args.Get(0).([]models.MessageResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return messages, msgOpts
}

// GetChannelMessages 獲取頻道訊息
func (m *ChatService) GetChannelMessages(userID string, channelID string, before string, after string, limit string) ([]models.MessageResponse, *models.MessageOptions) {
	args := m.Called(userID, channelID, before, after, limit)
	var messages []models.MessageResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		messages = args.Get(0).([]models.MessageResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return messages, msgOpts
}
