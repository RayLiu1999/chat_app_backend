package services

import (
	"chat_app_backend/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// MessageHandler 處理消息相關邏輯
type MessageHandler struct {
	mongoConnect *mongo.Database
	roomManager  *RoomManager
}

// NewMessageHandler 創建新的消息處理器
func NewMessageHandler(mongoConnect *mongo.Database, roomManager *RoomManager) *MessageHandler {
	return &MessageHandler{
		mongoConnect: mongoConnect,
		roomManager:  roomManager,
	}
}

// HandleMessage 處理消息邏輯
func (mh *MessageHandler) HandleMessage(message *WsMessage[MessageResponse]) {
	// 檢查房間是否存在
	room, exists := mh.roomManager.GetRoom(message.Data.RoomType, message.Data.RoomID)
	if !exists {
		log.Printf("Room %s not found", message.Data.RoomID)
		return
	}

	// 儲存消息到資料庫
	mh.saveMessageToDB(message.Data)

	// 發送消息給房間內的所有客戶端
	room.Mutex.RLock()
	defer room.Mutex.RUnlock()

	for client := range room.Clients {
		err := client.Conn.WriteJSON(message)
		if err != nil {
			log.Printf("Failed to send message to client %s: %v", client.UserID, err)
			mh.roomManager.LeaveRoom(client, message.Data.RoomType, message.Data.RoomID)
		}
	}
}

// saveMessageToDB 儲存消息到資料庫
func (mh *MessageHandler) saveMessageToDB(data MessageResponse) {
	roomObjectID, err := primitive.ObjectIDFromHex(data.RoomID)
	if err != nil {
		log.Printf("Failed to parse room_id: %v", err)
		return
	}

	senderObjectID, err := primitive.ObjectIDFromHex(data.SenderID)
	if err != nil {
		log.Printf("Failed to parse sender_id: %v", err)
		return
	}

	var collectionName string
	switch data.RoomType {
	case models.RoomTypeDM:
		collectionName = "dm_messages"
	case models.RoomTypeChannel:
		collectionName = "channel_messages"
	default:
		log.Printf("Unknown room type: %s", data.RoomType)
		return
	}

	message := models.Message{
		RoomID:    roomObjectID,
		SenderID:  senderObjectID,
		Content:   data.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := mh.mongoConnect.Collection(collectionName)
	_, err = collection.InsertOne(ctx, message)
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		return
	}

	// 更新房間的最後訊息時間
	mh.updateRoomLastMessage(data.RoomID, data.RoomType)
}

// updateRoomLastMessage 更新房間的最後訊息時間
func (mh *MessageHandler) updateRoomLastMessage(roomID string, roomType models.RoomType) {
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		log.Printf("Failed to parse room_id: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch roomType {
	case models.RoomTypeDM:
		collection := mh.mongoConnect.Collection("dm_rooms")
		_, err = collection.UpdateMany(ctx,
			bson.M{"room_id": roomObjectID},
			bson.M{"$set": bson.M{"updated_at": time.Now()}},
		)
		if err != nil {
			log.Printf("Failed to update dm room last message time: %v", err)
		}
	case models.RoomTypeChannel:
		collection := mh.mongoConnect.Collection("channels")
		_, err = collection.UpdateOne(ctx,
			bson.M{"_id": roomObjectID},
			bson.M{"$set": bson.M{"last_message_at": time.Now()}},
		)
		if err != nil {
			log.Printf("Failed to update channel last message time: %v", err)
		}
	}
}
