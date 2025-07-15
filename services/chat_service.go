package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/repositories"
	"chat_app_backend/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// WebSocket 消息結構
type WsMessage[T any] struct {
	Action string `json:"action"`
	Data   T      `json:"data"`
}

// MessageResponse 定義聊天室消息
type MessageResponse struct {
	RoomType  models.RoomType `json:"room_type"`
	RoomID    string          `json:"room_id"`
	SenderID  string          `json:"sender_id"`
	Content   string          `json:"content"`
	Timestamp int64           `json:"timestamp"`
}

// Notification 定義通知結構
type Notification struct {
	Action      string          `json:"action"`
	RoomType    models.RoomType `json:"room_type"`
	RoomID      string          `json:"room_id"`
	Message     string          `json:"message"`
	UnreadCount int             `json:"unread_count"`
}

// Client 定義 WebSocket 客戶端
type Client struct {
	UserID        string
	Conn          *websocket.Conn
	Subscribed    map[string]bool
	SubscribedMux sync.RWMutex
	// 新增房間活躍時間追蹤
	RoomActivity  map[string]time.Time // 房間ID -> 最後活躍時間
	ActivityMutex sync.RWMutex
}

// Room 定義房間結構
type Room struct {
	ID         string                           `json:"id"`           // channel_id or dm_room_id
	Type       models.RoomType                  `json:"type"`         // channel, dm
	IDTypeHash string                           `json:"id_type_hash"` // 複合ID
	Clients    map[*Client]bool                 // 房間中的客戶端
	Broadcast  chan *WsMessage[MessageResponse] // 房間廣播通道
	Mutex      sync.RWMutex                     // 保護 Clients
}

// ChatService 管理所有的聊天功能
type ChatService struct {
	config          *config.Config
	mongoConnect    *mongo.Database
	redisClient     *redis.Client
	chatRepo        repositories.ChatRepositoryInterface
	serverRepo      repositories.ServerRepositoryInterface
	userRepo        repositories.UserRepositoryInterface
	Clients         map[*Client]bool   // 所有連接的客戶端
	ClientsByUserID map[string]*Client // 用戶ID到客戶端的映射
	Rooms           map[string]*Room   // 總房間映射
	RoomsMutex      sync.RWMutex       // 保護 Rooms
	// GuildPubSubs    map[string]*redis.PubSub       // 追蹤 Guild PubSub
	RoomPubSubs map[string]*redis.PubSub // 追蹤 Room PubSub
	// UserPubSubs     map[string]*redis.PubSub       // 追蹤使用者通知 PubSub
	PubSubMutex sync.RWMutex
	Register    chan *Client // 註冊通道
	Unregister  chan *Client // 註銷通道
	Mutex       sync.RWMutex // 保護共享資源的鎖
}

// NewChatService 初始化聊天室服務
func NewChatService(cfg *config.Config, mongodb *mongo.Database, chatRepo repositories.ChatRepositoryInterface, serverRepo repositories.ServerRepositoryInterface, userRepo repositories.UserRepositoryInterface) *ChatService {
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
		utils.PrettyPrint("Failed to ping Redis:", err)
		return nil
	}

	cs := &ChatService{
		config:          cfg,
		mongoConnect:    mongodb,
		redisClient:     redisClient,
		chatRepo:        chatRepo,
		serverRepo:      serverRepo,
		userRepo:        userRepo,
		Clients:         make(map[*Client]bool, 1000),
		ClientsByUserID: make(map[string]*Client, 1000),
		Rooms:           make(map[string]*Room, 1000),
		Register:        make(chan *Client, 1000), // 增加緩衝，為了避免使用者重連ws時阻塞
		Unregister:      make(chan *Client, 1000), // 增加緩衝，為了避免使用者重連ws時阻塞
	}

	// go cs.startIdleTimeoutChecker()
	go cs.handleRegister()
	go cs.handleUnregister()

	// // 確保聊天服務只啟動一次
	// runOnce.Do(func() {
	// 	go func() {
	// 		log.Printf("===== ChatService run goroutine 開始啟動 =====")
	// 		cs.run()
	// 		log.Printf("===== ChatService run goroutine 已結束 =====") // 這行正常情況下不應該被執行到
	// 	}()
	// })

	return cs
}

