package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"context"

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
