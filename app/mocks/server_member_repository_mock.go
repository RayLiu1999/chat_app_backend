package mocks

import (
	"chat_app_backend/app/models"

	"github.com/stretchr/testify/mock"
)

// ServerMemberRepository 是 repositories.ServerMemberRepository 介面的 mock 實作
type ServerMemberRepository struct {
	mock.Mock
}

// AddMemberToServer 將用戶添加到伺服器
func (m *ServerMemberRepository) AddMemberToServer(serverID, userID string, role string) error {
	args := m.Called(serverID, userID, role)
	return args.Error(0)
}

// RemoveMemberFromServer 從伺服器移除用戶
func (m *ServerMemberRepository) RemoveMemberFromServer(serverID, userID string) error {
	args := m.Called(serverID, userID)
	return args.Error(0)
}

// GetServerMembers 獲取伺服器所有成員
func (m *ServerMemberRepository) GetServerMembers(serverID string, page, limit int) ([]models.ServerMember, int64, error) {
	args := m.Called(serverID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.ServerMember), args.Get(1).(int64), args.Error(2)
}

// GetUserServers 獲取用戶加入的所有伺服器
func (m *ServerMemberRepository) GetUserServers(userID string) ([]models.ServerMember, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ServerMember), args.Error(1)
}

// IsMemberOfServer 檢查用戶是否為伺服器成員
func (m *ServerMemberRepository) IsMemberOfServer(serverID, userID string) (bool, error) {
	args := m.Called(serverID, userID)
	return args.Bool(0), args.Error(1)
}

// UpdateMemberRole 更新成員角色
func (m *ServerMemberRepository) UpdateMemberRole(serverID, userID, newRole string) error {
	args := m.Called(serverID, userID, newRole)
	return args.Error(0)
}

// GetMemberCount 獲取伺服器成員數量
func (m *ServerMemberRepository) GetMemberCount(serverID string) (int64, error) {
	args := m.Called(serverID)
	return args.Get(0).(int64), args.Error(1)
}
