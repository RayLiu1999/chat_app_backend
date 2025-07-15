package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/repositories"
	"chat_app_backend/utils"
	"context"
	"log"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
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
