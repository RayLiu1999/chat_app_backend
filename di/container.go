package di

import (
	"chat_app_backend/app/http/controllers"
	"chat_app_backend/app/providers"
	"chat_app_backend/app/repositories"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"fmt"
)

// Repository容器
type RepositoryContainer struct {
	UserRepo            repositories.UserRepository
	ChatRepo            repositories.ChatRepository
	ServerRepo          repositories.ServerRepository
	ServerMemberRepo    repositories.ServerMemberRepository
	FriendRepo          repositories.FriendRepository
	ChannelRepo         repositories.ChannelRepository
	ChannelCategoryRepo repositories.ChannelCategoryRepository
	FileRepo            repositories.FileRepository
}

// Service容器
type ServiceContainer struct {
	UserService       services.UserService
	ChatService       services.ChatService
	ServerService     services.ServerService
	FriendService     services.FriendService
	ChannelService    services.ChannelService
	FileUploadService services.FileUploadService
	ClientManager     services.ClientManager
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
	ODM          providers.ODM
	FileProvider providers.FileProvider
	Cache        providers.CacheProvider
}

// 初始化Repositories
func initRepositories(
	cfg *config.Config,
	providers *ProviderContainer,
) *RepositoryContainer {
	return &RepositoryContainer{
		UserRepo:            repositories.NewUserRepository(cfg, providers.ODM, providers.Cache),
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
func initServices(
	cfg *config.Config,
	providers *ProviderContainer,
	repos *RepositoryContainer,
	redis *providers.RedisWrapper,
) *ServiceContainer {
	// 1. 將 ClientManager 的初始化提前
	clientManager := services.NewClientManager()

	// 2. 初始化檔案上傳服務
	fileUploadService := services.NewFileUploadService(
		cfg,
		providers.FileProvider,
		providers.ODM,
		repos.FileRepo,
	)

	// 3. 現在可以直接創建最終的 UserService
	userService := services.NewUserService(
		cfg,
		providers.ODM,
		repos.UserRepo,
		fileUploadService,
	)

	// 4. 創建 ChatService，並傳入已經建立好的 UserService
	chatService := services.NewChatService(
		cfg,
		providers.ODM,
		redis.Client,
		providers.Cache,
		repos.ChatRepo,
		repos.ServerRepo,
		repos.ServerMemberRepo,
		repos.UserRepo,
		userService,
		fileUploadService,
	)

	// 5. 創建其他服務
	serverService := services.NewServerService(
		cfg,
		providers.ODM,
		repos.ServerRepo,
		repos.ServerMemberRepo,
		repos.UserRepo,
		repos.ChannelRepo,
		repos.ChannelCategoryRepo,
		repos.ChatRepo,
		fileUploadService,
		userService,
		clientManager,
	)
	friendService := services.NewFriendService(
		cfg,
		providers.ODM,
		repos.FriendRepo,
		repos.UserRepo,
		fileUploadService,
		clientManager,
	)
	channelService := services.NewChannelService(
		cfg,
		providers.ODM,
		repos.ChannelRepo,
		repos.ServerRepo,
		repos.ServerMemberRepo,
		repos.UserRepo,
		repos.ChatRepo,
	)

	return &ServiceContainer{
		UserService:       userService,
		ChatService:       chatService,
		ServerService:     serverService,
		FriendService:     friendService,
		ChannelService:    channelService,
		FileUploadService: fileUploadService,
		ClientManager:     clientManager,
	}
}

// 初始化Controllers
func initControllers(
	cfg *config.Config,
	mongodb *providers.MongoWrapper,
	services *ServiceContainer,
	repos *RepositoryContainer,
) *ControllerContainer {
	return &ControllerContainer{
		HealthController: controllers.NewHealthController(cfg, mongodb),
		UserController: controllers.NewUserController(
			cfg,
			mongodb.DB,
			services.UserService,
			services.ClientManager,
		),
		ChatController: controllers.NewChatController(
			cfg,
			mongodb.DB,
			services.ChatService,
			services.UserService,
		),
		ServerController: controllers.NewServerController(
			cfg,
			mongodb.DB,
			services.ServerService,
		),
		FriendController: controllers.NewFriendController(
			cfg,
			mongodb.DB,
			services.FriendService,
		),
		ChannelController: controllers.NewChannelController(
			cfg,
			mongodb.DB,
			services.ChannelService,
		),
		FileController: controllers.NewFileController(
			cfg,
			mongodb.DB,
			services.FileUploadService,
		),
	}
}

// 初始化Providers
func initProviders(
	cfg *config.Config,
	mongodb *providers.MongoWrapper,
	redis *providers.RedisWrapper,
) *ProviderContainer {
	var fileProvider providers.FileProvider
	if cfg.Server.Mode == config.ProductionMode {
		minioProv, err := providers.NewMinIOProvider(cfg)
		if err != nil {
			panic(fmt.Sprintf("初始化 MinIO 失敗: %v", err))
		}
		fileProvider = minioProv
	} else {
		fileProvider = providers.NewFileProvider(cfg)
	}

	return &ProviderContainer{
		ODM:          providers.NewODM(mongodb.DB),
		FileProvider: fileProvider,
		Cache:        providers.NewRedisCacheProvider(redis.Client),
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
func BuildDependencies(
	cfg *config.Config,
	mongodb *providers.MongoWrapper,
	redis *providers.RedisWrapper,
) *ApplicationContainer {
	providers := initProviders(cfg, mongodb, redis)
	repos := initRepositories(cfg, providers)
	services := initServices(cfg, providers, repos, redis)
	controllers := initControllers(cfg, mongodb, services, repos)

	return &ApplicationContainer{
		Controllers: controllers,
		Services:    services,
		Repos:       repos,
		Providers:   providers,
	}
}
