package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServerRepository struct {
	config       *config.Config
	mongoConnect *mongo.Database
}

func NewServerRepository(cfg *config.Config, mongodb *mongo.Database) *ServerRepository {
	return &ServerRepository{
		config:       cfg,
		mongoConnect: mongodb,
	}
}

func (sr *ServerRepository) GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error) {
	var servers []models.Server
	var collection = sr.mongoConnect.Collection("servers")

	cursor, err := collection.Find(context.Background(), bson.M{"members.user_id": objectID})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.Background(), &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (sr *ServerRepository) AddServer(server *models.Server) (models.Server, error) {
	var collection = sr.mongoConnect.Collection("servers")

	// 新建測試用戶伺服器關聯
	_, err := collection.InsertOne(context.Background(), server)
	if err != nil {
		return models.Server{}, err
	}
	log.Println("Server added: %v", server)
	return *server, nil
}