// handleRegister 處理客戶端註冊
func (cs *ChatService) handleRegister() {
	for client := range cs.Register {
		log.Printf("Handling Register event for user %s", client.UserID)
		cs.registerClient(client)
		cs.updateClientStatus(client, "online")
		// cs.subscribeToUserServers(client)
		// go cs.pushUnreadMessages(client)
		log.Printf("Register event completed for user %s", client.UserID)
	}
}

// handleUnregister 處理客戶端註銷
func (cs *ChatService) handleUnregister() {
	for client := range cs.Unregister {
		log.Printf("Handling Unregister event for user %s", client.UserID)
		// cs.unsubscribeFromUserServers(client)
		cs.unregisterClient(client)
		cs.updateClientStatus(client, "offline")
		log.Printf("Unregister event completed for user %s", client.UserID)
	}
}

// registerClient 註冊客戶端
func (cs *ChatService) registerClient(client *Client) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	client.Subscribed = make(map[string]bool)
	client.RoomActivity = make(map[string]time.Time)
	cs.Clients[client] = true
	cs.ClientsByUserID[client.UserID] = client
}

// unregisterClient 註銷客戶端
func (cs *ChatService) unregisterClient(client *Client) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()

	delete(cs.Clients, client)
	delete(cs.ClientsByUserID, client.UserID)

	for roomID, room := range cs.Rooms {
		room.Mutex.Lock()
		delete(room.Clients, client)
		room.Mutex.Unlock()
		cs.redisClient.SRem(context.Background(), "room:"+roomID+":members", client.UserID)
		if len(room.Clients) == 0 {
			cs.cleanupRoom(cs.RoomPubSubs[roomID], roomID)
		}
	}
	cs.redisClient.Del(context.Background(), "user:"+client.UserID+":rooms")
	client.Conn.Close()
}

// updateClientStatus 更新客戶端狀態
func (cs *ChatService) updateClientStatus(client *Client, status string) {
	ctx := context.Background()
	cs.redisClient.Set(ctx, "user:"+client.UserID+":status", status, 24*time.Hour)
	log.Printf("Update status for user %s: %s", client.UserID, status)
}

// checkUserAllowedJoinRoom 檢查房間是否允許使用者進入
func (cs *ChatService) checkUserAllowedJoinRoom(userID string, roomID string, roomType models.RoomType) (bool, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return false, err
	}

	// 檢查房間數量限制
	// rooms, err := cs.redisClient.SMembers(context.Background(), "user_id:"+userID+":rooms").Result()
	// if err == nil && len(rooms) >= 50 {
	// 	ws.WriteJSON(map[string]string{"error": "Room limit reached"})
	// 	continue
	// }

	if roomType == models.RoomTypeDM {
		var dmRoom models.DMRoom
		err := cs.mongoConnect.Collection("dm_rooms").FindOne(context.Background(), bson.M{"room_id": roomObjectID, "user_id": userObjectID}).Decode(&dmRoom)
		if err != nil {
			return false, err
		}

		return true, nil
	} else if roomType == models.RoomTypeChannel {
		var channel models.Channel
		err := cs.mongoConnect.Collection("channels").FindOne(context.Background(), bson.M{"_id": roomObjectID}).Decode(&channel)
		if err != nil {
			return false, err
		}

		var server models.Server
		err = cs.mongoConnect.Collection("servers").FindOne(context.Background(), bson.M{"_id": channel.GetID()}).Decode(&server)
		if err != nil {
			return false, err
		}

		var allowedUsers []string
		for _, member := range server.Members {
			allowedUsers = append(allowedUsers, member.UserID.Hex())
		}

		if utils.ContainsString(allowedUsers, userID) {
			return true, nil
		}

		return false, nil
	}

	return false, nil
}

