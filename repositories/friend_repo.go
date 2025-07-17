package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FriendRepository struct {
	config *config.Config
	odm    *providers.ODM
}

func NewFriendRepository(cfg *config.Config, odm *providers.ODM) *FriendRepository {
	return &FriendRepository{
		config: cfg,
		odm:    odm,
	}
}

func (fr *FriendRepository) GetFriendById(userID string) (*models.Friend, error) {
	var friend models.Friend

	userObjectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	err = fr.odm.FindOne(context.Background(), bson.M{"_id": userObjectID}, &friend)
	if err != nil {
		return nil, err
	}

	return &friend, nil
}
