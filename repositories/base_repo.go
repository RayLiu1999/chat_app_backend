package repositories

import (
	"chat_app_backend/config"

	"go.mongodb.org/mongo-driver/mongo"
)

type BaseRepository struct {
	Config       *config.Config
	MongoConnect *mongo.Database
}

func NewBaseRepository(cfg *config.Config, mongodb *mongo.Database) *BaseRepository {
	return &BaseRepository{
		Config:       cfg,
		MongoConnect: mongodb,
	}
}
