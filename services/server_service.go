package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ServerService struct {
	config       *config.Config
	mongoConnect *mongo.Database
	serverRepo   ServerServiceInterface
}

func NewServerService(cfg *config.Config, mongodb *mongo.Database, serverRepo ServerServiceInterface) *ServerService {
	return &ServerService{
		config:       cfg,
		mongoConnect: mongodb,
		serverRepo:   serverRepo,
	}
}

// 添加新伺服器
func (ss *ServerService) CreateServer(server *models.Server) (models.Server, error) {
	return ss.serverRepo.CreateServer(server)
}

// GetServerListByUserId 獲取用戶的伺服器列表
func (ss *ServerService) GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error) {
	servers, err := ss.serverRepo.GetServerListByUserId(objectID)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
