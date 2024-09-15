package controllers

import (
	"chat_app_backend/config"

	"go.mongodb.org/mongo-driver/mongo"
)

type BaseController struct {
	Config       *config.Config
	MongoConnect *mongo.Database
}

func NewBaseController(cfg *config.Config, mongodb *mongo.Database) *BaseController {
	return &BaseController{
		Config:       cfg,
		MongoConnect: mongodb,
	}
}
