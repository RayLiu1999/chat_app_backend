package mocks

import (
	"chat_app_backend/app/models"

	"github.com/stretchr/testify/mock"
)

// UserRepository 是 repositories.UserRepository 介面的 mock 實作
type UserRepository struct {
	mock.Mock
}

func (m *UserRepository) GetUserById(userID string) (*models.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) GetUserListByIds(userIds []string) ([]models.User, error) {
	args := m.Called(userIds)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *UserRepository) CheckUsernameExists(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) CheckEmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *UserRepository) CreateUser(user models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepository) UpdateUserOnlineStatus(userID string, isOnline bool) error {
	args := m.Called(userID, isOnline)
	return args.Error(0)
}

func (m *UserRepository) UpdateUserLastActiveTime(userID string, timestamp int64) error {
	args := m.Called(userID, timestamp)
	return args.Error(0)
}

func (m *UserRepository) UpdateUser(userID string, updates map[string]any) error {
	args := m.Called(userID, updates)
	return args.Error(0)
}

func (m *UserRepository) DeleteUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}
