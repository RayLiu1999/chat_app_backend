package repositories

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/config"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ChatRepository 處理聊天相關的數據庫操作
type ChatRepository struct {
	config *config.Config
	odm    *providers.ODM
	// queryBuilder *providers.QueryBuilder // 如有需要可加
}

// NewChatRepository 創建一個新的聊天存儲庫實例
func NewChatRepository(cfg *config.Config, odm *providers.ODM) *ChatRepository {
	return &ChatRepository{
		config: cfg,
		odm:    odm,
		// queryBuilder: qb, // 如有需要
	}
}

// SaveMessage 將聊天消息保存到數據庫
func (cr *ChatRepository) SaveMessage(message models.Message) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := cr.odm.Create(ctx, &message)
	if err != nil {
		return "", err
	}

	return message.ID.Hex(), nil
}

// GetMessagesByRoomID 根據房間ID獲取消息
func (cr *ChatRepository) GetMessagesByRoomID(roomID string, limit int64) ([]models.Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return nil, err
	}

	filter := map[string]any{
		"room_id": roomObjectID,
	}

	queryOptions := providers.QueryOptions{
		Sort:  bson.D{{Key: "created_at", Value: -1}},
		Limit: &limit,
	}

	var messages []models.Message
	err = cr.odm.FindWithOptions(ctx, filter, &messages, &queryOptions)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// GetDMRoomListByUserID 獲取用戶的聊天列表
func (cr *ChatRepository) GetDMRoomListByUserID(userID string, includeHidden bool) ([]models.DMRoom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	filter := map[string]any{
		"user_id": userObjectID,
	}

	// includeHidden 為 false 時，過濾已隱藏的聊天列表
	if !includeHidden {
		filter["is_hidden"] = false
	}

	queryOptions := providers.QueryOptions{
		Sort: bson.D{{Key: "updated_at", Value: -1}},
	}

	var chatLists []models.DMRoom
	err = cr.odm.FindWithOptions(ctx, filter, &chatLists, &queryOptions)
	if err != nil {
		return nil, err
	}

	return chatLists, nil
}

// UpdateDMRoom 更新聊天列表的刪除狀態
func (cr *ChatRepository) UpdateDMRoom(userID string, chatWithUserID string, IsHidden bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	chatWithUserObjectID, err := primitive.ObjectIDFromHex(chatWithUserID)
	if err != nil {
		return err
	}

	filter := map[string]any{
		"user_id":           userObjectID,
		"chat_with_user_id": chatWithUserObjectID,
	}

	update := map[string]any{
		"$set": map[string]any{
			"is_hidden":  IsHidden,
			"updated_at": time.Now(),
		},
	}

	err = cr.odm.UpdateMany(ctx, &models.DMRoom{}, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// SaveOrUpdateDMRoom 保存或更新聊天列表
func (cr *ChatRepository) SaveOrUpdateDMRoom(chat models.DMRoom) (models.DMRoom, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":           chat.UserID,
		"chat_with_user_id": chat.ChatWithUserID,
	}

	// 檢查是否已存在該聊天列表
	var existingChatList models.DMRoom
	err := cr.odm.FindOne(ctx, filter, &existingChatList)
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

		err = cr.odm.UpdateMany(ctx, &models.DMRoom{}, filter, update)
		if err != nil {
			return models.DMRoom{}, err
		}
	} else if err == mongo.ErrNoDocuments {
		// 不存在，創建新的
		chat.ID = primitive.NewObjectID()
		chat.CreatedAt = date
		chat.UpdatedAt = date
		chat.IsHidden = false // 初始化為未刪除狀態

		err = cr.odm.Create(ctx, &chat)
		if err != nil {
			return models.DMRoom{}, err
		}
	} else {
		// 其他錯誤
		return models.DMRoom{}, err
	}

	return chat, nil
}

// DeleteMessagesByRoomID 根據房間ID刪除所有訊息
func (cr *ChatRepository) DeleteMessagesByRoomID(roomID string) error {
	ctx := context.Background()

	// 將 roomID 轉換為 ObjectID
	roomObjectID, err := primitive.ObjectIDFromHex(roomID)
	if err != nil {
		return err
	}

	// 刪除該房間的所有訊息
	filter := bson.M{"room_id": roomObjectID}
	_, err = cr.odm.GetDatabase().Collection("messages").DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}
