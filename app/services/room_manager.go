package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/app/repositories"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RoomManager 管理房間的創建、加入、離開等操作
type RoomManager struct {
	odm              *providers.ODM
	rooms            map[string]*Room
	roomPubSubs      map[string]*redis.PubSub
	redisClient      *redis.Client
	serverMemberRepo repositories.ServerMemberRepositoryInterface
	mutex            sync.RWMutex
	pubSubMutex      sync.RWMutex
}

// NewRoomManager 創建新的房間管理器
func NewRoomManager(odm *providers.ODM, redisClient *redis.Client, serverMemberRepo repositories.ServerMemberRepositoryInterface) *RoomManager {
	return &RoomManager{
		odm:              odm,
		rooms:            make(map[string]*Room, 1000),
		roomPubSubs:      make(map[string]*redis.PubSub),
		redisClient:      redisClient,
		serverMemberRepo: serverMemberRepo,
	}
}

// GetRoom 獲取房間
func (rm *RoomManager) GetRoom(roomType models.RoomType, roomID string) (*Room, bool) {
	key := RoomKey{Type: roomType, RoomID: roomID}
	rm.mutex.RLock()
	room, exists := rm.rooms[key.String()]
	rm.mutex.RUnlock()
	return room, exists
}

// AddRoom 新增房間
func (rm *RoomManager) AddRoom(room *Room) {
	rm.mutex.Lock()
	rm.rooms[room.Key.String()] = room
	rm.mutex.Unlock()
}

// InitRoom 動態初始化房間
func (rm *RoomManager) InitRoom(roomType models.RoomType, roomID string) *Room {
	key := RoomKey{Type: roomType, RoomID: roomID}

	// 先檢查房間是否存在（使用讀鎖）
	rm.mutex.RLock()
	room, exists := rm.rooms[key.String()]
	rm.mutex.RUnlock()

	if exists {
		log.Printf("Room %s already exists", key.String())
		return room
	}

	room = &Room{
		Key:       key,
		ID:        roomID,
		Type:      roomType,
		Clients:   make(map[*Client]bool),
		Broadcast: make(chan *WsMessage[MessageResponse], 1000),
	}
	rm.AddRoom(room)

	// 根據房間類型設定工作池大小
	workerCount := 3
	if roomType == models.RoomTypeChannel {
		workerCount = 5
	}
	for i := 0; i < workerCount; i++ {
		go rm.broadcastWorker(room)
	}

	// 設置 Redis Pub/Sub
	go func() {
		pubsub := rm.redisClient.Subscribe(context.Background(), "room:"+key.String())

		rm.pubSubMutex.Lock()
		rm.roomPubSubs[key.String()] = pubsub
		rm.pubSubMutex.Unlock()

		defer func() {
			pubsub.Close()
			rm.pubSubMutex.Lock()
			delete(rm.roomPubSubs, key.String())
			rm.pubSubMutex.Unlock()
		}()

		for msg := range pubsub.Channel() {
			var message *WsMessage[MessageResponse]
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			room.Mutex.RLock()
			for client := range room.Clients {
				go rm.safelyBroadcastToClient(client, message)
			}
			room.Mutex.RUnlock()
		}
	}()

	return room
}

// JoinRoom 讓使用者加入房間
func (rm *RoomManager) JoinRoom(client *Client, roomType models.RoomType, roomID string) {
	key := RoomKey{Type: roomType, RoomID: roomID}
	rm.mutex.RLock()
	room, exists := rm.rooms[key.String()]
	rm.mutex.RUnlock()
	if !exists {
		log.Printf("Room %s not found", key.String())
		return
	}

	room.Mutex.Lock()
	room.Clients[client] = true
	room.Mutex.Unlock()

	ctx := context.Background()
	rm.redisClient.SAdd(ctx, "room:"+key.String()+":members", client.UserID)
	rm.redisClient.SAdd(ctx, "user_id:"+client.UserID+":rooms", key.String())

	client.ActivityMutex.Lock()
	client.RoomActivity[key.String()] = time.Now()
	client.ActivityMutex.Unlock()
	rm.redisClient.Set(ctx, "user_id:"+client.UserID+":room:"+key.String()+":last_active", time.Now().UnixMilli(), 24*time.Hour)
}

