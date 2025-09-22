package repositories

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"chat_app_backend/config"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository struct {
	config *config.Config
	odm    *providers.ODM
}

func NewUserRepository(cfg *config.Config, odm *providers.ODM) *UserRepository {
	return &UserRepository{
		config: cfg,
		odm:    odm,
	}
}

func (ur *UserRepository) GetUserById(userID string) (*models.User, error) {
	var user models.User

	err := ur.odm.FindByID(context.Background(), userID, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) GetUserListByIds(userIds []string) ([]models.User, error) {
	var users []models.User

	userIdsObjectIds := make([]primitive.ObjectID, len(userIds))
	for i, userId := range userIds {
		userIdsObjectIds[i], _ = primitive.ObjectIDFromHex(userId)
	}

	err := ur.odm.Find(context.Background(), bson.M{"_id": bson.M{"$in": userIdsObjectIds}}, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (ur *UserRepository) CreateUser(user models.User) error {
	return ur.odm.Create(context.Background(), &user)
}

func (ur *UserRepository) CheckEmailExists(email string) (bool, error) {
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

func (ur *UserRepository) CheckUsernameExists(username string) (bool, error) {
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
func (ur *UserRepository) UpdateUserOnlineStatus(userID string, isOnline bool) error {
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

	return ur.odm.UpdateMany(context.Background(), &models.User{}, filter, update)
}

// UpdateUserLastActiveTime 更新用戶最後活動時間
func (ur *UserRepository) UpdateUserLastActiveTime(userID string, timestamp int64) error {
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

	return ur.odm.UpdateMany(context.Background(), &models.User{}, filter, update)
}

// UpdateUser 更新用戶信息
func (ur *UserRepository) UpdateUser(userID string, updates map[string]interface{}) error {
	ctx := context.Background()

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": userObjectID}
	update := bson.M{"$set": updates}

	return ur.odm.UpdateMany(ctx, &models.User{}, filter, update)
}

// DeleteUser 刪除用戶
func (ur *UserRepository) DeleteUser(userID string) error {
	return ur.odm.DeleteByID(context.Background(), userID, &models.User{})
}
