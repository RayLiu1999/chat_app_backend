package repositories

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChannelCategoryRepository struct {
	odm *providers.ODM
}

func NewChannelCategoryRepository(odm *providers.ODM) *ChannelCategoryRepository {
	return &ChannelCategoryRepository{
		odm: odm,
	}
}

// CreateChannelCategory 創建頻道類別
func (r *ChannelCategoryRepository) CreateChannelCategory(category *models.ChannelCategory) error {
	ctx := context.Background()
	err := r.odm.Create(ctx, category)
	if err != nil {
		return fmt.Errorf("創建頻道類別失敗: %v", err)
	}
	return nil
}

// GetChannelCategoriesByServerID 根據伺服器ID獲取頻道類別列表
func (r *ChannelCategoryRepository) GetChannelCategoriesByServerID(serverID string) ([]models.ChannelCategory, error) {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return nil, fmt.Errorf("無效的伺服器ID: %v", err)
	}

	ctx := context.Background()
	var categories []models.ChannelCategory
	filter := bson.M{"server_id": serverObjectID}

	err = r.odm.Find(ctx, filter, &categories)
	if err != nil {
		return nil, fmt.Errorf("查詢頻道類別失敗: %v", err)
	}

	return categories, nil
}

// GetChannelCategoryByID 根據類別ID獲取頻道類別
func (r *ChannelCategoryRepository) GetChannelCategoryByID(categoryID string) (*models.ChannelCategory, error) {
	ctx := context.Background()
	var category models.ChannelCategory

	err := r.odm.FindByID(ctx, categoryID, &category)
	if err != nil {
		return nil, fmt.Errorf("查詢頻道類別失敗: %v", err)
	}

	return &category, nil
}

// UpdateChannelCategory 更新頻道類別
func (r *ChannelCategoryRepository) UpdateChannelCategory(categoryID string, updates map[string]any) error {
	categoryObjectID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return fmt.Errorf("無效的類別ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{"_id": categoryObjectID}
	updateDoc := bson.M{"$set": updates}

	err = r.odm.UpdateMany(ctx, &models.ChannelCategory{}, filter, updateDoc)
	if err != nil {
		return fmt.Errorf("更新頻道類別失敗: %v", err)
	}

	return nil
}

// DeleteChannelCategory 刪除頻道類別
func (r *ChannelCategoryRepository) DeleteChannelCategory(categoryID string) error {
	ctx := context.Background()

	err := r.odm.DeleteByID(ctx, categoryID, &models.ChannelCategory{})
	if err != nil {
		return fmt.Errorf("刪除頻道類別失敗: %v", err)
	}

	return nil
}

// CheckChannelCategoryExists 檢查頻道類別是否存在
func (r *ChannelCategoryRepository) CheckChannelCategoryExists(categoryID string) (bool, error) {
	ctx := context.Background()

	exists, err := r.odm.ExistsByID(ctx, categoryID, &models.ChannelCategory{})
	if err != nil {
		return false, nil // 不存在
	}

	return exists, nil
}
