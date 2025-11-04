package services

import (
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Tests ---

// TestNewChatService 測試創建 ChatService
func TestNewChatService(t *testing.T) {
	// 使用 nil 進行簡化測試
	service := NewChatService(
		nil, // config
		nil, // odm
		nil, // redis
		nil, // cache
		nil, // chatRepo
		nil, // serverRepo
		nil, // serverMemberRepo
		nil, // userRepo
		nil, // userService
		nil, // fileService
	)

	assert.NotNil(t, service, "服務應該被成功創建")

	// 驗證內部組件被正確初始化
	cs := service.(*chatService)
	assert.NotNil(t, cs.clientManager, "ClientManager 應該被初始化")
	assert.NotNil(t, cs.roomManager, "RoomManager 應該被初始化")
	assert.NotNil(t, cs.messageHandler, "MessageHandler 應該被初始化")
	assert.NotNil(t, cs.websocketHandler, "WebSocketHandler 應該被初始化")
}

// TestChatService_Structure 測試 ChatService 結構
func TestChatService_Structure(t *testing.T) {
	service := NewChatService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	cs, ok := service.(*chatService)
	assert.True(t, ok, "服務應該可以轉換為 chatService 類型")

	// 驗證結構體包含預期的欄位
	assert.NotNil(t, cs.clientManager, "應該有 clientManager 欄位")
	assert.NotNil(t, cs.roomManager, "應該有 roomManager 欄位")
	assert.NotNil(t, cs.messageHandler, "應該有 messageHandler 欄位")
	assert.NotNil(t, cs.websocketHandler, "應該有 websocketHandler 欄位")
}

func TestGetDMRoomResponseList(t *testing.T) {
	t.Run("成功獲取聊天列表", func(t *testing.T) {
		mockChatRepo := new(mocks.ChatRepository)
		mockUserRepo := new(mocks.UserRepository)
		mockClientManager := new(mockClientManager)

		userID := primitive.NewObjectID()
		chatWithUserID := primitive.NewObjectID()
		roomID := primitive.NewObjectID()

		service := &chatService{
			chatRepo:      mockChatRepo,
			userRepo:      mockUserRepo,
			clientManager: mockClientManager,
		}

		// 模擬聊天列表
		dmRooms := []models.DMRoom{
			{
				RoomID:         roomID,
				UserID:         userID,
				ChatWithUserID: chatWithUserID,
				IsHidden:       false,
			},
		}

		// 模擬用戶資料
		users := []models.User{
			{
				BaseModel: providers.BaseModel{ID: chatWithUserID},
				Nickname:  "TestUser",
			},
		}

		mockChatRepo.On("GetDMRoomListByUserID", userID.Hex(), false).Return(dmRooms, nil).Once()
		mockUserRepo.On("GetUserListByIds", []string{chatWithUserID.Hex()}).Return(users, nil).Once()
		mockClientManager.On("IsUserOnline", chatWithUserID.Hex()).Return(true).Once()

		result, msgOpt := service.GetDMRoomResponseList(userID.Hex(), false)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, roomID, result[0].RoomID)
		assert.Equal(t, "TestUser", result[0].Nickname)
		assert.True(t, result[0].IsOnline)

		mockChatRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockClientManager.AssertExpectations(t)
	})

	t.Run("獲取聊天列表失敗", func(t *testing.T) {
		mockChatRepo := new(mocks.ChatRepository)
		userID := primitive.NewObjectID()

		service := &chatService{
			chatRepo: mockChatRepo,
		}

		mockChatRepo.On("GetDMRoomListByUserID", userID.Hex(), false).Return(nil, errors.New("database error")).Once()

		result, msgOpt := service.GetDMRoomResponseList(userID.Hex(), false)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockChatRepo.AssertExpectations(t)
	})

	t.Run("獲取用戶資訊失敗", func(t *testing.T) {
		mockChatRepo := new(mocks.ChatRepository)
		mockUserRepo := new(mocks.UserRepository)

		userID := primitive.NewObjectID()
		chatWithUserID := primitive.NewObjectID()

		service := &chatService{
			chatRepo: mockChatRepo,
			userRepo: mockUserRepo,
		}

		dmRooms := []models.DMRoom{
			{
				ChatWithUserID: chatWithUserID,
			},
		}

		mockChatRepo.On("GetDMRoomListByUserID", userID.Hex(), false).Return(dmRooms, nil).Once()
		mockUserRepo.On("GetUserListByIds", []string{chatWithUserID.Hex()}).Return(nil, errors.New("user fetch error")).Once()

		result, msgOpt := service.GetDMRoomResponseList(userID.Hex(), false)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockChatRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUpdateDMRoom(t *testing.T) {
	t.Run("成功更新聊天房間狀態", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		roomID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		dmRoom := models.DMRoom{
			BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
			RoomID:    roomID,
			UserID:    userID,
			IsHidden:  false,
		}

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(dmRoom, nil).Once()
		mockODM.On("Update", mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil).Once()

		msgOpt := service.UpdateDMRoom(userID.Hex(), roomID.Hex(), true)

		assert.Nil(t, msgOpt)
		mockODM.AssertExpectations(t)
	})

	t.Run("無效的用戶ID格式", func(t *testing.T) {
		service := &chatService{}

		msgOpt := service.UpdateDMRoom("invalid_id", primitive.NewObjectID().Hex(), true)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})

	t.Run("無效的房間ID格式", func(t *testing.T) {
		service := &chatService{}
		userID := primitive.NewObjectID()

		msgOpt := service.UpdateDMRoom(userID.Hex(), "invalid_id", true)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})

	t.Run("聊天房間不存在", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		roomID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil, providers.ErrDocumentNotFound).Once()

		msgOpt := service.UpdateDMRoom(userID.Hex(), roomID.Hex(), true)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrRoomNotFound, msgOpt.Code)
		mockODM.AssertExpectations(t)
	})
}

