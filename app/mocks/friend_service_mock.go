package mocks

import (
	"chat_app_backend/app/models"

	"github.com/stretchr/testify/mock"
)

// FriendService 是 services.FriendService 介面的 mock 實現
type FriendService struct {
	mock.Mock
}

// GetFriendList 獲取好友列表
func (m *FriendService) GetFriendList(userID string) ([]models.FriendResponse, *models.MessageOptions) {
	args := m.Called(userID)
	var friendList []models.FriendResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		friendList = args.Get(0).([]models.FriendResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return friendList, msgOpts
}

// AddFriendRequest 發送好友請求
func (m *FriendService) AddFriendRequest(userID string, username string) *models.MessageOptions {
	args := m.Called(userID, username)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// GetPendingRequests 獲取待處理好友請求
func (m *FriendService) GetPendingRequests(userID string) (*models.PendingRequestsResponse, *models.MessageOptions) {
	args := m.Called(userID)
	var resp *models.PendingRequestsResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		resp = args.Get(0).(*models.PendingRequestsResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return resp, msgOpts
}

// GetBlockedUsers 獲取封鎖用戶列表
func (m *FriendService) GetBlockedUsers(userID string) ([]models.BlockedUserResponse, *models.MessageOptions) {
	args := m.Called(userID)
	var blockedList []models.BlockedUserResponse
	var msgOpts *models.MessageOptions

	if args.Get(0) != nil {
		blockedList = args.Get(0).([]models.BlockedUserResponse)
	}
	if args.Get(1) != nil {
		msgOpts = args.Get(1).(*models.MessageOptions)
	}

	return blockedList, msgOpts
}

// AcceptFriendRequest 接受好友請求
func (m *FriendService) AcceptFriendRequest(userID string, requestID string) *models.MessageOptions {
	args := m.Called(userID, requestID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// DeclineFriendRequest 拒絕好友請求
func (m *FriendService) DeclineFriendRequest(userID string, requestID string) *models.MessageOptions {
	args := m.Called(userID, requestID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// RemoveFriend 移除好友
func (m *FriendService) RemoveFriend(userID string, friendID string) *models.MessageOptions {
	args := m.Called(userID, friendID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// BlockUser 封鎖用戶
func (m *FriendService) BlockUser(userID string, blockUserID string) *models.MessageOptions {
	args := m.Called(userID, blockUserID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// UnblockUser 解除封鎖用戶
func (m *FriendService) UnblockUser(userID string, blockUserID string) *models.MessageOptions {
	args := m.Called(userID, blockUserID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// CancelFriendRequest 取消好友請求
func (m *FriendService) CancelFriendRequest(userID string, requestID string) *models.MessageOptions {
	args := m.Called(userID, requestID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}
