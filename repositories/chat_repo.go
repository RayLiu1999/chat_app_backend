package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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

// GetChatListByUserID 獲取用戶的聊天列表
func (cr *ChatRepository) GetChatListByUserID(userID primitive.ObjectID, includeDeleted bool) ([]models.Chat, error) {
	collection := cr.mongoConnect.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := map[string]interface{}{
		"user_id": userID,
	}

	// 如果不包含已刪除的記錄，則添加過濾條件
	if !includeDeleted {
		filter["is_deleted"] = false
	}

	findOptions := options.Find()
	findOptions.SetSort(map[string]interface{}{"updated_at": -1}) // 按最後聊天時間倒序

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("查詢聊天列表失敗: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var chatLists []models.Chat
	if err = cursor.All(ctx, &chatLists); err != nil {
		log.Printf("解析聊天列表數據失敗: %v", err)
		return nil, err
	}

	return chatLists, nil
}

// UpdateChatListDeleteStatus 更新聊天列表的刪除狀態
func (cr *ChatRepository) UpdateChatListDeleteStatus(userID, chatWithUserID primitive.ObjectID, isDeleted bool) error {
	collection := cr.mongoConnect.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := map[string]interface{}{
		"user_id":           userID,
		"chat_with_user_id": chatWithUserID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_deleted": isDeleted,
			"updated_at": time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("更新聊天列表刪除狀態失敗: %v", err)
		return err
	}

	return nil
}

// SaveOrUpdateChat 保存或更新聊天列表
func (cr *ChatRepository) SaveOrUpdateChat(chat models.Chat) (models.Chat, error) {
	collection := cr.mongoConnect.Collection("chats")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":           chat.UserID,
		"chat_with_user_id": chat.ChatWithUserID,
	}

	// 檢查是否已存在該聊天列表
	var existingChatList models.Chat
	err := collection.FindOne(ctx, filter).Decode(&existingChatList)
	date := time.Now()

	if err == nil {
		// 已存在，更新
		update := bson.M{
			"$set": bson.M{
				"is_deleted": chat.IsDeleted,
				"updated_at": date,
			},
		}

		chat.ID = existingChatList.ID
		chat.UpdatedAt = date

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("更新聊天列表失敗: %v", err)
			return models.Chat{}, err
		}
	} else if err == mongo.ErrNoDocuments {
		// 不存在，創建新的
		chat.ID = primitive.NewObjectID()
		chat.CreatedAt = date
		chat.UpdatedAt = date
		chat.IsDeleted = false // 初始化為未刪除狀態

		_, err = collection.InsertOne(ctx, chat)
		if err != nil {
			log.Printf("創建聊天列表失敗: %v", err)
			return models.Chat{}, err
		}
	} else {
		// 其他錯誤
		log.Printf("查詢聊天列表時發生錯誤: %v", err)
		return models.Chat{}, err
	}

	return chat, nil
}
