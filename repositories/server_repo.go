package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"context"
	"log"

	"chat_app_backend/providers"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ServerRepository struct {
	config *config.Config
	odm    *providers.ODM
	// queryBuilder *providers.QueryBuilder // 如有需要可加
}

func NewServerRepository(cfg *config.Config, odm *providers.ODM) *ServerRepository {
	return &ServerRepository{
		config: cfg,
		odm:    odm,
		// queryBuilder: qb, // 如有需要
	}
}

func (sr *ServerRepository) GetServerListByUserId(userID string) ([]models.Server, error) {
	var servers []models.Server
	userObjectId, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	err = sr.odm.Find(context.Background(), bson.M{"members.user_id": userObjectId}, &servers)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func (sr *ServerRepository) CreateServer(server *models.Server) (models.Server, error) {
	err := sr.odm.InsertOne(context.Background(), server)
	if err != nil {
		log.Printf("保存伺服器失敗: %v", err)
		return models.Server{}, err
	}
	return *server, nil
}
