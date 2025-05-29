package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendRepository struct {
	config       *config.Config
	mongoConnect *mongo.Database
}

func NewFriendRepository(cfg *config.Config, mongodb *mongo.Database) *FriendRepository {
	return &FriendRepository{
		config:       cfg,
		mongoConnect: mongodb,
	}
}

func (fr *FriendRepository) GetFriendById(objectId primitive.ObjectID) (*models.Friend, error) {
	var friend models.Friend
	var collection = fr.mongoConnect.Collection("friends")

	err := collection.FindOne(context.Background(), bson.M{"_id": objectId}).Decode(&friend)
	if err != nil {
		return nil, err
	}

	return &friend, nil
}