// LeaveRoom 讓使用者離開房間
func (rm *RoomManager) LeaveRoom(client *Client, roomType models.RoomType, roomID string) {
	key := RoomKey{Type: roomType, RoomID: roomID}
	rm.mutex.RLock()
	room, exists := rm.rooms[key.String()]
	rm.mutex.RUnlock()
	if !exists {
		log.Printf("Room %s not found", key.String())
		return
	}

	room.Mutex.Lock()
	delete(room.Clients, client)
	clientCount := len(room.Clients)
	room.Mutex.Unlock()

	ctx := context.Background()
	rm.redisClient.SRem(ctx, "room:"+key.String()+":members", client.UserID)
	rm.redisClient.SRem(ctx, "user:"+client.UserID+":rooms", key.String())

	client.ActivityMutex.Lock()
	delete(client.RoomActivity, key.String())
	client.ActivityMutex.Unlock()
	rm.redisClient.Del(ctx, "user:"+client.UserID+":room:"+key.String()+":last_active")

	if clientCount == 0 {
		rm.cleanupRoom(key.String())
	}

	log.Printf("User %s left room %s", client.UserID, key.String())
}

// CleanupRoom 清理空房間
func (rm *RoomManager) cleanupRoom(roomKey string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if room, exists := rm.rooms[roomKey]; exists && len(room.Clients) == 0 {
		close(room.Broadcast)

		rm.pubSubMutex.Lock()
		if pubsub, exists := rm.roomPubSubs[roomKey]; exists {
			pubsub.Unsubscribe(context.Background(), "room:"+roomKey)
			delete(rm.roomPubSubs, roomKey)
		}
		rm.pubSubMutex.Unlock()

		delete(rm.rooms, roomKey)
		log.Printf("Room %s cleaned up", roomKey)
	}
}

// checkUserAllowedJoinRoom 檢查房間是否允許使用者進入
func (rm *RoomManager) checkUserAllowedJoinRoom(userID string, roomID string, roomType models.RoomType) (bool, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return false, err
	}

	ctx := context.Background()

	if roomType == models.RoomTypeDM {
		dmRoom := &models.DMRoom{}
		err := rm.odm.FindOne(ctx, bson.M{"room_id": roomObjectID, "user_id": userObjectID}, dmRoom)
		if err != nil {
			if err == providers.ErrDocumentNotFound {
				return false, nil
			}
			return false, err
		}
		return true, nil
	} else if roomType == models.RoomTypeChannel {
		channel := &models.Channel{}
		err := rm.odm.FindOne(ctx, bson.M{"_id": roomObjectID}, channel)
		if err != nil {
			if err == providers.ErrDocumentNotFound {
				return false, nil
			}
			return false, err
		}

		server := &models.Server{}
		err = rm.odm.FindOne(ctx, bson.M{"_id": channel.ServerID}, server)
		if err != nil {
			if err == providers.ErrDocumentNotFound {
				return false, nil
			}
			return false, err
		}

		// 使用 ServerMemberRepository 檢查用戶是否為伺服器成員
		isMember, err := rm.serverMemberRepo.IsMemberOfServer(channel.ServerID.Hex(), userID)
		if err != nil {
			log.Printf("Error checking server membership: %v", err)
			return false, err
		}

		return isMember, nil
	}
	return false, nil
}

// broadcastWorker 房間的廣播工作池
func (rm *RoomManager) broadcastWorker(room *Room) {
	for msg := range room.Broadcast {
		room.Mutex.RLock()
		for client := range room.Clients {
			go rm.safelyBroadcastToClient(client, msg)
		}
		room.Mutex.RUnlock()
	}
}

// safelyBroadcastToClient 安全發送消息
func (rm *RoomManager) safelyBroadcastToClient(client *Client, message *WsMessage[MessageResponse]) {
	// 使用統一的發送機制，而不是直接寫入 WebSocket
	if err := client.SendMessage(message); err != nil {
		log.Printf("Failed to send to user %s: %v", client.UserID, err)
		// 標記客戶端為非活躍，讓健康檢查清理
		client.IsActive = false
		return
	}

	// 更新房間活躍時間
	client.ActivityMutex.Lock()
	client.RoomActivity[message.Data.RoomID] = time.Now()
	client.ActivityMutex.Unlock()
	rm.redisClient.Set(context.Background(), "user:"+client.UserID+":room:"+message.Data.RoomID+":last_active", time.Now().UnixMilli(), 24*time.Hour)
}
