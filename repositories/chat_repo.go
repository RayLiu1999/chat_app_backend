package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ChatRepository 處理聊天相關的數據庫操作
type ChatRepository struct {
	config       *config.Config
	mongoConnect *mongo.Database
}

// NewChatRepository 創建一個新的聊天存儲庫實例
func NewChatRepository(cfg *config.Config, mongodb *mongo.Database) *ChatRepository {
	return &ChatRepository{
		config:       cfg,
		mongoConnect: mongodb,
	}
}

// SaveMessage 將聊天消息保存到數據庫
func (cr *ChatRepository) SaveMessage(message models.Message) (primitive.ObjectID, error) {
	collection := cr.mongoConnect.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		log.Printf("保存聊天消息失敗: %v", err)
		return primitive.NilObjectID, err
	}

	// 將插入結果的ID轉換為ObjectID
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Printf("無法獲取插入的消息ID")
		return primitive.NilObjectID, err
	}

	return id, nil
}

// GetMessagesByRoomID 根據房間ID獲取消息
func (cr *ChatRepository) GetMessagesByRoomID(roomID primitive.ObjectID, limit int64) ([]models.Message, error) {
	collection := cr.mongoConnect.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := map[string]interface{}{
		"room_id": roomID,
	}

	findOptions := options.Find()
	findOptions.SetSort(map[string]interface{}{"created_at": -1}) // 按時間倒序
	findOptions.SetLimit(limit)

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("查詢房間消息失敗: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var messages []models.Message
	if err = cursor.All(ctx, &messages); err != nil {
		log.Printf("解析消息數據失敗: %v", err)
		return nil, err
	}

	return messages, nil
}
