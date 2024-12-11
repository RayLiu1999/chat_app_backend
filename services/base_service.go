package services

import (
	"chat_app_backend/config"
	"chat_app_backend/repositories"

	"go.mongodb.org/mongo-driver/mongo"
)

type BaseService struct {
	Config       *config.Config
	MongoConnect *mongo.Database
	repo         *repositories.BaseRepository
}

func NewBaseService(cfg *config.Config, mongodb *mongo.Database) *BaseService {
	return &BaseService{
		Config:       cfg,
		MongoConnect: mongodb,
		repo:         repositories.NewBaseRepository(cfg, mongodb),
	}
}
