package di

import (
	"chat_app_backend/config"
	"chat_app_backend/controllers"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/services"
)

// 儲存庫容器
type RepositoryContainer struct {
	UserRepo   repositories.UserRepositoryInterface
	ChatRepo   repositories.ChatRepositoryInterface
	ServerRepo repositories.ServerRepositoryInterface
	FriendRepo repositories.FriendRepositoryInterface
}

// 服務容器
type ServiceContainer struct {
	UserService   services.UserServiceInterface
	ChatService   services.ChatServiceInterface
	ServerService services.ServerServiceInterface
	FriendService services.FriendServiceInterface
}

// 控制器容器
type ControllerContainer struct {
	UserController   *controllers.UserController
	ChatController   *controllers.ChatController
	ServerController *controllers.ServerController
	FriendController *controllers.FriendController
}

// 初始化儲存庫
func initRepositories(cfg *config.Config, mongodb *providers.MongoWrapper) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepo:   repositories.NewUserRepository(cfg, mongodb.DB),
		ChatRepo:   repositories.NewChatRepository(cfg, mongodb.DB),
		ServerRepo: repositories.NewServerRepository(cfg, mongodb.DB),
		FriendRepo: repositories.NewFriendRepository(cfg, mongodb.DB),
	}
}

// 初始化服務
func initServices(cfg *config.Config, mongodb *providers.MongoWrapper, repos *RepositoryContainer) *ServiceContainer {
	return &ServiceContainer{
		UserService:   services.NewUserService(cfg, mongodb.DB, repos.UserRepo),
		ChatService:   services.NewChatService(cfg, mongodb.DB, repos.ChatRepo, repos.ServerRepo, repos.UserRepo),
		ServerService: services.NewServerService(cfg, mongodb.DB, repos.ServerRepo),
		FriendService: services.NewFriendService(cfg, mongodb.DB, repos.FriendRepo, repos.UserRepo),
	}
}

// 初始化控制器
func initControllers(cfg *config.Config, mongodb *providers.MongoWrapper, services *ServiceContainer, repos *RepositoryContainer) *ControllerContainer {
	return &ControllerContainer{
		UserController:   controllers.NewUserController(cfg, mongodb.DB, services.UserService),
		ChatController:   controllers.NewChatController(cfg, mongodb.DB, services.ChatService, services.UserService),
		ServerController: controllers.NewServerController(cfg, mongodb.DB, services.ServerService, repos.UserRepo),
		FriendController: controllers.NewFriendController(cfg, mongodb.DB, services.FriendService),
	}
}

// 創建應用程式依賴
func BuildDependencies(cfg *config.Config, mongodb *providers.MongoWrapper) *ControllerContainer {
	repos := initRepositories(cfg, mongodb)
	services := initServices(cfg, mongodb, repos)
	return initControllers(cfg, mongodb, services, repos)
}
