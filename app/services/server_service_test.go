package services

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"errors"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Mock Components ---

// mockServerRepository 模擬 ServerRepository
type mockServerRepository struct {
	mock.Mock
}

func (m *mockServerRepository) CreateServer(server *models.Server) (models.Server, error) {
	args := m.Called(server)
	return args.Get(0).(models.Server), args.Error(1)
}

func (m *mockServerRepository) SearchPublicServers(request models.ServerSearchRequest) ([]models.Server, int64, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]models.Server), args.Get(1).(int64), args.Error(2)
}

func (m *mockServerRepository) GetServerWithOwnerInfo(serverID string) (*models.Server, *models.User, error) {
	args := m.Called(serverID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(*models.Server), args.Get(1).(*models.User), args.Error(2)
}

func (m *mockServerRepository) CheckUserInServer(userID, serverID string) (bool, error) {
	args := m.Called(userID, serverID)
	return args.Bool(0), args.Error(1)
}

func (m *mockServerRepository) UpdateServer(serverID string, updates map[string]any) error {
	args := m.Called(serverID, updates)
	return args.Error(0)
}

func (m *mockServerRepository) DeleteServer(serverID string) error {
	args := m.Called(serverID)
	return args.Error(0)
}

func (m *mockServerRepository) GetServerByID(serverID string) (*models.Server, error) {
	args := m.Called(serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Server), args.Error(1)
}

func (m *mockServerRepository) UpdateMemberCount(serverID string, count int) error {
	args := m.Called(serverID, count)
	return args.Error(0)
}

// mockChannelRepository 模擬 ChannelRepository
type mockChannelRepository struct {
	mock.Mock
}

func (m *mockChannelRepository) GetChannelsByServerID(serverID string) ([]models.Channel, error) {
	args := m.Called(serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Channel), args.Error(1)
}

func (m *mockChannelRepository) GetChannelByID(channelID string) (*models.Channel, error) {
	args := m.Called(channelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Channel), args.Error(1)
}

func (m *mockChannelRepository) CreateChannel(channel *models.Channel) error {
	args := m.Called(channel)
	return args.Error(0)
}

func (m *mockChannelRepository) UpdateChannel(channelID string, updates map[string]any) error {
	args := m.Called(channelID, updates)
	return args.Error(0)
}

func (m *mockChannelRepository) DeleteChannel(channelID string) error {
	args := m.Called(channelID)
	return args.Error(0)
}

func (m *mockChannelRepository) CheckChannelExists(channelID string) (bool, error) {
	args := m.Called(channelID)
	return args.Bool(0), args.Error(1)
}

// mockChannelCategoryRepository 模擬 ChannelCategoryRepository
type mockChannelCategoryRepository struct {
	mock.Mock
}

func (m *mockChannelCategoryRepository) CreateChannelCategory(category *models.ChannelCategory) error {
	args := m.Called(category)
	return args.Error(0)
}

func (m *mockChannelCategoryRepository) GetChannelCategoriesByServerID(serverID string) ([]models.ChannelCategory, error) {
	args := m.Called(serverID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ChannelCategory), args.Error(1)
}

func (m *mockChannelCategoryRepository) GetChannelCategoryByID(categoryID string) (*models.ChannelCategory, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ChannelCategory), args.Error(1)
}

func (m *mockChannelCategoryRepository) UpdateChannelCategory(categoryID string, updates map[string]any) error {
	args := m.Called(categoryID, updates)
	return args.Error(0)
}

func (m *mockChannelCategoryRepository) DeleteChannelCategory(categoryID string) error {
	args := m.Called(categoryID)
	return args.Error(0)
}

func (m *mockChannelCategoryRepository) CheckChannelCategoryExists(categoryID string) (bool, error) {
	args := m.Called(categoryID)
	return args.Bool(0), args.Error(1)
}

type mockServerClientManager struct {
	mock.Mock
}

func (m *mockServerClientManager) NewClient(userID string, ws *websocket.Conn) *Client {
	return nil
}

func (m *mockServerClientManager) Register(client *Client) {
}

func (m *mockServerClientManager) Unregister(client *Client) {
}

func (m *mockServerClientManager) GetClient(userID string) (*Client, bool) {
	return nil, false
}

