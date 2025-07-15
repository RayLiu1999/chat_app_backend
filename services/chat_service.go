package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/utils"
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ChatService 管理所有的聊天功能
type ChatService struct {
	config       *config.Config
	mongoConnect *mongo.Database
	redisClient  *redis.Client
	chatRepo     repositories.ChatRepositoryInterface
	serverRepo   repositories.ServerRepositoryInterface
	userRepo     repositories.UserRepositoryInterface
	odm          *providers.ODM

	// 新增的模組化組件
	clientManager    *ClientManager
	roomManager      *RoomManager
	messageHandler   *MessageHandler
	websocketHandler *WebSocketHandler
}

// NewChatService 初始化聊天室服務
func NewChatService(cfg *config.Config, mongodb *mongo.Database, chatRepo repositories.ChatRepositoryInterface, serverRepo repositories.ServerRepositoryInterface, userRepo repositories.UserRepositoryInterface) *ChatService {
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
		utils.PrettyPrint("Failed to ping Redis:", err)
		return nil
	}

	// 創建ODM實例
	odm := providers.NewODM(mongodb)

	// 創建模組化組件
	clientManager := NewClientManager(redisClient)
	roomManager := NewRoomManager(mongodb, redisClient)
	messageHandler := NewMessageHandler(mongodb, roomManager)
	websocketHandler := NewWebSocketHandler(mongodb, clientManager, roomManager, messageHandler)

	cs := &ChatService{
		config:           cfg,
		mongoConnect:     mongodb,
		redisClient:      redisClient,
		chatRepo:         chatRepo,
		serverRepo:       serverRepo,
		userRepo:         userRepo,
		odm:              odm,
		clientManager:    clientManager,
		roomManager:      roomManager,
		messageHandler:   messageHandler,
		websocketHandler: websocketHandler,
	}

	return cs
}

// HandleWebSocket 處理 WebSocket 連線
func (cs *ChatService) HandleWebSocket(ws *websocket.Conn, userID string) {
	cs.websocketHandler.HandleWebSocket(ws, userID)
	utils.PrettyPrintf("WebSocket connection established for user: %s", userID)
}

// 取得聊天記錄response
func (cs *ChatService) GetDMRoomResponseList(userID string, includeHidden bool) ([]models.DMRoomResponse, error) {
	chatList, err := cs.chatRepo.GetDMRoomListByUserID(userID, includeHidden)
	if err != nil {
		return nil, err
	}

	var userIds []string
	for _, chat := range chatList {
		userIds = append(userIds, chat.ChatWithUserID.Hex())
	}

	// 取得用戶id陣列
	userList, err := cs.userRepo.GetUserListByIds(userIds)
	if err != nil {
		return nil, err
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
		chatResponseList = append(chatResponseList, models.DMRoomResponse{
			RoomID:    chat.RoomID,
			Nickname:  user.Nickname,
			Picture:   user.Picture,
			Timestamp: chat.UpdatedAt.Unix(),
		})
	}

	return chatResponseList, nil
}

// UpdateDMRoom 更新聊天房間狀態
func (cs *ChatService) UpdateDMRoom(userID string, roomID string, isHidden bool) error {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return err
	}

	// 使用QueryBuilder構建查詢
	qb := providers.NewQueryBuilder()
	qb.Where("room_id", roomObjectID).Where("user_id", userObjectID)

	// 檢查room_id是否存在
	var dmRoom models.DMRoom
	err = cs.odm.FindOne(context.Background(), qb.GetFilter(), &dmRoom)
	if err != nil {
		return err
	}

	// 更新狀態
	updateFields := bson.M{"is_hidden": isHidden}
	return cs.odm.UpdateFields(context.Background(), &dmRoom, updateFields)
}

// CreateDMRoom 創建私聊房間
func (cs *ChatService) CreateDMRoom(userID string, chatWithUserID string) (*models.DMRoomResponse, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	chatWithUserObjectID, err := primitive.ObjectIDFromHex(chatWithUserID)
	if err != nil {
		return nil, err
	}

	// 檢查chat_with_user_id是否存在
	var user models.User
	err = cs.odm.FindByID(context.Background(), chatWithUserID, &user)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// 定義回傳格式
	var dmRoomResponse models.DMRoomResponse

	// 如果雙方都建立則直接回傳
	if len(roomList) == 2 {
		for _, room := range roomList {
			if room.UserID == userObjectID {
				dmRoomResponse = models.DMRoomResponse{
					RoomID:    room.RoomID,
					Nickname:  user.Nickname,
					Picture:   user.Picture,
					Timestamp: room.UpdatedAt.Unix(),
				}

				// 如果isHidden為true，則將isHidden設為false
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
				return nil, err
			}

			dmRoomResponse = models.DMRoomResponse{
				RoomID:    room.RoomID,
				Nickname:  user.Nickname,
				Picture:   user.Picture,
				Timestamp: dmRoom.UpdatedAt.Unix(),
			}

			return &dmRoomResponse, nil
		}

		// 如果自己建立過則直接回傳
		if room.UserID == userObjectID {
			dmRoomResponse = models.DMRoomResponse{
				RoomID:    room.RoomID,
				Nickname:  user.Nickname,
				Picture:   user.Picture,
				Timestamp: room.UpdatedAt.Unix(),
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
			return nil, err
		}

		dmRoomResponse = models.DMRoomResponse{
			RoomID:    dmRoom.RoomID,
			Nickname:  user.Nickname,
			Picture:   user.Picture,
			Timestamp: dmRoom.UpdatedAt.Unix(),
		}

		return &dmRoomResponse, nil
	}

	return nil, errors.New("unknown error occurred")
}

// GetDMMessages 獲取私聊訊息
func (cs *ChatService) GetDMMessages(userID string, roomID string, before string, after string, limit string) ([]models.MessageResponse, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, err
	}

	// 檢查room_id是否存在
	qb := providers.NewQueryBuilder()
	qb.Where("room_id", roomObjectID).Where("user_id", userObjectID)

	var room models.DMRoom
	err = cs.odm.FindOne(context.Background(), qb.GetFilter(), &room)
	if err != nil {
		return nil, err
	}

	// 構建訊息查詢
	messageQb := providers.NewQueryBuilder()
	messageQb.Where("room_id", roomObjectID)

	if before != "" {
		beforeObjectID, err := primitive.ObjectIDFromHex(before)
		if err != nil {
			return nil, err
		}
		messageQb.WhereLt("_id", beforeObjectID)
	}

	if after != "" {
		afterObjectID, err := primitive.ObjectIDFromHex(after)
		if err != nil {
			return nil, err
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
		return nil, err
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
