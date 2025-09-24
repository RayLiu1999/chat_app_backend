package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/app/repositories"
	"chat_app_backend/config"
	"context"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatService 管理所有的聊天功能
type ChatService struct {
	config            *config.Config
	redisClient       *redis.Client
	chatRepo          repositories.ChatRepositoryInterface
	serverRepo        repositories.ServerRepositoryInterface
	serverMemberRepo  repositories.ServerMemberRepositoryInterface
	userRepo          repositories.UserRepositoryInterface
	odm               *providers.ODM
	userService       UserServiceInterface
	fileUploadService FileUploadServiceInterface // 添加 FileUploadService 依賴

	// 新增的模組化組件
	clientManager    *ClientManager
	roomManager      *RoomManager
	messageHandler   *MessageHandler
	websocketHandler *WebSocketHandler
}

// NewChatService 初始化聊天室服務
func NewChatService(cfg *config.Config,
	odm *providers.ODM,
	redisClient *redis.Client,
	chatRepo repositories.ChatRepositoryInterface,
	serverRepo repositories.ServerRepositoryInterface,
	serverMemberRepo repositories.ServerMemberRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	userService UserServiceInterface,
	fileUploadService FileUploadServiceInterface) *ChatService {

	// 創建模組化組件
	clientManager := NewClientManager(redisClient)
	roomManager := NewRoomManager(odm, redisClient, serverMemberRepo)
	messageHandler := NewMessageHandler(odm, roomManager)
	websocketHandler := NewWebSocketHandler(odm, clientManager, roomManager, messageHandler, userService)

	cs := &ChatService{
		config:            cfg,
		redisClient:       redisClient,
		chatRepo:          chatRepo,
		serverRepo:        serverRepo,
		serverMemberRepo:  serverMemberRepo,
		userRepo:          userRepo,
		odm:               odm,
		userService:       userService,
		fileUploadService: fileUploadService,
		clientManager:     clientManager,
		roomManager:       roomManager,
		messageHandler:    messageHandler,
		websocketHandler:  websocketHandler,
	}

	return cs
}

// getUserPictureURL 獲取用戶頭像 URL（從 ObjectID 解析）
func (cs *ChatService) getUserPictureURL(user *models.User) string {
	if user.PictureID.IsZero() || cs.fileUploadService == nil {
		return ""
	}

	pictureURL, err := cs.fileUploadService.GetFileURLByID(user.PictureID.Hex())
	if err != nil {
		return ""
	}
	return pictureURL
}

// HandleWebSocket 處理 WebSocket 連線
func (cs *ChatService) HandleWebSocket(ws *websocket.Conn, userID string) {
	cs.websocketHandler.HandleWebSocket(ws, userID)
}

// GetClientManager 獲取客戶端管理器
func (cs *ChatService) GetClientManager() *ClientManager {
	return cs.clientManager
}

// UpdateUserService 更新 UserService 引用
func (cs *ChatService) UpdateUserService(userService UserServiceInterface) {
	cs.userService = userService
	cs.websocketHandler.userService = userService
}

// 取得聊天記錄response
func (cs *ChatService) GetDMRoomResponseList(userID string, includeHidden bool) ([]models.DMRoomResponse, *models.MessageOptions) {
	chatList, err := cs.chatRepo.GetDMRoomListByUserID(userID, includeHidden)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取聊天列表失敗",
		}
	}

	var userIds []string
	for _, chat := range chatList {
		userIds = append(userIds, chat.ChatWithUserID.Hex())
	}

	// 取得用戶id陣列
	userList, err := cs.userRepo.GetUserListByIds(userIds)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取用戶資訊失敗",
		}
	}

	userListById := make(map[string]models.User)
	for _, user := range userList {
		userListById[user.ID.Hex()] = user
	}

	// 轉換為 ChatResponse 格式
	chatResponseList := []models.DMRoomResponse{}
	for _, chat := range chatList {
		user, ok := userListById[chat.ChatWithUserID.Hex()]
		if !ok {
			continue
		}

		// 檢查用戶在線狀態
		isOnline := false
		if cs.clientManager != nil {
			isOnline = cs.clientManager.IsUserOnline(chat.ChatWithUserID.Hex())
		}

		chatResponseList = append(chatResponseList, models.DMRoomResponse{
			RoomID:     chat.RoomID,
			Nickname:   user.Nickname,
			PictureURL: cs.getUserPictureURL(&user),
			Timestamp:  chat.UpdatedAt.UnixMilli(),
			IsOnline:   isOnline,
		})
	}

	return chatResponseList, nil
}

// UpdateDMRoom 更新聊天房間狀態
func (cs *ChatService) UpdateDMRoom(userID string, roomID string, isHidden bool) *models.MessageOptions {
	// 使用ODM直接操作
	ctx := context.Background()

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的用戶ID格式",
		}
	}

	roomObjID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的房間ID格式",
		}
	}

	var dmRoom models.DMRoom
	filter := map[string]any{
		"user_id": userObjID,
		"room_id": roomObjID,
	}

	err = cs.odm.FindOne(ctx, filter, &dmRoom)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrRoomNotFound,
			Details: err,
			Message: "聊天房間不存在",
		}
	}

	dmRoom.IsHidden = isHidden
	dmRoom.UpdatedAt = time.Now()

	err = cs.odm.Update(ctx, &dmRoom)
	if err != nil {
		return &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "更新聊天房間狀態失敗",
		}
	}

	return nil
}

