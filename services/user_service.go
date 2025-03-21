package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserService struct {
	config       *config.Config
	mongoConnect *mongo.Database
	userRepo     repositories.UserRepositoryInterface
}

func NewUserService(cfg *config.Config, mongodb *mongo.Database, userRepo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		config:       cfg,
		mongoConnect: mongodb,
		userRepo:     userRepo,
	}
}

func (us *UserService) GetUserById(objectID primitive.ObjectID) (*models.User, error) {
	user, err := us.userRepo.GetUserById(objectID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
