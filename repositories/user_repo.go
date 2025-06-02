package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	config       *config.Config
	mongoConnect *mongo.Database
	odm          *providers.ODM
}

func NewUserRepository(cfg *config.Config, mongodb *mongo.Database) *UserRepository {
	return &UserRepository{
		config:       cfg,
		mongoConnect: mongodb,
		odm:          providers.NewODM(mongodb),
	}
}

func (ur *UserRepository) GetUserById(objectId primitive.ObjectID) (*models.User, error) {
	var user models.User

	err := ur.odm.FindByID(context.Background(), objectId, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) GetUserListByIds(objectIds []primitive.ObjectID) ([]models.User, error) {
	var users []models.User

	err := ur.odm.Find(context.Background(), bson.M{"_id": bson.M{"$in": objectIds}}, &users)
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