// CreateDMRoom 創建私聊房間
func (cs *ChatService) CreateDMRoom(userID string, chatWithUserID string) (*models.DMRoomResponse, *models.MessageOptions) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的用戶ID格式",
		}
	}

	chatWithUserObjectID, err := primitive.ObjectIDFromHex(chatWithUserID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的聊天對象ID格式",
		}
	}

	// 檢查chat_with_user_id是否存在
	var user models.User
	err = cs.odm.FindByID(context.Background(), chatWithUserID, &user)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrUserNotFound,
			Details: err,
			Message: "聊天對象不存在",
		}
	}

	// 檢查房間是否存在
	qb := providers.NewQueryBuilder()
	orConditions := []bson.M{
		{
			"user_id":           chatWithUserObjectID,
			"chat_with_user_id": userObjectID,
		},
		{
			"user_id":           userObjectID,
			"chat_with_user_id": chatWithUserObjectID,
		},
	}
	qb.OrWhere(orConditions)

	var roomList []models.DMRoom
	err = cs.odm.Find(context.Background(), qb.GetFilter(), &roomList)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "查詢聊天房間失敗",
		}
	}

	// 定義回傳格式
	var dmRoomResponse models.DMRoomResponse

	// 如果雙方都建立則直接回傳
	if len(roomList) == 2 {
		for _, room := range roomList {
			if room.UserID == userObjectID {
				dmRoomResponse = models.DMRoomResponse{
					RoomID:     room.RoomID,
					Nickname:   user.Nickname,
					PictureURL: cs.getUserPictureURL(&user),
					Timestamp:  room.UpdatedAt.UnixMilli(),
				} // 如果isHidden為true，則將isHidden設為false
				if room.IsHidden {
					updateFields := bson.M{"is_hidden": false}
					cs.odm.UpdateFields(context.Background(), &room, updateFields)
				}
				break
			}
		}

		return &dmRoomResponse, nil
	}

	// 如果只有一邊建立過
	if len(roomList) == 1 {
		room := roomList[0]
		// 如果對方建立過，自己沒建過，則取得對方RoomID，並建立user_id為自己的房間
		if room.ChatWithUserID == userObjectID {
			// 建立user_id為自己的房間
			dmRoom := models.DMRoom{
				RoomID:         room.RoomID,
				UserID:         userObjectID,
				ChatWithUserID: chatWithUserObjectID,
				IsHidden:       false,
			}

			err := cs.odm.Create(context.Background(), &dmRoom)
			if err != nil {
				return nil, &models.MessageOptions{
					Code:    models.ErrInternalServer,
					Details: err,
					Message: "創建聊天房間失敗",
				}
			}

			dmRoomResponse = models.DMRoomResponse{
				RoomID:     room.RoomID,
				Nickname:   user.Nickname,
				PictureURL: cs.getUserPictureURL(&user),
				Timestamp:  dmRoom.UpdatedAt.UnixMilli(),
			}

			return &dmRoomResponse, nil
		}

		// 如果自己建立過則直接回傳
		if room.UserID == userObjectID {
			dmRoomResponse = models.DMRoomResponse{
				RoomID:     room.RoomID,
				Nickname:   user.Nickname,
				PictureURL: cs.getUserPictureURL(&user),
				Timestamp:  room.UpdatedAt.UnixMilli(),
			}

			return &dmRoomResponse, nil
		}
	}

	// 如果雙方都沒有建立過
	if len(roomList) == 0 {
		// 建立房間
		dmRoom := models.DMRoom{
			RoomID:         primitive.NewObjectID(),
			UserID:         userObjectID,
			ChatWithUserID: chatWithUserObjectID,
			IsHidden:       false,
		}

		err := cs.odm.Create(context.Background(), &dmRoom)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInternalServer,
				Details: err,
				Message: "創建聊天房間失敗",
			}
		}

		dmRoomResponse = models.DMRoomResponse{
			RoomID:     dmRoom.RoomID,
			Nickname:   user.Nickname,
			PictureURL: cs.getUserPictureURL(&user),
			Timestamp:  dmRoom.UpdatedAt.UnixMilli(),
		}

		return &dmRoomResponse, nil
	}

	return nil, &models.MessageOptions{
		Code:    models.ErrInternalServer,
		Message: "未知錯誤",
	}
}