func TestCreateDMRoom(t *testing.T) {
	t.Run("創建新的聊天房間（雙方都沒有）", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		chatWithUserID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		// 模擬查找用戶 - 返回 nil error
		mockODM.On("FindByID", mock.Anything, chatWithUserID.Hex(), mock.AnythingOfType("*models.User")).Return(nil).Once()

		// 模擬查詢現有房間（沒有找到）
		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.DMRoom)
			*arg = []models.DMRoom{}
		}).Return(nil).Once()

		// 模擬創建房間
		mockODM.On("Create", mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil).Once()

		result, msgOpt := service.CreateDMRoom(userID.Hex(), chatWithUserID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		// 注意：由於 FindByID 不會填充 user，我們無法驗證 Nickname

		mockODM.AssertExpectations(t)
	})

	t.Run("雙方都已有房間記錄", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		chatWithUserID := primitive.NewObjectID()
		roomID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("FindByID", mock.Anything, chatWithUserID.Hex(), mock.AnythingOfType("*models.User")).Return(nil).Once()

		dmRooms := []models.DMRoom{
			{
				RoomID:         roomID,
				UserID:         userID,
				ChatWithUserID: chatWithUserID,
				IsHidden:       false,
			},
			{
				RoomID:         roomID,
				UserID:         chatWithUserID,
				ChatWithUserID: userID,
				IsHidden:       false,
			},
		}

		mockODM.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.DMRoom")).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.DMRoom)
			*arg = dmRooms
		}).Return(nil).Once()

		result, msgOpt := service.CreateDMRoom(userID.Hex(), chatWithUserID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, roomID, result.RoomID)

		mockODM.AssertExpectations(t)
	})

	t.Run("無效的用戶ID格式", func(t *testing.T) {
		service := &chatService{}

		result, msgOpt := service.CreateDMRoom("invalid_id", primitive.NewObjectID().Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})

	t.Run("聊天對象不存在", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		chatWithUserID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("FindByID", mock.Anything, chatWithUserID.Hex(), mock.AnythingOfType("*models.User")).Return(providers.ErrDocumentNotFound).Once()

		result, msgOpt := service.CreateDMRoom(userID.Hex(), chatWithUserID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUserNotFound, msgOpt.Code)

		mockODM.AssertExpectations(t)
	})
}

func TestGetDMMessages(t *testing.T) {
	t.Run("成功獲取私聊訊息", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		roomID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		dmRoom := models.DMRoom{
			RoomID: roomID,
			UserID: userID,
		}

		messages := []models.Message{
			{
				BaseModel: providers.BaseModel{
					ID:        primitive.NewObjectID(),
					UpdatedAt: time.Now(),
				},
				RoomType: models.RoomTypeDM,
				RoomID:   roomID,
				SenderID: userID,
				Content:  "Hello",
			},
		}

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(dmRoom, nil).Once()
		mockODM.On("FindWithOptions", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.Message"), mock.Anything).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.Message)
			*arg = messages
		}).Return(nil).Once()

		result, msgOpt := service.GetDMMessages(userID.Hex(), roomID.Hex(), "", "", "50")

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "Hello", result[0].Content)

		mockODM.AssertExpectations(t)
	})

	t.Run("無效的用戶ID", func(t *testing.T) {
		service := &chatService{}

		result, msgOpt := service.GetDMMessages("invalid_id", primitive.NewObjectID().Hex(), "", "", "")

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})

	t.Run("房間不存在", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		roomID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("FindOne", mock.Anything, mock.Anything, mock.AnythingOfType("*models.DMRoom")).Return(nil, providers.ErrDocumentNotFound).Once()

		result, msgOpt := service.GetDMMessages(userID.Hex(), roomID.Hex(), "", "", "")

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrRoomNotFound, msgOpt.Code)

		mockODM.AssertExpectations(t)
	})
}

