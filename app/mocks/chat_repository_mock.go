package mocks

import (
	"chat_app_backend/app/models"
	"context"

	"github.com/stretchr/testify/mock"
)

// ChatRepository 是 repositories.ChatRepository 介面的 mock 實作
type ChatRepository struct {
	mock.Mock
}

func (m *ChatRepository) SaveMessage(ctx context.Context, message models.Message) (string, error) {
	args := m.Called(ctx, message)
	return args.String(0), args.Error(1)
}

func (m *ChatRepository) GetMessagesByRoomID(ctx context.Context, roomID string, limit int64) ([]models.Message, error) {
	args := m.Called(ctx, roomID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Message), args.Error(1)
}

func (m *ChatRepository) GetDMRoomListByUserID(ctx context.Context, userID string, includeNotVisible bool) ([]models.DMRoom, error) {
	args := m.Called(ctx, userID, includeNotVisible)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.DMRoom), args.Error(1)
}

func (m *ChatRepository) UpdateDMRoom(ctx context.Context, userID string, chatWithUserID string, IsHidden bool) error {
	args := m.Called(ctx, userID, chatWithUserID, IsHidden)
	return args.Error(0)
}

func (m *ChatRepository) SaveOrUpdateDMRoom(ctx context.Context, chat models.DMRoom) (models.DMRoom, error) {
	args := m.Called(ctx, chat)
	return args.Get(0).(models.DMRoom), args.Error(1)
}

func (m *ChatRepository) DeleteMessagesByRoomID(ctx context.Context, roomID string) error {
	args := m.Called(ctx, roomID)
	return args.Error(0)
}
