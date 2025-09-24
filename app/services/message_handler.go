package services

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/utils"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MessageHandler 處理消息相關邏輯
type MessageHandler struct {
	odm         *providers.ODM
	roomManager *RoomManager
}

// NewMessageHandler 創建新的消息處理器
func NewMessageHandler(odm *providers.ODM, roomManager *RoomManager) *MessageHandler {
	return &MessageHandler{
		odm:         odm,
		roomManager: roomManager,
	}
}

// HandleMessage 處理消息邏輯
func (mh *MessageHandler) HandleMessage(message *MessageResponse) {
	// 先儲存消息到資料庫（不管房間是否存在客戶端）
	err := mh.saveMessageToDB(message)
	if err != nil {
		utils.PrettyPrintf("儲存消息到資料庫失敗: %v", err)
		return
	}

	// 檢查房間是否存在
	room, exists := mh.roomManager.GetRoom(message.RoomType, message.RoomID)
	if !exists {
		utils.PrettyPrintf("房間 %s 不存在，消息已儲存到資料庫，但沒有客戶端可廣播", message.RoomID)
		return
	}

	// 發送消息給房間內的所有客戶端
	room.Mutex.RLock()
	clients := make([]*Client, 0, len(room.Clients))
	for client := range room.Clients {
		clients = append(clients, client)
	}
	room.Mutex.RUnlock()

	// 在外部發送消息，避免長時間持有鎖
	for _, client := range clients {
		go func(c *Client) {
			// 檢查連線是否仍然有效
			if !mh.isClientConnectionValid(c) {
				utils.PrettyPrintf("客戶端 %s 連線已失效，從房間中移除", c.UserID)
				go mh.roomManager.LeaveRoom(c, message.RoomType, message.RoomID)
				return
			}

			// 根據用戶是否為發送者決定 action
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
				utils.PrettyPrintf("發送消息失敗: %v", err)
				// 異步移除有問題的客戶端
				go mh.roomManager.LeaveRoom(c, message.RoomType, message.RoomID)
			}
		}(client)
	}
}

// saveMessageToDB 儲存消息到資料庫
func (mh *MessageHandler) saveMessageToDB(data *MessageResponse) error {
	roomObjectID, err := primitive.ObjectIDFromHex(data.RoomID)
	if err != nil {
		utils.PrettyPrintf("解析房間ID失敗: %v", err)
		return err
	}

	senderObjectID, err := primitive.ObjectIDFromHex(data.SenderID)
	if err != nil {
		utils.PrettyPrintf("解析發送者ID失敗: %v", err)
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
		utils.PrettyPrintf("儲存消息失敗: %v", err)
		return err
	}

	utils.PrettyPrintf("消息已儲存到資料庫: 房間=%s, 發送者=%s, 內容=%s", data.RoomID, data.SenderID, data.Content)

	mh.updateRoomLastMessage(data.RoomID, data.RoomType)
	return nil
}

// updateRoomLastMessage 更新房間的最後訊息時間
func (mh *MessageHandler) updateRoomLastMessage(roomID string, roomType models.RoomType) {
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		utils.PrettyPrintf("解析房間ID失敗: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch roomType {
	case models.RoomTypeDM:
		// 更新 dm_rooms 的 updated_at
		dmRoom := &models.DMRoom{}
		err = mh.odm.UpdateMany(ctx, dmRoom, map[string]any{"room_id": roomObjectID}, map[string]any{"$set": map[string]any{"updated_at": time.Now()}})
		if err != nil {
			utils.PrettyPrintf("更新 dm 房間最後訊息時間失敗: %v", err)
		}
	case models.RoomTypeChannel:
		// 更新 channels 的 last_message_at
		now := time.Now()
		err = mh.odm.UpdateMany(ctx, &models.Channel{},
			map[string]any{"_id": roomObjectID},
			map[string]any{"$set": map[string]any{"last_message_at": now}})
		if err != nil {
			utils.PrettyPrintf("更新 channel 房間最後訊息時間失敗: %v", err)
		}
	}
}

// isClientConnectionValid 檢查客戶端連線是否仍然有效
func (mh *MessageHandler) isClientConnectionValid(client *Client) bool {
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
		utils.PrettyPrintf("客戶端 %s 最後 Pong 時間過久: %v", client.UserID, client.LastPongTime)
		client.IsActive = false
		return false
	}

	// 檢查連線時間是否過久（超過 24 小時）
	if time.Since(client.ConnectedAt) > 24*time.Hour {
		utils.PrettyPrintf("客戶端 %s 連線時間過久: %v", client.UserID, client.ConnectedAt)
		client.IsActive = false
		return false
	}

	return true
}
