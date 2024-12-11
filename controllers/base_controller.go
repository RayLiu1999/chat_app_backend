package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/services"

	"go.mongodb.org/mongo-driver/mongo"
)

type BaseController struct {
	Config       *config.Config
	MongoConnect *mongo.Database
	service      *services.BaseService
}

func NewBaseController(cfg *config.Config, mongodb *mongo.Database) *BaseController {
	return &BaseController{
		Config:       cfg,
		MongoConnect: mongodb,
		service:      services.NewBaseService(cfg, mongodb),
	}
}
