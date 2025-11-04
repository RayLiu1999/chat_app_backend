package mocks

import (
	"chat_app_backend/app/models"

	"github.com/stretchr/testify/mock"
)

// ChatRepository 是 repositories.ChatRepository 介面的 mock 實作
type ChatRepository struct {
	mock.Mock
}

func (m *ChatRepository) SaveMessage(message models.Message) (string, error) {
	args := m.Called(message)
	return args.String(0), args.Error(1)
}

func (m *ChatRepository) GetMessagesByRoomID(roomID string, limit int64) ([]models.Message, error) {
	args := m.Called(roomID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *ChatRepository) GetDMRoomListByUserID(userID string, includeNotVisible bool) ([]models.DMRoom, error) {
	args := m.Called(userID, includeNotVisible)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DMRoom), args.Error(1)
}

func (m *ChatRepository) UpdateDMRoom(userID string, chatWithUserID string, IsHidden bool) error {
	args := m.Called(userID, chatWithUserID, IsHidden)
	return args.Error(0)
}

func (m *ChatRepository) SaveOrUpdateDMRoom(chat models.DMRoom) (models.DMRoom, error) {
	args := m.Called(chat)
	return args.Get(0).(models.DMRoom), args.Error(1)
}

func (m *ChatRepository) DeleteMessagesByRoomID(roomID string) error {
	args := m.Called(roomID)
	return args.Error(0)
}