// safelyBroadcastToClient 安全發送消息
func (cs *ChatService) safelyBroadcastToClient(client *Client, message *WsMessage[MessageResponse]) {
	select {
	case <-time.After(5 * time.Second):
		log.Printf("Write timeout for user %s in room %s", client.UserID, message.Data.RoomID)
	default:
		err := client.Conn.WriteJSON(message)
		if err != nil {
			log.Printf("Failed to send to user %s: %v", client.UserID, err)
			cs.Unregister <- client
		} else {
			// 更新房間活躍時間
			client.ActivityMutex.Lock()
			client.RoomActivity[message.Data.RoomID] = time.Now()
			client.ActivityMutex.Unlock()
			cs.redisClient.Set(context.Background(), "user:"+client.UserID+":room:"+message.Data.RoomID+":last_active", time.Now().Format(time.RFC3339), 24*time.Hour)
		}
	}
}

// broadcastWorker 房間的廣播工作池
func (cs *ChatService) broadcastWorker(room *Room) {
	for msg := range room.Broadcast {
		room.Mutex.RLock()
		for client := range room.Clients {
			go cs.safelyBroadcastToClient(client, msg)
		}
		room.Mutex.RUnlock()
	}
}

// initRoom 動態初始化房間
func (cs *ChatService) initRoom(roomID string, roomType models.RoomType) *Room {
	roomIDTypeHash := generateRoomID(roomType, roomID)

	// 先檢查房間是否存在（使用讀鎖）
	cs.RoomsMutex.RLock()
	room, exists := cs.Rooms[roomIDTypeHash]
	cs.RoomsMutex.RUnlock()

	if exists {
		log.Printf("Room %s already exists", roomIDTypeHash)
		return room
	}

	// 如果不存在，再創建新房間（使用寫鎖）
	cs.RoomsMutex.Lock()
	defer cs.RoomsMutex.Unlock()

	// 再次檢查（可能在獲取寫鎖的過程中，其他 goroutine 已經創建了房間）
	if room, exists := cs.Rooms[roomIDTypeHash]; exists {
		log.Printf("Room %s was created by another goroutine", roomIDTypeHash)
		return room
	}

	room = &Room{
		ID:         roomID,
		Type:       roomType,
		IDTypeHash: roomIDTypeHash,
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan *WsMessage[MessageResponse], 1000),
	}
	cs.Rooms[roomIDTypeHash] = room

	workerCount := 3
	if roomType == models.RoomTypeChannel {
		workerCount = 5
	}
	for i := 0; i < workerCount; i++ {
		go cs.broadcastWorker(room)
	}

	go func() {
		pubsub := cs.redisClient.Subscribe(context.Background(), "room:"+roomIDTypeHash)
		defer pubsub.Close()
		for msg := range pubsub.Channel() {
			utils.PrettyPrint("Redis Received message:", msg)
			var message *WsMessage[MessageResponse]
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			room.Mutex.RLock()
			for client := range room.Clients {
				go cs.safelyBroadcastToClient(client, message)
			}
			room.Mutex.RUnlock()
		}
	}()

	return room
}

// cleanupRoom 清理空房間
func (cs *ChatService) cleanupRoom(pubsub *redis.PubSub, roomIDTypeHash string) {
	cs.Mutex.Lock()
	defer cs.Mutex.Unlock()
	cs.PubSubMutex.Lock()
	defer cs.PubSubMutex.Unlock()

	if room, exists := cs.Rooms[roomIDTypeHash]; exists && len(room.Clients) == 0 {
		close(room.Broadcast)
		pubsub.Unsubscribe(context.Background(), "room:"+roomIDTypeHash)
		cs.RoomsMutex.Lock()
		delete(cs.Rooms, roomIDTypeHash)
		cs.RoomsMutex.Unlock()
		delete(cs.RoomPubSubs, roomIDTypeHash)
		log.Printf("Room %s cleaned up", roomIDTypeHash)
	}
}

