package services

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Mock Components for Channel Service ---

// mockChannelServiceChannelRepository 模擬 ChannelRepository (Channel Service 專用)
type mockChannelServiceChannelRepository struct {
	mock.Mock
}

func (m *mockChannelServiceChannelRepository) GetChannelsByServerID(serverID string) ([]models.Channel, error) {
	args := m.Called(serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Channel), args.Error(1)
}

func (m *mockChannelServiceChannelRepository) GetChannelByID(channelID string) (*models.Channel, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *mockChannelServiceChannelRepository) CreateChannel(channel *models.Channel) error {
	args := m.Called(channel)
	return args.Error(0)
}

func (m *mockChannelServiceChannelRepository) UpdateChannel(channelID string, updates map[string]any) error {
	args := m.Called(channelID, updates)
	return args.Error(0)
}

func (m *mockChannelServiceChannelRepository) DeleteChannel(channelID string) error {
	args := m.Called(channelID)
	return args.Error(0)
}

func (m *mockChannelServiceChannelRepository) CheckChannelExists(channelID string) (bool, error) {
	args := m.Called(channelID)
	return args.Bool(0), args.Error(1)
}

// mockChannelServiceServerMemberRepository 模擬 ServerMemberRepository
type mockChannelServiceServerMemberRepository struct {
	mock.Mock
}

func (m *mockChannelServiceServerMemberRepository) AddMemberToServer(serverID, userID string, role string) error {
	args := m.Called(serverID, userID, role)
	return args.Error(0)
}

func (m *mockChannelServiceServerMemberRepository) RemoveMemberFromServer(serverID, userID string) error {
	args := m.Called(serverID, userID)
	return args.Error(0)
}

func (m *mockChannelServiceServerMemberRepository) GetServerMembers(serverID string, page, limit int) ([]models.ServerMember, int64, error) {
	args := m.Called(serverID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.ServerMember), args.Get(1).(int64), args.Error(2)
}

func (m *mockChannelServiceServerMemberRepository) GetUserServers(userID string) ([]models.ServerMember, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ServerMember), args.Error(1)
}

func (m *mockChannelServiceServerMemberRepository) IsMemberOfServer(serverID, userID string) (bool, error) {
	args := m.Called(serverID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *mockChannelServiceServerMemberRepository) UpdateMemberRole(serverID, userID, newRole string) error {
	args := m.Called(serverID, userID, newRole)
	return args.Error(0)
}

func (m *mockChannelServiceServerMemberRepository) GetMemberCount(serverID string) (int64, error) {
	args := m.Called(serverID)
	return args.Get(0).(int64), args.Error(1)
}

// --- Tests ---

func TestNewChannelService(t *testing.T) {
	mockChannelRepo := new(mockChannelServiceChannelRepository)
	mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)
	mockChatRepo := new(mocks.ChatRepository)

	service := NewChannelService(
		nil,
		nil,
		mockChannelRepo,
		nil,
		mockServerMemberRepo,
		nil,
		mockChatRepo,
	)

	assert.NotNil(t, service)
	assert.Equal(t, mockChannelRepo, service.channelRepo)
	assert.Equal(t, mockServerMemberRepo, service.serverMemberRepo)
	assert.Equal(t, mockChatRepo, service.chatRepo)
}

