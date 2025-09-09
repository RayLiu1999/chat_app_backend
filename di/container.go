package di

import (
	"chat_app_backend/config"
	"chat_app_backend/controllers"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/services"
)

// Repository容器
type RepositoryContainer struct {
	UserRepo            repositories.UserRepositoryInterface
	ChatRepo            repositories.ChatRepositoryInterface
	ServerRepo          repositories.ServerRepositoryInterface
	ServerMemberRepo    repositories.ServerMemberRepositoryInterface
	FriendRepo          repositories.FriendRepositoryInterface
	ChannelRepo         repositories.ChannelRepositoryInterface
	ChannelCategoryRepo repositories.ChannelCategoryRepositoryInterface
	FileRepo            repositories.FileRepositoryInterface
}

// Service容器
type ServiceContainer struct {
	UserService       services.UserServiceInterface
	ChatService       services.ChatServiceInterface
	ServerService     services.ServerServiceInterface
	FriendService     services.FriendServiceInterface
	ChannelService    services.ChannelServiceInterface
	FileUploadService services.FileUploadServiceInterface
}

// Controller容器
type ControllerContainer struct {
	HealthController  *controllers.HealthController
	UserController    *controllers.UserController
	ChatController    *controllers.ChatController
	ServerController  *controllers.ServerController
	FriendController  *controllers.FriendController
	ChannelController *controllers.ChannelController
	FileController    *controllers.FileController
}

// Providers容器
type ProviderContainer struct {
	ODM          *providers.ODM
	FileProvider providers.FileProviderInterface
}

// 初始化Repositories
func initRepositories(cfg *config.Config, providers *ProviderContainer) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepo:            repositories.NewUserRepository(cfg, providers.ODM),
		ChatRepo:            repositories.NewChatRepository(cfg, providers.ODM),
		ServerRepo:          repositories.NewServerRepository(cfg, providers.ODM),
		ServerMemberRepo:    repositories.NewServerMemberRepository(providers.ODM),
		FriendRepo:          repositories.NewFriendRepository(cfg, providers.ODM),
		ChannelRepo:         repositories.NewChannelRepository(cfg, providers.ODM),
		ChannelCategoryRepo: repositories.NewChannelCategoryRepository(providers.ODM),
		FileRepo:            repositories.NewFileRepository(cfg, providers.ODM),
	}
}

// 初始化Services
func initServices(cfg *config.Config, providers *ProviderContainer, repos *RepositoryContainer) *ServiceContainer {
	// 先初始化檔案上傳服務
	fileUploadService := services.NewFileUploadService(cfg, providers.FileProvider, providers.ODM, repos.FileRepo)

	// 先創建一個臨時的 UserService（沒有 ClientManager）
	tempUserService := services.NewUserService(cfg, providers.ODM, repos.UserRepo, nil, fileUploadService)

	// 創建 ChatService（會初始化 ClientManager）
	chatService := services.NewChatService(cfg, providers.ODM, repos.ChatRepo, repos.ServerRepo, repos.ServerMemberRepo, repos.UserRepo, tempUserService, fileUploadService)

	// 獲取 ClientManager 並創建最終的 UserService
	clientManager := chatService.GetClientManager()
	finalUserService := services.NewUserService(cfg, providers.ODM, repos.UserRepo, clientManager, fileUploadService)

	// 更新 ChatService 中的 UserService 引用
	chatService.UpdateUserService(finalUserService)

	// 創建其他服務並傳入最終的 UserService
	serverService := services.NewServerService(cfg, providers.ODM, repos.ServerRepo, repos.ServerMemberRepo, repos.UserRepo, repos.ChannelRepo, repos.ChannelCategoryRepo, repos.ChatRepo, fileUploadService, finalUserService)

	// 更新 ServerService 中的 UserService 引用
	serverService.UpdateUserService(finalUserService)

	return &ServiceContainer{
		UserService:   finalUserService,
		ChatService:   chatService,
		ServerService: serverService,
		FriendService: services.NewFriendService(
			cfg,
			providers.ODM,
			repos.FriendRepo,
			repos.UserRepo,
			finalUserService,
			fileUploadService),
		ChannelService: services.NewChannelService(
			cfg,
			providers.ODM,
			repos.ChannelRepo,
			repos.ServerRepo,
			repos.ServerMemberRepo,
			repos.UserRepo,
			repos.ChatRepo),
		FileUploadService: fileUploadService,
	}
}

// 初始化Controllers
func initControllers(cfg *config.Config, mongodb *providers.MongoWrapper, services *ServiceContainer, repos *RepositoryContainer) *ControllerContainer {
	return &ControllerContainer{
		HealthController:  controllers.NewHealthController(cfg, mongodb),
		UserController:    controllers.NewUserController(cfg, mongodb.DB, services.UserService),
		ChatController:    controllers.NewChatController(cfg, mongodb.DB, services.ChatService, services.UserService),
		ServerController:  controllers.NewServerController(cfg, mongodb.DB, services.ServerService),
		FriendController:  controllers.NewFriendController(cfg, mongodb.DB, services.FriendService),
		ChannelController: controllers.NewChannelController(cfg, mongodb.DB, services.ChannelService),
		FileController:    controllers.NewFileController(cfg, mongodb.DB, services.FileUploadService),
	}
}

// 初始化Providers
func initProviders(cfg *config.Config, mongodb *providers.MongoWrapper) *ProviderContainer {
	return &ProviderContainer{
		ODM:          providers.NewODM(mongodb.DB),
		FileProvider: providers.NewFileProvider(cfg),
	}
}

// ApplicationContainer 包含所有依賴
type ApplicationContainer struct {
	Controllers *ControllerContainer
	Services    *ServiceContainer
	Repos       *RepositoryContainer
	Providers   *ProviderContainer
}

// 創建應用程式依賴
func BuildDependencies(cfg *config.Config, mongodb *providers.MongoWrapper) *ApplicationContainer {
	providers := initProviders(cfg, mongodb)
	repos := initRepositories(cfg, providers)
	services := initServices(cfg, providers, repos)
	controllers := initControllers(cfg, mongodb, services, repos)

	return &ApplicationContainer{
		Controllers: controllers,
		Services:    services,
		Repos:       repos,
		Providers:   providers,
	}
}
