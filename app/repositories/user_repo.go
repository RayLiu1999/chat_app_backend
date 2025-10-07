package repositories

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"context"
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type userRepository struct {
	config *config.Config
	odm    providers.ODM
	cache  providers.CacheProvider
}

func NewUserRepository(cfg *config.Config, odm providers.ODM, cache providers.CacheProvider) *userRepository {
	return &userRepository{
		config: cfg,
		odm:    odm,
		cache:  cache,
	}
}

// GetUserById 根據用戶 ID 獲取用戶信息，優先從快取中獲取
func (ur *userRepository) GetUserById(userID string) (*models.User, error) {
	cacheKey := utils.UserProfileCacheKey(userID)

	// 1. 從快取中獲取用戶
	cachedUser, err := ur.cache.Get(cacheKey)
	if err != nil {
		// 紀錄快取錯誤但繼續從資料庫獲取
	}

	utils.PrettyPrintf("Cache hit for user %s: %v\n", userID, cachedUser != "")

	if cachedUser != "" {
		var user models.User
		if err := json.Unmarshal([]byte(cachedUser), &user); err == nil {
			return &user, nil
		}
	}

	// 2. 從資料庫中獲取用戶
	var user models.User
	err = ur.odm.FindByID(context.Background(), userID, &user)
	if err != nil {
		return nil, err
	}

	// 3. 將用戶存入快取
	userBytes, err := json.Marshal(user)
	if err == nil {
		ur.cache.Set(cacheKey, string(userBytes), time.Hour*1) // Cache for 1 hour
	}

	return &user, nil
}

// GetUserByUsername 根據用戶名獲取用戶信息（測試用）
func (ur *userRepository) GetUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := ur.odm.FindOne(context.Background(), bson.M{"username": username}, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserListByIds 根據用戶 ID 列表獲取用戶列表
func (ur *userRepository) GetUserListByIds(userIds []string) ([]models.User, error) {
	var users []models.User

	userIdsObjectIds := make([]primitive.ObjectID, len(userIds))
	for i, userId := range userIds {
		cacheKey := utils.UserProfileCacheKey(userId)

		// 1. 從快取中獲取用戶
		cachedUser, err := ur.cache.Get(cacheKey)
		if err != nil {
			// 紀錄快取錯誤但繼續從資料庫獲取
		}

		if cachedUser != "" {
			var user models.User
			if err := utils.JSONToStruct(cachedUser, &user); err == nil {
				users = append(users, user)

				utils.PrettyPrintf("Cache hit for user %s: %v\n", userId, cachedUser != "")
			}
		}

		userIdsObjectIds[i], _ = primitive.ObjectIDFromHex(userId)
	}

	// 如果快取命中所有用戶，直接返回
	if len(users) == len(userIds) {
		return users, nil
	}

	// 從資料庫中查詢剩餘的用戶
	err := ur.odm.Find(context.Background(), bson.M{"_id": bson.M{"$in": userIdsObjectIds}}, &users)
	if err != nil {
		return nil, err
	}

	// 將查詢到的用戶存入快取
	for _, user := range users {
		cacheKey := utils.UserProfileCacheKey(user.ID.Hex())
		userBytes, err := json.Marshal(user)
		if err == nil {
			ur.cache.Set(cacheKey, string(userBytes), time.Hour*1) // Cache for 1 hour
		}
	}

	return users, nil
}

func (ur *userRepository) CreateUser(user models.User) error {
	return ur.odm.Create(context.Background(), &user)
}

func (ur *userRepository) CheckEmailExists(email string) (bool, error) {
	var user models.User
	err := ur.odm.FindOne(context.Background(), bson.M{"email": email}, &user)
	if err == providers.ErrDocumentNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (ur *userRepository) CheckUsernameExists(username string) (bool, error) {
	var user models.User
	err := ur.odm.FindOne(context.Background(), bson.M{"username": username}, &user)
	if err == providers.ErrDocumentNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateUserOnlineStatus 更新用戶在線狀態
func (ur *userRepository) UpdateUserOnlineStatus(userID string, isOnline bool) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"is_online":      isOnline,
			"last_active_at": time.Now().Unix(),
			"updated_at":     time.Now(),
		},
	}

	// 更新上線狀態後移除快取
	defer ur.cache.Delete(utils.UserStatusCacheKey(userID))

	return ur.odm.UpdateMany(context.Background(), &models.User{}, filter, update)
}

// UpdateUserLastActiveTime 更新用戶最後活動時間
func (ur *userRepository) UpdateUserLastActiveTime(userID string, timestamp int64) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"last_active_at": timestamp,
			"updated_at":     time.Now(),
		},
	}

	// 更新後移除快取
	defer ur.cache.Delete(utils.UserActivityThrottleCacheKey(userID))

	return ur.odm.UpdateMany(context.Background(), &models.User{}, filter, update)
}

// UpdateUser 更新用戶信息
func (ur *userRepository) UpdateUser(userID string, updates map[string]any) error {
	ctx := context.Background()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": userObjectID}
	update := bson.M{"$set": updates}

	// 更新後移除快取
	defer ur.cache.Delete(utils.UserProfileCacheKey(userID))

	return ur.odm.UpdateMany(ctx, &models.User{}, filter, update)
}

// DeleteUser 刪除用戶
func (ur *userRepository) DeleteUser(userID string) error {
	// 刪除後移除快取
	defer ur.cache.Delete(utils.UserProfileCacheKey(userID))
	return ur.odm.DeleteByID(context.Background(), userID, &models.User{})
}
