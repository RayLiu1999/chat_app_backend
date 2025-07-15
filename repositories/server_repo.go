package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"
	"log"
	"time"

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

func (sr *ServerRepository) GetServerListByUserId(userID string) ([]models.Server, error) {
	var servers []models.Server
	var collection = sr.mongoConnect.Collection("servers")

	userObjectId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(context.Background(), bson.M{"members.user_id": userObjectId})
	if err != nil {
		return nil, err
	}
	if err := cursor.All(context.Background(), &servers); err != nil {
		return nil, err
	}
	return servers, nil
}

func (sr *ServerRepository) CreateServer(server *models.Server) (models.Server, error) {
	collection := sr.mongoConnect.Collection("servers")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, server)
	if err != nil {
		log.Printf("保存伺服器失敗: %v", err)
		return models.Server{}, err
	}

	// 將插入結果的ID轉換為ObjectID
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		log.Printf("無法獲取插入的伺服器ID")
		return models.Server{}, err
	}

	// 更新伺服器ID
	server.ID = id

	return *server, nil
}
