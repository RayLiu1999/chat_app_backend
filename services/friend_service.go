package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendService struct {
	config       *config.Config
	mongoConnect *mongo.Database
	friendRepo   repositories.FriendRepositoryInterface
}

func NewFriendService(cfg *config.Config, mongodb *mongo.Database, friendRepo repositories.FriendRepositoryInterface) *FriendService {
	return &FriendService{
		config:       cfg,
		mongoConnect: mongodb,
		friendRepo:   friendRepo,
	}
}

func (fs *FriendService) GetFriendById(objectID primitive.ObjectID) (*models.Friend, error) {
	friend, err := fs.friendRepo.GetFriendById(objectID)
	if err != nil {
		return nil, err
	}

	return friend, nil
}