func (m *mockServerClientManager) GetAllClients() map[*Client]bool {
	return nil
}

func (m *mockServerClientManager) IsUserOnline(userID string) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

// --- Tests ---

func TestNewServerService(t *testing.T) {
	mockODM := new(mocks.ODM)
	mockServerRepo := new(mockServerRepository)
	mockServerMemberRepo := new(mocks.ServerMemberRepository)
	mockUserRepo := new(mocks.UserRepository)
	mockChannelRepo := new(mockChannelRepository)
	mockCategoryRepo := new(mockChannelCategoryRepository)
	mockChatRepo := new(mocks.ChatRepository)
	mockFileService := new(mocks.FileUploadService)

	mockClientMgr := new(mockServerClientManager)

	service := NewServerService(
		nil,
		mockODM,
		mockServerRepo,
		mockServerMemberRepo,
		mockUserRepo,
		mockChannelRepo,
		mockCategoryRepo,
		mockChatRepo,
		mockFileService,
		nil,
		mockClientMgr,
	)

	assert.NotNil(t, service)
	assert.Equal(t, mockODM, service.odm)
	assert.Equal(t, mockServerRepo, service.serverRepo)
	assert.Equal(t, mockServerMemberRepo, service.serverMemberRepo)
}

func TestGetUserPictureURL_ServerService(t *testing.T) {
	t.Run("成功獲取頭像URL", func(t *testing.T) {
		pictureID := primitive.NewObjectID()
		mockFileService := new(mocks.FileUploadService)
		mockFileService.On("GetFileURLByID", pictureID.Hex()).Return("https://example.com/avatar.jpg", nil)

		service := &serverService{
			fileUploadService: mockFileService,
		}

		user := &models.User{
			PictureID: pictureID,
		}

		url := service.getUserPictureURL(user)
		assert.Equal(t, "https://example.com/avatar.jpg", url)
		mockFileService.AssertExpectations(t)
	})

	t.Run("用戶無頭像", func(t *testing.T) {
		service := &serverService{}
		user := &models.User{
			PictureID: primitive.ObjectID{},
		}

		url := service.getUserPictureURL(user)
		assert.Empty(t, url)
	})

	t.Run("FileUploadService 為 nil", func(t *testing.T) {
		service := &serverService{
			fileUploadService: nil,
		}

		user := &models.User{
			PictureID: primitive.NewObjectID(),
		}

		url := service.getUserPictureURL(user)
		assert.Empty(t, url)
	})
}

