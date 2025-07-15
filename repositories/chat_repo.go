package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
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
	odm          *providers.ODM
}

// NewChatRepository 創建一個新的聊天存儲庫實例
func NewChatRepository(cfg *config.Config, mongodb *mongo.Database) *ChatRepository {
	return &ChatRepository{
		config:       cfg,
		mongoConnect: mongodb,
		odm:          providers.NewODM(mongodb),
	}
}

// SaveMessage 將聊天消息保存到數據庫
func (cr *ChatRepository) SaveMessage(message models.Message) (string, error) {
	collection := cr.mongoConnect.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, message)
	if err != nil {
		log.Printf("保存聊天消息失敗: %v", err)
		return "", err
	}

	// 將插入結果的ID轉換為ObjectID
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Printf("無法獲取插入的消息ID")
		return "", err
	}

	return id.Hex(), nil
}

// GetMessagesByRoomID 根據房間ID獲取消息
func (cr *ChatRepository) GetMessagesByRoomID(roomID string, limit int64) ([]models.Message, error) {
	collection := cr.mongoConnect.Collection("messages")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		log.Printf("轉換房間ID失敗: %v", err)
		return nil, err
	}

	filter := map[string]interface{}{
		"room_id": roomObjectID,
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

// GetDMRoomListByUserID 獲取用戶的聊天列表
func (cr *ChatRepository) GetDMRoomListByUserID(userID string, includeHidden bool) ([]models.DMRoom, error) {
	collection := cr.mongoConnect.Collection("dm_rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("轉換用戶ID失敗: %v", err)
		return nil, err
	}

	filter := map[string]interface{}{
		"user_id": userObjectID,
	}

	// includeHidden 為 false 時，過濾已隱藏的聊天列表
	if !includeHidden {
		filter["is_hidden"] = false
	}

	findOptions := options.Find()
	findOptions.SetSort(map[string]interface{}{"updated_at": -1}) // 按最後聊天時間倒序

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Printf("查詢聊天列表失敗: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var chatLists []models.DMRoom
	if err = cursor.All(ctx, &chatLists); err != nil {
		log.Printf("解析聊天列表數據失敗: %v", err)
		return nil, err
	}

	return chatLists, nil
}

// UpdateDMRoom 更新聊天列表的刪除狀態
func (cr *ChatRepository) UpdateDMRoom(userID string, chatWithUserID string, IsHidden bool) error {
	collection := cr.mongoConnect.Collection("dm_rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("轉換用戶ID失敗: %v", err)
		return err
	}
	chatWithUserObjectID, err := primitive.ObjectIDFromHex(chatWithUserID)
	if err != nil {
		log.Printf("轉換聊天對象ID失敗: %v", err)
		return err
	}

	filter := map[string]interface{}{
		"user_id":           userObjectID,
		"chat_with_user_id": chatWithUserObjectID,
	}

	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"is_hidden":  IsHidden,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("更新聊天列表刪除狀態失敗: %v", err)
		return err
	}

	return nil
}

// SaveOrUpdateDMRoom 保存或更新聊天列表
func (cr *ChatRepository) SaveOrUpdateDMRoom(chat models.DMRoom) (models.DMRoom, error) {
	collection := cr.mongoConnect.Collection("dm_rooms")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":           chat.UserID,
		"chat_with_user_id": chat.ChatWithUserID,
	}

	// 檢查是否已存在該聊天列表
	var existingChatList models.DMRoom
	err := collection.FindOne(ctx, filter).Decode(&existingChatList)
	date := time.Now()

	if err == nil {
		// 已存在，更新
		update := bson.M{
			"$set": bson.M{
				"is_hidden":  chat.IsHidden,
				"updated_at": date,
			},
		}

		chat.ID = existingChatList.ID
		chat.UpdatedAt = date

		_, err = collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Printf("更新聊天列表失敗: %v", err)
			return models.DMRoom{}, err
		}
	} else if err == mongo.ErrNoDocuments {
		// 不存在，創建新的
		chat.ID = primitive.NewObjectID()
		chat.CreatedAt = date
		chat.UpdatedAt = date
		chat.IsHidden = false // 初始化為未刪除狀態

		_, err = collection.InsertOne(ctx, chat)
		if err != nil {
			log.Printf("創建聊天列表失敗: %v", err)
			return models.DMRoom{}, err
		}
	} else {
		// 其他錯誤
		log.Printf("查詢聊天列表時發生錯誤: %v", err)
		return models.DMRoom{}, err
	}

	return chat, nil
}