func (cs *ChatService) joinRoom(client *Client, roomID string, roomType models.RoomType) {
	roomIDTypeHash := generateRoomID(roomType, roomID)
	cs.RoomsMutex.RLock()
	room, exists := cs.Rooms[roomIDTypeHash]
	cs.RoomsMutex.RUnlock()
	if !exists {
		log.Printf("Room %s not found", roomIDTypeHash)
		return
	}

	utils.PrettyPrint("Joining room:", room)

	room.Mutex.Lock()
	room.Clients[client] = true
	room.Mutex.Unlock()

	ctx := context.Background()
	cs.redisClient.SAdd(ctx, "room:"+roomIDTypeHash+":members", client.UserID)
	cs.redisClient.SAdd(ctx, "user_id:"+client.UserID+":rooms", roomIDTypeHash)

	client.ActivityMutex.Lock()
	client.RoomActivity[roomIDTypeHash] = time.Now()
	client.ActivityMutex.Unlock()
	cs.redisClient.Set(ctx, "user_id:"+client.UserID+":room:"+roomIDTypeHash+":last_active", time.Now().Format(time.RFC3339), 24*time.Hour)
}

// leaveRoom 讓使用者離開房間
func (cs *ChatService) leaveRoom(client *Client, roomID string, roomType models.RoomType) {
	roomIDTypeHash := generateRoomID(roomType, roomID)
	cs.RoomsMutex.RLock()
	room, exists := cs.Rooms[roomIDTypeHash]
	cs.RoomsMutex.RUnlock()
	if !exists {
		log.Printf("Room %s not found", roomIDTypeHash)
		return
	}

	room.Mutex.Lock()
	delete(room.Clients, client)
	room.Mutex.Unlock()

	ctx := context.Background()
	cs.redisClient.SRem(ctx, "room:"+roomIDTypeHash+":members", client.UserID)
	cs.redisClient.SRem(ctx, "user:"+client.UserID+":rooms", roomIDTypeHash)

	client.ActivityMutex.Lock()
	delete(client.RoomActivity, roomIDTypeHash)
	client.ActivityMutex.Unlock()
	cs.redisClient.Del(ctx, "user:"+client.UserID+":room:"+roomIDTypeHash+":last_active")

	if len(room.Clients) == 0 {
		cs.cleanupRoom(cs.RoomPubSubs[roomIDTypeHash], roomIDTypeHash)
	}

	// cs.updateLastReadTime(client.UserID, roomID)
	log.Printf("User %s left room %s", client.UserID, roomIDTypeHash)
}

func (cs *ChatService) updateLastReadTime(userID string, roomID string) {
	ctx := context.Background()
	cs.redisClient.Set(ctx, "user_id:"+userID+":room:"+roomID+":last_read", time.Now().Format(time.RFC3339), 24*time.Hour)
}

