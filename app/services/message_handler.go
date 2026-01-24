package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/utils"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// messageHandler 處理消息相關邏輯
type messageHandler struct {
	odm         providers.ODM
	roomManager RoomManager
	redisClient *redis.Client
}

// NewMessageHandler 創建新的消息處理器
func NewMessageHandler(odm providers.ODM, roomManager RoomManager, redisClient *redis.Client) *messageHandler {
	return &messageHandler{
		odm:         odm,
		roomManager: roomManager,
		redisClient: redisClient,
	}
}

// HandleMessage 處理消息邏輯
// 透過 Redis Pub/Sub 實現跨實例廣播
func (mh *messageHandler) HandleMessage(message *MessageResponse) {
	// 非同步儲存消息到資料庫
	go func() {
		err := mh.saveMessageToDB(message)
		if err != nil {
			utils.Log.Error("儲存消息到資料庫失敗", "error", err)
		}
	}()

	// 構建要發送的訊息結構
	wsMsg := &WsMessage[*MessageResponse]{
		Action: "new_message", // 所有實例收到後會根據 senderID 決定是 message_sent 還是 new_message
		Data:   message,
	}

	// 序列化訊息
	msgJSON, err := json.Marshal(wsMsg)
	if err != nil {
		utils.Log.Error("序列化訊息失敗", "error", err)
		return
	}

	// 構建 room key
	roomKey := RoomKey{Type: message.RoomType, RoomID: message.RoomID}
	channel := "room:" + roomKey.String()

	// 透過 Redis Publish 發送訊息給所有訂閱的實例
	ctx := context.Background()
	if err := mh.redisClient.Publish(ctx, channel, msgJSON).Err(); err != nil {
		utils.Log.Error("Redis Publish 失敗", "error", err)
		// 如果 Redis 失敗，回退到本地廣播
		mh.localBroadcast(message)
		return
	}
}

// localBroadcast 本地廣播（Redis 失敗時的回退方案）
func (mh *messageHandler) localBroadcast(message *MessageResponse) {
	room, exists := mh.roomManager.GetRoom(message.RoomType, message.RoomID)
	if !exists {
		return
	}

	room.Mutex.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for client := range room.Clients {
		clients = append(clients, client)
	}
	room.Mutex.RUnlock()

	for _, client := range clients {
		go func(c *Client) {
			if !mh.isClientConnectionValid(c) {
				go mh.roomManager.LeaveRoom(c, message.RoomType, message.RoomID)
				return
			}

			var action string
			if c.UserID == message.SenderID {
				action = "message_sent"
			} else {
				action = "new_message"
			}

			outMsg := &WsMessage[*MessageResponse]{
				Action: action,
				Data:   message,
			}

			if err := c.SendMessage(outMsg); err != nil {
				go mh.roomManager.LeaveRoom(c, message.RoomType, message.RoomID)
			}
		}(client)
	}
}

// saveMessageToDB 儲存消息到資料庫
func (mh *messageHandler) saveMessageToDB(data *MessageResponse) error {
	roomObjectID, err := primitive.ObjectIDFromHex(data.RoomID)
	if err != nil {
		return err
	}

	senderObjectID, err := primitive.ObjectIDFromHex(data.SenderID)
	if err != nil {
		return err
	}

	message := &models.Message{
		RoomID:   roomObjectID,
		SenderID: senderObjectID,
		Content:  data.Content,
		RoomType: data.RoomType,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = mh.odm.Create(ctx, message)
	if err != nil {
		return err
	}

	mh.updateRoomLastMessage(data.RoomID, data.RoomType)
	return nil
}

// updateRoomLastMessage 更新房間的最後訊息時間
func (mh *messageHandler) updateRoomLastMessage(roomID string, roomType models.RoomType) {
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch roomType {
	case models.RoomTypeDM:
		// 更新 dm_rooms 的 updated_at
		dmRoom := &models.DMRoom{}
		mh.odm.UpdateMany(ctx, dmRoom, map[string]any{"room_id": roomObjectID}, map[string]any{"$set": map[string]any{"updated_at": time.Now()}})
	case models.RoomTypeChannel:
		// 更新 channels 的 last_message_at
		now := time.Now()
		mh.odm.UpdateMany(ctx, &models.Channel{},
			map[string]any{"_id": roomObjectID},
			map[string]any{"$set": map[string]any{"last_message_at": now}})
	}
}

// isClientConnectionValid 檢查客戶端連線是否仍然有效
func (mh *messageHandler) isClientConnectionValid(client *Client) bool {
	// 檢查連線是否為 nil
	if client == nil || client.Conn == nil {
		return false
	}

	// 檢查連線是否已標記為非活躍
	if !client.IsActive {
		return false
	}

	// 檢查最後 pong 時間是否過久（超過 2 分鐘）
	if time.Since(client.LastPongTime) > 2*time.Minute {
		client.IsActive = false
		return false
	}

	// 檢查連線時間是否過久（超過 24 小時）
	if time.Since(client.ConnectedAt) > 24*time.Hour {
		client.IsActive = false
		return false
	}

	return true
}
