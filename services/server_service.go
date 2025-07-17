package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
)

type ServerService struct {
	config     *config.Config
	serverRepo repositories.ServerRepositoryInterface
	odm        *providers.ODM
}

func NewServerService(cfg *config.Config, odm *providers.ODM, serverRepo repositories.ServerRepositoryInterface) *ServerService {
	return &ServerService{
		config:     cfg,
		serverRepo: serverRepo,
		odm:        odm,
	}
}

// 添加新伺服器
func (ss *ServerService) CreateServer(server *models.Server) (models.Server, error) {
	return ss.serverRepo.CreateServer(server)
}

// GetServerListByUserId 獲取用戶的伺服器列表
func (ss *ServerService) GetServerListByUserId(userID string) ([]models.Server, error) {
	servers, err := ss.serverRepo.GetServerListByUserId(userID)
	if err != nil {
		return nil, err
	}

	return servers, nil
}