// GetDMMessages 獲取私聊訊息
func (cs *ChatService) GetDMMessages(userID string, roomID string, before string, after string, limit string) ([]models.MessageResponse, *models.MessageOptions) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的用戶ID格式",
		}
	}

	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的房間ID格式",
		}
	}

	// 檢查room_id是否存在
	qb := providers.NewQueryBuilder()
	qb.Where("room_id", roomObjectID).Where("user_id", userObjectID)

	var room models.DMRoom
	err = cs.odm.FindOne(context.Background(), qb.GetFilter(), &room)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrRoomNotFound,
			Details: err,
			Message: "聊天房間不存在",
		}
	}

	// 構建訊息查詢
	messageQb := providers.NewQueryBuilder()
	messageQb.Where("room_id", roomObjectID)

	if before != "" {
		beforeObjectID, err := primitive.ObjectIDFromHex(before)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Details: err,
				Message: "無效的before參數格式",
			}
		}
		messageQb.WhereLt("_id", beforeObjectID)
	}

	if after != "" {
		afterObjectID, err := primitive.ObjectIDFromHex(after)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Details: err,
				Message: "無效的after參數格式",
			}
		}
		messageQb.WhereGt("_id", afterObjectID)
	}

	messageQb.SortDesc("_id")

	if limit != "" {
		limitVal, err := strconv.ParseInt(limit, 10, 64)
		if err == nil && limitVal > 0 {
			messageQb.Limit(limitVal)
		}
	}

	var messageList []models.Message
	err = cs.odm.FindWithOptions(context.Background(), messageQb.GetFilter(), &messageList, messageQb.GetQueryOptions())
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取訊息失敗",
		}
	}

	var messageResponse []models.MessageResponse
	for _, message := range messageList {
		messageResponse = append(messageResponse, models.MessageResponse{
			ID:        message.ID,
			RoomType:  models.RoomType(message.RoomType),
			RoomID:    message.RoomID.Hex(),
			SenderID:  message.SenderID.Hex(),
			Content:   message.Content,
			Timestamp: message.UpdatedAt.UnixMilli(),
		})
	}

	return messageResponse, nil
}

// GetChannelMessages 獲取頻道訊息
func (cs *ChatService) GetChannelMessages(userID string, channelID string, before string, after string, limit string) ([]models.MessageResponse, *models.MessageOptions) {
	channelObjectID, err := primitive.ObjectIDFromHex(channelID)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Details: err,
			Message: "無效的頻道ID格式",
		}
	}

	// 首先檢查頻道是否存在
	var channel models.Channel
	err = cs.odm.FindByID(context.Background(), channelID, &channel)
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrChannelNotFound,
			Details: err,
			Message: "頻道不存在",
		}
	}

	// 檢查用戶是否有權限訪問此頻道（檢查是否為伺服器成員）
	// 這裡我們需要檢查用戶是否是該伺服器的成員
	isMember, err := cs.checkUserServerMembership(userID, channel.ServerID.Hex())
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "檢查伺服器成員身份失敗",
		}
	}
	if !isMember {
		return nil, &models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "您沒有權限訪問此頻道",
		}
	}

	// 構建訊息查詢
	messageQb := providers.NewQueryBuilder()
	messageQb.Where("room_id", channelObjectID).Where("room_type", string(models.RoomTypeChannel))

	if before != "" {
		beforeObjectID, err := primitive.ObjectIDFromHex(before)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Details: err,
				Message: "無效的before參數格式",
			}
		}
		messageQb.WhereLt("_id", beforeObjectID)
	}

	if after != "" {
		afterObjectID, err := primitive.ObjectIDFromHex(after)
		if err != nil {
			return nil, &models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Details: err,
				Message: "無效的after參數格式",
			}
		}
		messageQb.WhereGt("_id", afterObjectID)
	}

	messageQb.SortDesc("_id")

	if limit != "" {
		limitVal, err := strconv.ParseInt(limit, 10, 64)
		if err == nil && limitVal > 0 {
			// 限制最大獲取數量為100
			if limitVal > 100 {
				limitVal = 100
			}
			messageQb.Limit(limitVal)
		}
	} else {
		// 如果沒有指定 limit，默認返回最近 50 條訊息
		messageQb.Limit(50)
	}

	var messageList []models.Message
	err = cs.odm.FindWithOptions(context.Background(), messageQb.GetFilter(), &messageList, messageQb.GetQueryOptions())
	if err != nil {
		return nil, &models.MessageOptions{
			Code:    models.ErrInternalServer,
			Details: err,
			Message: "獲取頻道訊息失敗",
		}
	}

	// 轉換為響應格式
	var messageResponse []models.MessageResponse
	for _, message := range messageList {
		messageResponse = append(messageResponse, models.MessageResponse{
			ID:        message.ID,
			RoomType:  models.RoomType(message.RoomType),
			RoomID:    message.RoomID.Hex(),
			SenderID:  message.SenderID.Hex(),
			Content:   message.Content,
			Timestamp: message.UpdatedAt.UnixMilli(),
		})
	}

	return messageResponse, nil
}

// checkUserServerMembership 檢查用戶是否為伺服器成員
func (cs *ChatService) checkUserServerMembership(userID, serverID string) (bool, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return false, err
	}

	// 檢查 server_members 集合
	filter := bson.M{
		"user_id":   userObjectID,
		"server_id": serverObjectID,
	}

	return cs.odm.Exists(context.Background(), filter, &models.ServerMember{})
}