func TestGetChannelMessages(t *testing.T) {
	t.Run("成功獲取頻道訊息", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		messages := []models.Message{
			{
				BaseModel: providers.BaseModel{
					ID:        primitive.NewObjectID(),
					UpdatedAt: time.Now(),
				},
				RoomType: models.RoomTypeChannel,
				RoomID:   channelID,
				SenderID: userID,
				Content:  "Channel message",
			},
		}

		mockODM.On("FindByID", mock.Anything, channelID.Hex(), mock.AnythingOfType("*models.Channel")).Return(nil).Once()
		mockODM.On("Exists", mock.Anything, mock.Anything, mock.AnythingOfType("*models.ServerMember")).Return(true, nil).Once()
		mockODM.On("FindWithOptions", mock.Anything, mock.Anything, mock.AnythingOfType("*[]models.Message"), mock.Anything).Run(func(args mock.Arguments) {
			arg := args.Get(2).(*[]models.Message)
			*arg = messages
		}).Return(nil).Once()

		result, msgOpt := service.GetChannelMessages(userID.Hex(), channelID.Hex(), "", "", "")

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		assert.Equal(t, "Channel message", result[0].Content)

		mockODM.AssertExpectations(t)
	})

	t.Run("頻道不存在", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("FindByID", mock.Anything, channelID.Hex(), mock.AnythingOfType("*models.Channel")).Return(providers.ErrDocumentNotFound).Once()

		result, msgOpt := service.GetChannelMessages(userID.Hex(), channelID.Hex(), "", "", "")

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrChannelNotFound, msgOpt.Code)

		mockODM.AssertExpectations(t)
	})

	t.Run("用戶無權限訪問頻道", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		channelID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("FindByID", mock.Anything, channelID.Hex(), mock.AnythingOfType("*models.Channel")).Return(nil).Once()
		mockODM.On("Exists", mock.Anything, mock.Anything, mock.AnythingOfType("*models.ServerMember")).Return(false, nil).Once()

		result, msgOpt := service.GetChannelMessages(userID.Hex(), channelID.Hex(), "", "", "")

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)

		mockODM.AssertExpectations(t)
	})
}

func TestCheckUserServerMembership(t *testing.T) {
	t.Run("用戶是伺服器成員", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("Exists", mock.Anything, mock.Anything, mock.AnythingOfType("*models.ServerMember")).Return(true, nil).Once()

		isMember, err := service.checkUserServerMembership(userID.Hex(), serverID.Hex())

		assert.NoError(t, err)
		assert.True(t, isMember)

		mockODM.AssertExpectations(t)
	})

	t.Run("用戶不是伺服器成員", func(t *testing.T) {
		mockODM := new(mocks.ODM)
		userID := primitive.NewObjectID()
		serverID := primitive.NewObjectID()

		service := &chatService{
			odm: mockODM,
		}

		mockODM.On("Exists", mock.Anything, mock.Anything, mock.AnythingOfType("*models.ServerMember")).Return(false, nil).Once()

		isMember, err := service.checkUserServerMembership(userID.Hex(), serverID.Hex())

		assert.NoError(t, err)
		assert.False(t, isMember)

		mockODM.AssertExpectations(t)
	})

	t.Run("無效的用戶ID", func(t *testing.T) {
		service := &chatService{}

		isMember, err := service.checkUserServerMembership("invalid_id", primitive.NewObjectID().Hex())

		assert.Error(t, err)
		assert.False(t, isMember)
	})
}

func TestChatService_GetUserPictureURL(t *testing.T) {
	t.Run("成功獲取頭像URL", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)
		pictureID := primitive.NewObjectID()

		service := &chatService{
			fileUploadService: mockFileService,
		}

		user := &models.User{
			PictureID: pictureID,
		}

		mockFileService.On("GetFileURLByID", pictureID.Hex()).Return("https://example.com/avatar.jpg", nil).Once()

		url := service.getUserPictureURL(user)

		assert.Equal(t, "https://example.com/avatar.jpg", url)
		mockFileService.AssertExpectations(t)
	})

	t.Run("用戶無頭像", func(t *testing.T) {
		service := &chatService{}

		user := &models.User{
			PictureID: primitive.ObjectID{},
		}

		url := service.getUserPictureURL(user)

		assert.Empty(t, url)
	})

	t.Run("FileService 為 nil", func(t *testing.T) {
		service := &chatService{
			fileUploadService: nil,
		}

		user := &models.User{
			PictureID: primitive.NewObjectID(),
		}

		url := service.getUserPictureURL(user)

		assert.Empty(t, url)
	})
}
