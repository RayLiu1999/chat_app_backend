package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChannelRepository struct {
	config *config.Config
	odm    *providers.ODM
}

func NewChannelRepository(cfg *config.Config, odm *providers.ODM) ChannelRepositoryInterface {
	return &ChannelRepository{
		config: cfg,
		odm:    odm,
	}
}

// GetChannelsByServerID 根據伺服器ID獲取頻道列表
func (cr *ChannelRepository) GetChannelsByServerID(serverID string) ([]models.Channel, error) {
	serverObjID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return nil, err
	}

	qb := providers.NewQueryBuilder()
	qb.Where("server_id", serverObjID)

	var channels []models.Channel
	err = cr.odm.Find(context.Background(), qb.GetFilter(), &channels)
	if err != nil {
		return nil, err
	}

	return channels, nil
}

// GetChannelByID 根據頻道ID獲取頻道
func (cr *ChannelRepository) GetChannelByID(channelID string) (*models.Channel, error) {
	var channel models.Channel
	err := cr.odm.FindByID(context.Background(), channelID, &channel)
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

// CreateChannel 創建新頻道
func (cr *ChannelRepository) CreateChannel(channel *models.Channel) error {
	return cr.odm.Create(context.Background(), channel)
}

// UpdateChannel 更新頻道
func (cr *ChannelRepository) UpdateChannel(channelID string, updates map[string]interface{}) error {
	var channel models.Channel
	err := cr.odm.FindByID(context.Background(), channelID, &channel)
	if err != nil {
		return err
	}

	// 將map轉換為bson.M
	bsonUpdates := make(map[string]interface{})
	for k, v := range updates {
		bsonUpdates[k] = v
	}

	return cr.odm.UpdateFields(context.Background(), &channel, bsonUpdates)
}

// DeleteChannel 刪除頻道
func (cr *ChannelRepository) DeleteChannel(channelID string) error {
	var channel models.Channel
	return cr.odm.DeleteByID(context.Background(), channelID, &channel)
}

// CheckChannelExists 檢查頻道是否存在
func (cr *ChannelRepository) CheckChannelExists(channelID string) (bool, error) {
	channelObjID, err := primitive.ObjectIDFromHex(channelID)
	if err != nil {
		return false, err
	}

	qb := providers.NewQueryBuilder()
	qb.Where("_id", channelObjID)

	var channel models.Channel
	err = cr.odm.FindOne(context.Background(), qb.GetFilter(), &channel)
	if err != nil {
		return false, nil // 沒有找到視為不存在
	}
	return true, nil
}
