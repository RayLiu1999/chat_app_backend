package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"
	"log"

	"chat_app_backend/providers"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerRepository struct {
	config *config.Config
	odm    *providers.ODM
	// queryBuilder *providers.QueryBuilder // 如有需要可加
}

func NewServerRepository(cfg *config.Config, odm *providers.ODM) *ServerRepository {
	return &ServerRepository{
		config: cfg,
		odm:    odm,
		// queryBuilder: qb, // 如有需要
	}
}

func (sr *ServerRepository) CreateServer(server *models.Server) (models.Server, error) {
	err := sr.odm.Create(context.Background(), server)
	if err != nil {
		log.Printf("保存伺服器失敗: %v", err)
		return models.Server{}, err
	}
	return *server, nil
}

// SearchPublicServers 搜尋公開伺服器
func (sr *ServerRepository) SearchPublicServers(request models.ServerSearchRequest) ([]models.Server, int64, error) {
	ctx := context.Background()

	// 建構搜尋條件
	filter := bson.M{"is_public": true}

	// 如果有搜尋關鍵字，加入文字搜尋條件
	if request.Query != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": request.Query, "$options": "i"}},
			{"description": bson.M{"$regex": request.Query, "$options": "i"}},
			{"tags": bson.M{"$in": []string{request.Query}}},
		}
	}

	// 計算總數
	totalCount, err := sr.odm.Count(ctx, filter, &models.Server{})
	if err != nil {
		return nil, 0, err
	}

	// 設定預設值
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Limit <= 0 {
		request.Limit = 20
	}
	if request.Limit > 100 {
		request.Limit = 100 // 限制最大數量
	}

	// 設定排序
	sortField := "created_at"
	sortOrder := -1 // 預設降序

	if request.SortBy != "" {
		switch request.SortBy {
		case "name":
			sortField = "name"
		case "members":
			sortField = "member_count"
		case "created_at":
			sortField = "created_at"
		}
	}

	if request.SortOrder == "asc" {
		sortOrder = 1
	}

	// 建構查詢選項
	skip := int64((request.Page - 1) * request.Limit)
	limit := int64(request.Limit)

	options := &providers.QueryOptions{
		Skip:  &skip,
		Limit: &limit,
		Sort:  bson.D{{Key: sortField, Value: sortOrder}},
	}

	var servers []models.Server
	err = sr.odm.FindWithOptions(ctx, filter, &servers, options)
	if err != nil {
		return nil, 0, err
	}

	return servers, totalCount, nil
}

// GetServerWithOwnerInfo 獲取包含擁有者信息的伺服器
func (sr *ServerRepository) GetServerWithOwnerInfo(serverID string) (*models.Server, *models.User, error) {
	ctx := context.Background()

	// 獲取伺服器信息
	var server models.Server
	err := sr.odm.FindByID(ctx, serverID, &server)
	if err != nil {
		return nil, nil, err
	}

	// 獲取擁有者信息
	var owner models.User
	err = sr.odm.FindByID(ctx, server.OwnerID.Hex(), &owner)
	if err != nil {
		return &server, nil, err
	}

	return &server, &owner, nil
}

// CheckUserInServer 檢查用戶是否在伺服器中
func (sr *ServerRepository) CheckUserInServer(userID, serverID string) (bool, error) {
	// 注意：這個方法現在應該使用 ServerMemberRepository.IsMemberOfServer
	// 這裡只檢查是否為擁有者
	ctx := context.Background()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, err
	}

	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return false, err
	}

	filter := bson.M{
		"_id":      serverObjectID,
		"owner_id": userObjectID,
	}

	return sr.odm.Exists(ctx, filter, &models.Server{})
}

// GetServerByID 根據ID獲取伺服器
func (sr *ServerRepository) GetServerByID(serverID string) (*models.Server, error) {
	ctx := context.Background()
	var server models.Server

	err := sr.odm.FindByID(ctx, serverID, &server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

// UpdateServer 更新伺服器信息
func (sr *ServerRepository) UpdateServer(serverID string, updates map[string]interface{}) error {
	ctx := context.Background()
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": serverObjectID}
	updateDoc := bson.M{"$set": updates}

	return sr.odm.UpdateMany(ctx, &models.Server{}, filter, updateDoc)
}

// DeleteServer 刪除伺服器
func (sr *ServerRepository) DeleteServer(serverID string) error {
	ctx := context.Background()
	return sr.odm.DeleteByID(ctx, serverID, &models.Server{})
}

// UpdateMemberCount 更新成員數量快取
func (sr *ServerRepository) UpdateMemberCount(serverID string, count int) error {
	updates := map[string]interface{}{
		"member_count": count,
	}
	return sr.UpdateServer(serverID, updates)
}
