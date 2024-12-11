package repositories

import (
	"chat_app_backend/models"
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (br *BaseRepository) GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error) {
	var servers []models.Server
	var collection = br.MongoConnect.Collection("servers")

	cursor, err := collection.Find(context.Background(), bson.M{"members.user_id": objectID})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.Background(), &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (br *BaseRepository) AddServer(server *models.Server) (models.Server, error) {
	var collection = br.MongoConnect.Collection("servers")

	// 新建測試用戶伺服器關聯
	_, err := collection.InsertOne(context.Background(), server)
	if err != nil {
		return models.Server{}, err
	}
	log.Println("Server added: %v", server)
	return *server, nil
}
