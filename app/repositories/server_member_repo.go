package repositories

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerMemberRepository struct {
	odm *providers.ODM
}

func NewServerMemberRepository(odm *providers.ODM) *ServerMemberRepository {
	return &ServerMemberRepository{
		odm: odm,
	}
}

// AddMemberToServer 將用戶添加到伺服器
func (smr *ServerMemberRepository) AddMemberToServer(serverID, userID string, role string) error {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return fmt.Errorf("無效的伺服器ID: %v", err)
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("無效的用戶ID: %v", err)
	}

	// 檢查是否已經是成員
	exists, err := smr.IsMemberOfServer(serverID, userID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("用戶已經是該伺服器成員")
	}

	ctx := context.Background()
	member := &models.ServerMember{
		ServerID:     serverObjectID,
		UserID:       userObjectID,
		Role:         role,
		JoinedAt:     time.Now(),
		LastActiveAt: time.Now(),
	}

	return smr.odm.Create(ctx, member)
}

// RemoveMemberFromServer 從伺服器移除用戶
func (smr *ServerMemberRepository) RemoveMemberFromServer(serverID, userID string) error {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return fmt.Errorf("無效的伺服器ID: %v", err)
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("無效的用戶ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{
		"server_id": serverObjectID,
		"user_id":   userObjectID,
	}

	return smr.odm.DeleteMany(ctx, &models.ServerMember{}, filter)
}

// GetServerMembers 獲取伺服器所有成員
func (smr *ServerMemberRepository) GetServerMembers(serverID string, page, limit int) ([]models.ServerMember, int64, error) {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return nil, 0, fmt.Errorf("無效的伺服器ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{"server_id": serverObjectID}

	// 計算總數
	totalCount, err := smr.odm.Count(ctx, filter, &models.ServerMember{})
	if err != nil {
		return nil, 0, err
	}

	// 設定分頁
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}

	skip := int64((page - 1) * limit)
	limitInt64 := int64(limit)

	options := &providers.QueryOptions{
		Skip:  &skip,
		Limit: &limitInt64,
		Sort:  bson.D{{Key: "joined_at", Value: 1}}, // 按加入時間排序
	}

	var members []models.ServerMember
	err = smr.odm.FindWithOptions(ctx, filter, &members, options)
	if err != nil {
		return nil, 0, err
	}

	return members, totalCount, nil
}

// GetUserServers 獲取用戶加入的所有伺服器
func (smr *ServerMemberRepository) GetUserServers(userID string) ([]models.ServerMember, error) {
	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("無效的用戶ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{"user_id": userObjectID}

	var memberships []models.ServerMember
	err = smr.odm.Find(ctx, filter, &memberships)
	if err != nil {
		return nil, err
	}

	return memberships, nil
}

// IsMemberOfServer 檢查用戶是否為伺服器成員
func (smr *ServerMemberRepository) IsMemberOfServer(serverID, userID string) (bool, error) {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return false, fmt.Errorf("無效的伺服器ID: %v", err)
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return false, fmt.Errorf("無效的用戶ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{
		"server_id": serverObjectID,
		"user_id":   userObjectID,
	}

	return smr.odm.Exists(ctx, filter, &models.ServerMember{})
}

// UpdateMemberRole 更新成員角色
func (smr *ServerMemberRepository) UpdateMemberRole(serverID, userID, newRole string) error {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return fmt.Errorf("無效的伺服器ID: %v", err)
	}

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("無效的用戶ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{
		"server_id": serverObjectID,
		"user_id":   userObjectID,
	}

	update := bson.M{
		"$set": bson.M{
			"role":       newRole,
			"updated_at": time.Now(),
		},
	}

	return smr.odm.UpdateMany(ctx, &models.ServerMember{}, filter, update)
}

// GetMemberCount 獲取伺服器成員數量
func (smr *ServerMemberRepository) GetMemberCount(serverID string) (int64, error) {
	serverObjectID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		return 0, fmt.Errorf("無效的伺服器ID: %v", err)
	}

	ctx := context.Background()
	filter := bson.M{"server_id": serverObjectID}

	return smr.odm.Count(ctx, filter, &models.ServerMember{})
}
