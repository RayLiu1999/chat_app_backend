package services

import (
	"chat_app_backend/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 取得伺服器列表
func (bs *BaseService) GetServerListByUserId(objectID primitive.ObjectID) ([]models.Server, error) {
	servers, err := bs.repo.GetServerListByUserId(objectID)
	if err != nil {
		return nil, err
	}

	return servers, nil
}

func (bs *BaseService) AddServer(server *models.Server) (models.Server, error) {
	_, err := bs.repo.AddServer(server)
	if err != nil {
		return models.Server{}, err
	}

	return *server, nil
}