// HandleWebSocket 處理 WebSocket 連線
func (cs *ChatService) HandleWebSocket(ws *websocket.Conn, userID string) {
	// 初始化用戶
	client := &Client{
		UserID:        userID,
		Conn:          ws,
		Subscribed:    make(map[string]bool),
		RoomActivity:  make(map[string]time.Time),
		ActivityMutex: sync.RWMutex{},
	}
	cs.Register <- client

	for {
		var msg WsMessage[json.RawMessage]
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Read message failed: %v", err)
			cs.Unregister <- client
			break
		}

		switch msg.Action {
		case "join_room":
			// 解析 join_room 請求數據
			var data struct {
				RoomID   string          `json:"room_id"`
				RoomType models.RoomType `json:"room_type"`
			}
			err := json.Unmarshal(msg.Data, &data)
			if err != nil {
				log.Printf("Failed to parse join_room data: %v", err)
				continue
			}

			roomType := data.RoomType
			roomID := data.RoomID

			type joinRoomResponse struct {
				Status  string `json:"status"`
				Message string `json:"message"`
			}

			allowed, err := cs.checkUserAllowedJoinRoom(userID, roomID, roomType)
			if err != nil {
				message := &WsMessage[joinRoomResponse]{
					Action: "join_room",
					Data: joinRoomResponse{
						Status:  "error",
						Message: "Failed to check user allowed join room",
					},
				}
				ws.WriteJSON(message)
				continue
			}
			if !allowed {
				message := &WsMessage[joinRoomResponse]{
					Action: "join_room",
					Data: joinRoomResponse{
						Status:  "error",
						Message: "User not allowed to join room",
					},
				}
				ws.WriteJSON(message)
				continue
			}

			// 自動離開同一伺服器下的其他房間（僅伺服器頻道）
			// if roomType == RoomTypeChannel && channelID != "" {
			// 	currentRooms, err := cs.redisClient.SMembers(context.Background(), "user_id:"+userIDStr+":rooms").Result()
			// 	if err == nil {
			// 		for _, oldRoomID := range currentRooms {
			// 			cs.Mutex.RLock()
			// 			if oldRoom, exists := cs.Rooms[oldRoomID]; exists && oldRoom.ChannelID == channelID && oldRoomID != msg.RoomID {
			// 				cs.Mutex.RUnlock()
			// 				cs.leaveRoom(client, oldRoomID)
			// 			} else {
			// 				cs.Mutex.RUnlock()
			// 			}
			// 		}
			// 	}
			// }

			// 加入房間
			cs.joinRoom(client, roomID, roomType)

			// cs.updateLastReadTime(userID, msg.RoomID)

			message := &WsMessage[joinRoomResponse]{
				Action: "join_room",
				Data: joinRoomResponse{
					Status:  "success",
					Message: "Joined " + string(roomType) + " room " + roomID,
				},
			}

			ws.WriteJSON(message)
		case "leave_room":
			// 解析 leave_room 請求數據
			var data struct {
				RoomID   string          `json:"room_id"`
				RoomType models.RoomType `json:"room_type"`
			}
			err := json.Unmarshal(msg.Data, &data)
			if err != nil {
				log.Printf("Failed to parse leave_room data: %v", err)
				continue
			}

			roomID := data.RoomID
			roomType := data.RoomType

			type leaveRoomResponse struct {
				Action  string `json:"action"`
				Status  string `json:"status"`
				Message string `json:"message"`
			}

			cs.leaveRoom(client, roomID, roomType)
			message := &WsMessage[leaveRoomResponse]{
				Action: "leave_room",
				Data: leaveRoomResponse{
					Status:  "success",
					Message: "Left " + string(roomType) + " room " + roomID,
				},
			}
			ws.WriteJSON(message)
		case "send_message":
			// 解析 send_message 請求數據
			var data struct {
				RoomID   string          `json:"room_id"`
				RoomType models.RoomType `json:"room_type"`
				Content  string          `json:"content"`
			}
			err := json.Unmarshal(msg.Data, &data)
			if err != nil {
				log.Printf("Failed to parse send_message data: %v", err)
				continue
			}

			// 私聊則判斷對方房間是否已建立
			if data.RoomType == models.RoomTypeDM {
				RoomIDTypeHash := generateRoomID(data.RoomType, data.RoomID)

				// 檢查房間是否存在
				cs.RoomsMutex.RLock()
				utils.PrettyPrint("Checking room:", cs.Rooms)
				room, exists := cs.Rooms[RoomIDTypeHash]
				cs.RoomsMutex.RUnlock()

				if !exists {
					log.Printf("Room %s not found", RoomIDTypeHash)
					continue
				}

				// 檢查房間是否只有一個客戶端
				room.Mutex.RLock()
				isOnlyOneClient := len(room.Clients) == 1
				room.Mutex.RUnlock()

				// 如果是私聊且只有一個客戶端，則判斷是否建立房間
				if isOnlyOneClient {
					roomObjectID, err := primitive.ObjectIDFromHex(data.RoomID)
					if err != nil {
						log.Printf("Failed to parse room_id: %v", err)
						continue
					}

					userObjectID, err := primitive.ObjectIDFromHex(userID)
					if err != nil {
						log.Printf("Failed to parse user_id: %v", err)
						continue
					}

					// 先用RoomID找到ChatWithUserID
					var dmRoomList []models.DMRoom
					dbRoomCollection := cs.mongoConnect.Collection("dm_rooms")
					cursor, err := dbRoomCollection.Find(context.Background(), bson.M{
						"room_id": roomObjectID,
						"$or": []bson.M{
							{"user_id": userObjectID},
							{"chat_with_user_id": userObjectID},
						},
					})
					if err != nil {
						log.Printf("Failed to find dm room: %v", err)
						continue
					}

					cursor.All(context.Background(), &dmRoomList)

					// 如果對方房間不存在，則建立
					if len(dmRoomList) == 1 {
						for _, dmRoom := range dmRoomList {
							RoomID := dmRoom.RoomID
							dbRoomCollection.InsertOne(context.Background(), models.DMRoom{
								RoomID:         RoomID,
								UserID:         dmRoom.ChatWithUserID,
								ChatWithUserID: dmRoom.UserID,
								IsHidden:       false,
								CreatedAt:      time.Now(),
								UpdatedAt:      time.Now(),
							})
						}
					}
				}
			}

			message := &WsMessage[MessageResponse]{
				Action: "send_message",
				Data: MessageResponse{
					RoomID:    data.RoomID,
					RoomType:  data.RoomType,
					SenderID:  userID,
					Content:   data.Content,
					Timestamp: time.Now().UnixMilli(),
				},
			}
			cs.HandleMessage(message)
		}
	}
}