func TestGetChannelsByServerID(t *testing.T) {
	t.Run("成功獲取頻道列表", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID1 := primitive.NewObjectID()
		channelID2 := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		channels := []models.Channel{
			{
				BaseModel: providers.BaseModel{ID: channelID1},
				ServerID:  serverID,
				Name:      "general",
				Type:      "text",
			},
			{
				BaseModel: providers.BaseModel{ID: channelID2},
				ServerID:  serverID,
				Name:      "voice-channel",
				Type:      "voice",
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("GetChannelsByServerID", serverID.Hex()).Return(channels, nil).Once()

		result, msgOpt := service.GetChannelsByServerID(userID.Hex(), serverID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "general", result[0].Name)
		assert.Equal(t, "text", result[0].Type)
		assert.Equal(t, "voice-channel", result[1].Name)
		assert.Equal(t, "voice", result[1].Type)

		mockServerMemberRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("獲取用戶伺服器列表失敗", func(t *testing.T) {
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &channelService{
			serverMemberRepo: mockServerMemberRepo,
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(nil, errors.New("database error")).Once()

		result, msgOpt := service.GetChannelsByServerID(userID.Hex(), serverID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "獲取用戶伺服器列表失敗")

		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("用戶沒有權限訪問該伺服器", func(t *testing.T) {
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		otherServerID := primitive.NewObjectID()

		service := &channelService{
			serverMemberRepo: mockServerMemberRepo,
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  otherServerID, // 不同的伺服器ID
				UserID:    userID,
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()

		result, msgOpt := service.GetChannelsByServerID(userID.Hex(), serverID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "用戶沒有權限訪問該伺服器")

		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("獲取頻道列表失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("GetChannelsByServerID", serverID.Hex()).Return(nil, errors.New("database error")).Once()

		result, msgOpt := service.GetChannelsByServerID(userID.Hex(), serverID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "獲取頻道列表失敗")

		mockServerMemberRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("用戶沒有加入任何伺服器", func(t *testing.T) {
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &channelService{
			serverMemberRepo: mockServerMemberRepo,
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return([]models.ServerMember{}, nil).Once()

		result, msgOpt := service.GetChannelsByServerID(userID.Hex(), serverID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestGetChannelByID(t *testing.T) {
	t.Run("成功獲取頻道信息", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "general",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()

		result, msgOpt := service.GetChannelByID(userID.Hex(), channelID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, channelID, result.ID)
		assert.Equal(t, "general", result.Name)
		assert.Equal(t, "text", result.Type)

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("獲取頻道信息失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)

		userID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo: mockChannelRepo,
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(nil, errors.New("channel not found")).Once()

		result, msgOpt := service.GetChannelByID(userID.Hex(), channelID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "獲取頻道信息失敗")

		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("用戶沒有權限訪問該頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		otherServerID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "general",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  otherServerID, // 不同的伺服器
				UserID:    userID,
			},
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()

		result, msgOpt := service.GetChannelByID(userID.Hex(), channelID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "用戶沒有權限訪問該頻道")

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestCreateChannel(t *testing.T) {
	t.Run("成功創建頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			ServerID: serverID,
			Name:     "new-channel",
			Type:     "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("CreateChannel", mock.AnythingOfType("*models.Channel")).Return(nil).Once()

		result, msgOpt := service.CreateChannel(userID.Hex(), channel)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "new-channel", result.Name)
		assert.Equal(t, "text", result.Type)
		assert.False(t, result.ID.IsZero())

		mockServerMemberRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("用戶沒有權限創建頻道", func(t *testing.T) {
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		otherServerID := primitive.NewObjectID()

		service := &channelService{
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			ServerID: serverID,
			Name:     "new-channel",
			Type:     "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  otherServerID, // 不同的伺服器
				UserID:    userID,
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()

		result, msgOpt := service.CreateChannel(userID.Hex(), channel)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "用戶沒有權限在該伺服器創建頻道")

		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("創建頻道失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			ServerID: serverID,
			Name:     "new-channel",
			Type:     "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("CreateChannel", mock.AnythingOfType("*models.Channel")).Return(errors.New("database error")).Once()

		result, msgOpt := service.CreateChannel(userID.Hex(), channel)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "創建頻道失敗")

		mockServerMemberRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("頻道已有ID則使用該ID", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "new-channel",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("CreateChannel", mock.AnythingOfType("*models.Channel")).Return(nil).Once()

		result, msgOpt := service.CreateChannel(userID.Hex(), channel)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, channelID, result.ID)

		mockServerMemberRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
	})
}

func TestUpdateChannel(t *testing.T) {
	t.Run("成功更新頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "old-name",
			Type:      "text",
		}

		updatedChannel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "new-name",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		updates := map[string]any{"name": "new-name"}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("UpdateChannel", channelID.Hex(), updates).Return(nil).Once()
		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(updatedChannel, nil).Once()

		result, msgOpt := service.UpdateChannel(userID.Hex(), channelID.Hex(), updates)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "new-name", result.Name)

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("獲取頻道信息失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)

		userID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo: mockChannelRepo,
		}

		updates := map[string]any{"name": "new-name"}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(nil, errors.New("channel not found")).Once()

		result, msgOpt := service.UpdateChannel(userID.Hex(), channelID.Hex(), updates)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("用戶沒有權限更新頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		otherServerID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "old-name",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  otherServerID, // 不同的伺服器
				UserID:    userID,
			},
		}

		updates := map[string]any{"name": "new-name"}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()

		result, msgOpt := service.UpdateChannel(userID.Hex(), channelID.Hex(), updates)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("更新頻道失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "old-name",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		updates := map[string]any{"name": "new-name"}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChannelRepo.On("UpdateChannel", channelID.Hex(), updates).Return(errors.New("database error")).Once()

		result, msgOpt := service.UpdateChannel(userID.Hex(), channelID.Hex(), updates)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "更新頻道失敗")

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestDeleteChannel(t *testing.T) {
	t.Run("成功刪除頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)
		mockChatRepo := new(mocks.ChatRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
			chatRepo:         mockChatRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "to-delete",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChatRepo.On("DeleteMessagesByRoomID", channelID.Hex()).Return(nil).Once()
		mockChannelRepo.On("DeleteChannel", channelID.Hex()).Return(nil).Once()

		msgOpt := service.DeleteChannel(userID.Hex(), channelID.Hex())

		assert.Nil(t, msgOpt)

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
		mockChatRepo.AssertExpectations(t)
	})

	t.Run("獲取頻道信息失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)

		userID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo: mockChannelRepo,
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(nil, errors.New("channel not found")).Once()

		msgOpt := service.DeleteChannel(userID.Hex(), channelID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("用戶沒有權限刪除頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		otherServerID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "to-delete",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  otherServerID, // 不同的伺服器
				UserID:    userID,
			},
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()

		msgOpt := service.DeleteChannel(userID.Hex(), channelID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("刪除訊息失敗但繼續刪除頻道", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)
		mockChatRepo := new(mocks.ChatRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
			chatRepo:         mockChatRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "to-delete",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChatRepo.On("DeleteMessagesByRoomID", channelID.Hex()).Return(errors.New("delete messages error")).Once()
		mockChannelRepo.On("DeleteChannel", channelID.Hex()).Return(nil).Once()

		msgOpt := service.DeleteChannel(userID.Hex(), channelID.Hex())

		// 即使刪除訊息失敗，頻道仍應該被刪除
		assert.Nil(t, msgOpt)

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
		mockChatRepo.AssertExpectations(t)
	})

	t.Run("刪除頻道失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelServiceChannelRepository)
		mockServerMemberRepo := new(mockChannelServiceServerMemberRepository)
		mockChatRepo := new(mocks.ChatRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &channelService{
			channelRepo:      mockChannelRepo,
			serverMemberRepo: mockServerMemberRepo,
			chatRepo:         mockChatRepo,
		}

		channel := &models.Channel{
			BaseModel: providers.BaseModel{ID: channelID},
			ServerID:  serverID,
			Name:      "to-delete",
			Type:      "text",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		mockChannelRepo.On("GetChannelByID", channelID.Hex()).Return(channel, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockChatRepo.On("DeleteMessagesByRoomID", channelID.Hex()).Return(nil).Once()
		mockChannelRepo.On("DeleteChannel", channelID.Hex()).Return(errors.New("database error")).Once()

		msgOpt := service.DeleteChannel(userID.Hex(), channelID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "刪除頻道失敗")

		mockChannelRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
		mockChatRepo.AssertExpectations(t)
	})
}
