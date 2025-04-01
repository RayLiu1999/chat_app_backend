package routes

import (
	"chat_app_backend/config"
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"
	"chat_app_backend/repositories"
	"chat_app_backend/services"
	"path/filepath"
	"time"

	"chat_app_backend/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 初始化配置
	cfg := config.GetConfig()

	// 初始化資料庫連接
	mongodb := database.MongoDBConnect()

	// 獲取服務實例
	userService := services.NewUserService(cfg, mongodb, repositories.NewUserRepository(cfg, mongodb))
	chatService := services.NewChatService(cfg, mongodb, repositories.NewChatRepository(cfg, mongodb), repositories.NewServerRepository(cfg, mongodb))
	serverService := services.NewServerService(cfg, mongodb, repositories.NewServerRepository(cfg, mongodb))

	// 初始化控制器
	userController := controllers.NewUserController(cfg, mongodb, userService)
	chatController := controllers.NewChatController(cfg, mongodb, chatService, userService)
	serverController := controllers.NewServerController(cfg, mongodb, serverService, userService)

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

	// 將多個中介軟體組合成一個切片
	// middlewareArr := []gin.HandlerFunc{
	// 	middlewares.Auth(),
	// 	middlewares.VerifyCsrfToken(),
	// 	// 其他中介軟體
	// }

	csrf := r.Group("/")
	csrf.Use(middlewares.VerifyCsrfToken())
	csrf.POST("/register", userController.Register)
	csrf.POST("/login", userController.Login)
	csrf.POST("/logout", userController.Logout)
	csrf.POST("/refresh_token", userController.Refresh)

	auth := r.Group("/")
	auth.Use(middlewares.Auth())
	auth.GET("/ws", chatController.HandleConnections)
	auth.GET("/user", userController.GetUser)
	auth.GET("/servers", serverController.GetServerList)
	auth.POST("/servers", serverController.CreateServer)
	// auth.GET("/channels/:server_id", chatController.GetChannelList)
	// auth.GET("/messages/:room_id", chatController.GetMessages)

	// GET以外的請求需要驗證 CSRF Token
	auth.Use(middlewares.VerifyCsrfToken())
	// auth.POST("/message", baseController.SendMessage)
}