// HandleMessage 處理新消息並發布
func (cs *ChatService) HandleMessage(msg *WsMessage[MessageResponse]) {
	cs.Mutex.RLock()
	roomIDTypeHash := generateRoomID(msg.Data.RoomType, msg.Data.RoomID)
	room, exists := cs.Rooms[roomIDTypeHash]
	cs.Mutex.RUnlock()
	if !exists {
		log.Printf("%s Room %s not found", msg.Data.RoomType, msg.Data.RoomID)
		return
	}

	room.Mutex.RLock()
	_, isMember := room.Clients[cs.ClientsByUserID[msg.Data.SenderID]]
	room.Mutex.RUnlock()
	if !isMember {
		log.Printf("User ID %s not in room %s", msg.Data.SenderID, msg.Data.RoomID)
		return
	}

	ctx := context.Background()

	if msg.Data.RoomType == models.RoomTypeDM || msg.Data.RoomType == models.RoomTypeChannel {
		notifBytes, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Failed to encode notification: %v", err)
			return
		}

		err = cs.redisClient.Publish(ctx, "room:"+roomIDTypeHash, notifBytes).Err()
		if err != nil {
			log.Printf("Failed to publish channel notification: %v", err)
		}

		go cs.saveMessage(msg)
	}
}

// saveMessage 儲存消息到 MongoDB
func (cs *ChatService) saveMessage(msg *WsMessage[MessageResponse]) {
	collection := cs.mongoConnect.Collection("messages")
	roomID, err := primitive.ObjectIDFromHex(msg.Data.RoomID)
	if err != nil {
		log.Printf("Failed to convert room ID to ObjectID: %v", err)
		return
	}

	senderID, err := primitive.ObjectIDFromHex(msg.Data.SenderID)
	if err != nil {
		log.Printf("Failed to convert sender ID to ObjectID: %v", err)
		return
	}

	_, err = collection.InsertOne(context.Background(), models.Message{
		RoomID:    roomID,
		RoomType:  msg.Data.RoomType,
		SenderID:  senderID,
		Content:   msg.Data.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Printf("Failed to save message: %v", err)
	}
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

// 生成房間ID
func generateRoomID(roomType models.RoomType, channelID string) string {
	return fmt.Sprintf("%s:%s", roomType, channelID)
}
