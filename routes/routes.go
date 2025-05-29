package routes

import (
	"chat_app_backend/config"
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/services"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, mongodb *providers.MongoWrapper) {
	// 獲取服務實例
	userService := services.NewUserService(cfg, mongodb.DB, repositories.NewUserRepository(cfg, mongodb.DB))
	chatService := services.NewChatService(cfg, mongodb.DB, repositories.NewChatRepository(cfg, mongodb.DB), repositories.NewServerRepository(cfg, mongodb.DB), repositories.NewUserRepository(cfg, mongodb.DB))
	serverService := services.NewServerService(cfg, mongodb.DB, repositories.NewServerRepository(cfg, mongodb.DB))
	friendService := services.NewFriendService(cfg, mongodb.DB, repositories.NewFriendRepository(cfg, mongodb.DB))

	// 初始化控制器
	userController := controllers.NewUserController(cfg, mongodb.DB, userService)
	chatController := controllers.NewChatController(cfg, mongodb.DB, chatService, userService)
	serverController := controllers.NewServerController(cfg, mongodb.DB, serverService, userService)
	friendController := controllers.NewFriendController(cfg, mongodb.DB, friendService)

	// 設定靜態文件服務
	// 使用絕對路徑，確保在任何環境下都可以正確訪問上傳的文件
	uploadsAbsPath := filepath.Join(".", "uploads")
	r.Static("/uploads", uploadsAbsPath)

	// 設定 CORS 中介軟體
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,                                                                 // 允許的來源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},                                           // 允許的方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-NAME", "X-CSRF-TOKEN"}, // 允許的標頭
		ExposeHeaders:    []string{"Content-Length"},                                                         // 允許暴露的標頭
		AllowCredentials: true,                                                                               // 是否允許憑證
		MaxAge:           12 * time.Hour,                                                                     // 預檢請求的緩存時間
	}))

	// 未認證的路由，只需要 CSRF 驗證
	public := r.Group("/")
	public.Use(middlewares.VerifyCsrfToken())
	public.POST("/register", userController.Register)
	public.POST("/login", userController.Login)
	public.POST("/logout", userController.Logout)
	public.POST("/refresh_token", userController.Refresh)

	// 需要認證的路由
	auth := r.Group("/")
	auth.Use(middlewares.Auth())

	// CSRF 驗證
	auth.Use(middlewares.VerifyCsrfToken())

	// chat
	auth.GET("/ws", chatController.HandleConnections)
	auth.GET("/chats", chatController.GetChatList) // 獲取聊天列表
	auth.PUT("/chats", chatController.UpdateChat)  // 更新聊天列表狀態
	auth.POST("/chats", chatController.SaveChat)   // 保存聊天列表

	// user
	auth.GET("/user", userController.GetUser)

	// friend
	auth.GET("/friends", friendController.GetFriendList)                 // 取得好友清單
	auth.POST("/friends", friendController.AddFriendRequest)             // 建立好友請求
	auth.PUT("/friends/:friend_id", friendController.UpdateFriendStatus) // 更新好友狀態
	// auth.DELETE("/friends/:friend_id", friendController.RemoveFriend)     // 刪除好友

	// server
	auth.GET("/servers", serverController.GetServerList)
	auth.POST("/servers", serverController.CreateServer)
	// auth.DELETE("/servers/:server_id", serverController.DeleteServer)

	// channel
	// auth.GET("/channels/:server_id", chatController.GetChannelList)
	// auth.GET("/messages/:room_id", chatController.GetMessages)
}
