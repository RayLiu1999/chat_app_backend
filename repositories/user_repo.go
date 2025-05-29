package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	config       *config.Config
	mongoConnect *mongo.Database
}

func NewUserRepository(cfg *config.Config, mongodb *mongo.Database) *UserRepository {
	return &UserRepository{
		config:       cfg,
		mongoConnect: mongodb,
	}
}

func (ur *UserRepository) GetUserById(objectId primitive.ObjectID) (*models.User, error) {
	var user models.User
	var collection = ur.mongoConnect.Collection("users")

	err := collection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (ur *UserRepository) GetUserListByIds(objectIds []primitive.ObjectID) ([]models.User, error) {
	var users []models.User
	var collection = ur.mongoConnect.Collection("users")

	cursor, err := collection.Find(context.Background(), bson.M{"_id": bson.M{"$in": objectIds}})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.Background(), &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}