func TestGetServerListResponse(t *testing.T) {
	t.Run("成功獲取伺服器列表", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		mockUserRepo := new(mocks.UserRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)
		mockFileService := new(mocks.FileUploadService)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		imageID := primitive.NewObjectID()

		service := &serverService{
			odm:               mockODM,
			userRepo:          mockUserRepo,
			serverMemberRepo:  mockServerMemberRepo,
			fileUploadService: mockFileService,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
			Username:  "test_user",
		}

		serverMembers := []models.ServerMember{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				ServerID:  serverID,
				UserID:    userID,
			},
		}

		servers := []models.Server{
			{
				BaseModel:   providers.BaseModel{ID: serverID},
				Name:        "Test Server",
				ImageID:     imageID,
				Description: "Test Description",
			},
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(serverMembers, nil).Once()
		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.Server")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.Server)
			*arg = servers
		}).Return(nil).Once()
		mockFileService.On("GetFileURLByID", imageID.Hex()).Return("https://example.com/server.jpg", nil).Once()

		result, msgOpt := service.GetServerListResponse(userID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "Test Server", result[0].Name)
		assert.Equal(t, "https://example.com/server.jpg", result[0].PictureURL)

		mockUserRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
		mockODM.AssertExpectations(t)
		mockFileService.AssertExpectations(t)
	})

	t.Run("用戶不存在", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		userID := primitive.NewObjectID()

		service := &serverService{
			userRepo: mockUserRepo,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(nil, errors.New("user not found")).Once()

		result, msgOpt := service.GetServerListResponse(userID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUserNotFound, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("獲取用戶伺服器列表失敗", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)
		userID := primitive.NewObjectID()

		service := &serverService{
			userRepo:         mockUserRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerMemberRepo.On("GetUserServers", userID.Hex()).Return(nil, errors.New("database error")).Once()

		result, msgOpt := service.GetServerListResponse(userID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestCreateServer(t *testing.T) {
	t.Run("成功創建伺服器（無圖片）", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)
		mockChannelRepo := new(mockChannelRepository)
		mockCategoryRepo := new(mockChannelCategoryRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:            mockUserRepo,
			serverRepo:          mockServerRepo,
			serverMemberRepo:    mockServerMemberRepo,
			channelRepo:         mockChannelRepo,
			channelCategoryRepo: mockCategoryRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
			Username:  "test_user",
		}

		createdServer := models.Server{
			BaseModel:   providers.BaseModel{ID: serverID},
			Name:        "Test Server",
			OwnerID:     userID,
			Description: "This is a test server",
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("CreateServer", mock.AnythingOfType("*models.Server")).Return(createdServer, nil).Once()
		mockServerMemberRepo.On("AddMemberToServer", serverID.Hex(), userID.Hex(), "owner").Return(nil).Once()

		// Mock 創建預設類別和頻道
		mockCategoryRepo.On("CreateChannelCategory", mock.AnythingOfType("*models.ChannelCategory")).Return(nil).Times(2)
		mockCategoryRepo.On("GetChannelCategoriesByServerID", serverID.Hex()).Return([]models.ChannelCategory{
			{
				BaseModel:    providers.BaseModel{ID: primitive.NewObjectID()},
				CategoryType: "text",
			},
			{
				BaseModel:    providers.BaseModel{ID: primitive.NewObjectID()},
				CategoryType: "voice",
			},
		}, nil).Once()
		mockChannelRepo.On("CreateChannel", mock.AnythingOfType("*models.Channel")).Return(nil).Times(2)

		result, msgOpt := service.CreateServer(userID.Hex(), "Test Server", nil, nil)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Server", result.Name)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("用戶不存在", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		userID := primitive.NewObjectID()

		service := &serverService{
			userRepo: mockUserRepo,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(nil, errors.New("user not found")).Once()

		result, msgOpt := service.CreateServer(userID.Hex(), "Test Server", nil, nil)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUserNotFound, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("無效的用戶ID格式", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)

		service := &serverService{
			userRepo: mockUserRepo,
		}

		user := &models.User{
			Username: "test_user",
		}

		mockUserRepo.On("GetUserById", "invalid_id").Return(user, nil).Once()

		result, msgOpt := service.CreateServer("invalid_id", "Test Server", nil, nil)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestSearchPublicServers(t *testing.T) {
	t.Run("成功搜尋公開伺服器", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)
		mockFileService := new(mocks.FileUploadService)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()
		ownerID := primitive.NewObjectID()

		service := &serverService{
			userRepo:          mockUserRepo,
			serverRepo:        mockServerRepo,
			serverMemberRepo:  mockServerMemberRepo,
			fileUploadService: mockFileService,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		owner := &models.User{
			BaseModel: providers.BaseModel{ID: ownerID},
			Username:  "server_owner",
		}

		servers := []models.Server{
			{
				BaseModel:   providers.BaseModel{ID: serverID, CreatedAt: time.Now()},
				Name:        "Public Server",
				Description: "A public server",
				OwnerID:     ownerID,
				MemberCount: 10,
				IsPublic:    true,
			},
		}

		request := models.ServerSearchRequest{
			Query: "Public",
			Page:  1,
			Limit: 10,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("SearchPublicServers", request).Return(servers, int64(1), nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(false, nil).Once()
		mockUserRepo.On("GetUserById", ownerID.Hex()).Return(owner, nil).Once()
		mockFileService.On("GetFileURLByID", mock.Anything).Return("", nil).Maybe()

		result, msgOpt := service.SearchPublicServers(userID.Hex(), request)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result.Servers, 1)
		assert.Equal(t, "Public Server", result.Servers[0].Name)
		assert.False(t, result.Servers[0].IsJoined)
		assert.Equal(t, "server_owner", result.Servers[0].OwnerName)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("用戶不存在", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		userID := primitive.NewObjectID()

		service := &serverService{
			userRepo: mockUserRepo,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(nil, errors.New("user not found")).Once()

		request := models.ServerSearchRequest{}
		result, msgOpt := service.SearchPublicServers(userID.Hex(), request)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrNotFound, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUpdateServer(t *testing.T) {
	t.Run("成功更新伺服器", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockFileService := new(mocks.FileUploadService)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:          mockUserRepo,
			serverRepo:        mockServerRepo,
			fileUploadService: mockFileService,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			Name:      "Old Name",
			OwnerID:   userID,
		}

		updatedServer := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			Name:      "New Name",
			OwnerID:   userID,
		}

		updates := map[string]any{"name": "New Name"}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockServerRepo.On("UpdateServer", serverID.Hex(), updates).Return(nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(updatedServer, nil).Once()
		mockFileService.On("GetFileURLByID", mock.Anything).Return("", nil).Maybe()

		result, msgOpt := service.UpdateServer(userID.Hex(), serverID.Hex(), updates)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "New Name", result.Name)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
	})

	t.Run("無權限更新（非擁有者）", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)

		userID := primitive.NewObjectID()
		ownerID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:   mockUserRepo,
			serverRepo: mockServerRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			OwnerID:   ownerID,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()

		result, msgOpt := service.UpdateServer(userID.Hex(), serverID.Hex(), map[string]any{})

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
	})
}

func TestDeleteServer(t *testing.T) {
	t.Run("成功刪除伺服器", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)
		mockChannelRepo := new(mockChannelRepository)
		mockCategoryRepo := new(mockChannelCategoryRepository)
		mockChatRepo := new(mocks.ChatRepository)
		mockFileService := new(mocks.FileUploadService)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:            mockUserRepo,
			serverRepo:          mockServerRepo,
			serverMemberRepo:    mockServerMemberRepo,
			channelRepo:         mockChannelRepo,
			channelCategoryRepo: mockCategoryRepo,
			chatRepo:            mockChatRepo,
			fileUploadService:   mockFileService,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			OwnerID:   userID,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockServerMemberRepo.On("GetServerMembers", serverID.Hex(), 1, 1000).Return([]models.ServerMember{}, int64(0), nil).Once()
		mockChannelRepo.On("GetChannelsByServerID", serverID.Hex()).Return([]models.Channel{}, nil).Once()
		mockCategoryRepo.On("GetChannelCategoriesByServerID", serverID.Hex()).Return([]models.ChannelCategory{}, nil).Once()
		mockServerRepo.On("DeleteServer", serverID.Hex()).Return(nil).Once()

		msgOpt := service.DeleteServer(userID.Hex(), serverID.Hex())

		assert.Nil(t, msgOpt)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
		mockChannelRepo.AssertExpectations(t)
		mockCategoryRepo.AssertExpectations(t)
	})

	t.Run("無權限刪除（非擁有者）", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)

		userID := primitive.NewObjectID()
		ownerID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:   mockUserRepo,
			serverRepo: mockServerRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			OwnerID:   ownerID,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()

		msgOpt := service.DeleteServer(userID.Hex(), serverID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
	})
}

func TestJoinServer(t *testing.T) {
	t.Run("成功加入伺服器", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:         mockUserRepo,
			serverRepo:       mockServerRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel:  providers.BaseModel{ID: serverID},
			IsPublic:   true,
			MaxMembers: 100,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(false, nil).Once()
		mockServerMemberRepo.On("GetMemberCount", serverID.Hex()).Return(int64(10), nil).Once()
		mockServerMemberRepo.On("AddMemberToServer", serverID.Hex(), userID.Hex(), "member").Return(nil).Once()
		mockServerRepo.On("UpdateMemberCount", serverID.Hex(), 11).Return(nil).Once()

		msgOpt := service.JoinServer(userID.Hex(), serverID.Hex())

		assert.Nil(t, msgOpt)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("伺服器不開放加入", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:   mockUserRepo,
			serverRepo: mockServerRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			IsPublic:  false,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()

		msgOpt := service.JoinServer(userID.Hex(), serverID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrForbidden, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
	})

	t.Run("已經是成員", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:         mockUserRepo,
			serverRepo:       mockServerRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			IsPublic:  true,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(true, nil).Once()

		msgOpt := service.JoinServer(userID.Hex(), serverID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrOperationFailed, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestLeaveServer(t *testing.T) {
	t.Run("成功離開伺服器", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)

		userID := primitive.NewObjectID()
		ownerID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:         mockUserRepo,
			serverRepo:       mockServerRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			OwnerID:   ownerID,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(true, nil).Once()
		mockServerMemberRepo.On("RemoveMemberFromServer", serverID.Hex(), userID.Hex()).Return(nil).Once()
		mockServerMemberRepo.On("GetMemberCount", serverID.Hex()).Return(int64(10), nil).Once()
		mockServerRepo.On("UpdateMemberCount", serverID.Hex(), 10).Return(nil).Once()

		msgOpt := service.LeaveServer(userID.Hex(), serverID.Hex())

		assert.Nil(t, msgOpt)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("擁有者無法離開", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:   mockUserRepo,
			serverRepo: mockServerRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			OwnerID:   userID,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()

		msgOpt := service.LeaveServer(userID.Hex(), serverID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrForbidden, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
	})

	t.Run("不是成員無法離開", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)

		userID := primitive.NewObjectID()
		ownerID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:         mockUserRepo,
			serverRepo:       mockServerRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			OwnerID:   ownerID,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(false, nil).Once()

		msgOpt := service.LeaveServer(userID.Hex(), serverID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrOperationFailed, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestGetServerByID(t *testing.T) {
	t.Run("成功獲取伺服器（作為成員）", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)
		mockFileService := new(mocks.FileUploadService)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:          mockUserRepo,
			serverRepo:        mockServerRepo,
			serverMemberRepo:  mockServerMemberRepo,
			fileUploadService: mockFileService,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel:   providers.BaseModel{ID: serverID},
			Name:        "Test Server",
			Description: "Test",
			IsPublic:    false,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(true, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()
		mockFileService.On("GetFileURLByID", mock.Anything).Return("", nil).Maybe()

		result, msgOpt := service.GetServerByID(userID.Hex(), serverID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Server", result.Name)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})

	t.Run("無權限查看（非成員且私有伺服器）", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepository)
		mockServerRepo := new(mockServerRepository)
		mockServerMemberRepo := new(mocks.ServerMemberRepository)

		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &serverService{
			userRepo:         mockUserRepo,
			serverRepo:       mockServerRepo,
			serverMemberRepo: mockServerMemberRepo,
		}

		user := &models.User{
			BaseModel: providers.BaseModel{ID: userID},
		}

		server := &models.Server{
			BaseModel: providers.BaseModel{ID: serverID},
			IsPublic:  false,
		}

		mockUserRepo.On("GetUserById", userID.Hex()).Return(user, nil).Once()
		mockServerMemberRepo.On("IsMemberOfServer", serverID.Hex(), userID.Hex()).Return(false, nil).Once()
		mockServerRepo.On("GetServerByID", serverID.Hex()).Return(server, nil).Once()

		result, msgOpt := service.GetServerByID(userID.Hex(), serverID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockUserRepo.AssertExpectations(t)
		mockServerRepo.AssertExpectations(t)
		mockServerMemberRepo.AssertExpectations(t)
	})
}

func TestGetServerChannels(t *testing.T) {
	t.Run("成功獲取伺服器頻道列表", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		serverID := primitive.NewObjectID()

		service := &serverService{
			channelRepo: mockChannelRepo,
		}

		channels := []models.Channel{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				Name:      "general",
				Type:      "text",
			},
		}

		mockChannelRepo.On("GetChannelsByServerID", serverID.Hex()).Return(channels, nil).Once()

		result, msgOpt := service.GetServerChannels(serverID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "general", result[0].Name)

		mockChannelRepo.AssertExpectations(t)
	})

	t.Run("獲取頻道列表失敗", func(t *testing.T) {
		mockChannelRepo := new(mockChannelRepository)
		serverID := primitive.NewObjectID()

		service := &serverService{
			channelRepo: mockChannelRepo,
		}

		mockChannelRepo.On("GetChannelsByServerID", serverID.Hex()).Return(nil, errors.New("database error")).Once()

		result, msgOpt := service.GetServerChannels(serverID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockChannelRepo.AssertExpectations(t)
	})
}
