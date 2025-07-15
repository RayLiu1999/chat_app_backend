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

func (fr *FriendRepository) GetFriendById(userID string) (*models.Friend, error) {
	var friend models.Friend
	var collection = fr.mongoConnect.Collection("friends")

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": userObjectID}).Decode(&friend)
	if err != nil {
		return nil, err
	}

	return &friend, nil
}
