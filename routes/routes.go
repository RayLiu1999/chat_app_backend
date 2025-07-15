package routes

import (
	"chat_app_backend/config"
	"chat_app_backend/di"
	"chat_app_backend/middlewares"
	"chat_app_backend/providers"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, mongodb *providers.MongoWrapper) {
	// 建立依賴
	controllers := di.BuildDependencies(cfg, mongodb)

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
	public.POST("/register", controllers.UserController.Register)
	public.POST("/login", controllers.UserController.Login)
	public.POST("/logout", controllers.UserController.Logout)
	public.POST("/refresh_token", controllers.UserController.RefreshToken)

	// 需要認證的路由
	auth := r.Group("/")
	auth.Use(middlewares.Auth())

	// CSRF 驗證
	auth.Use(middlewares.VerifyCsrfToken())

	// WebSocket
	auth.GET("/ws", controllers.ChatController.HandleConnections)

	// user
	auth.GET("/user", controllers.UserController.GetUser)

	// friend
	auth.GET("/friends", controllers.FriendController.GetFriendList)                 // 取得好友清單
	auth.POST("/friends", controllers.FriendController.AddFriendRequest)             // 建立好友請求
	auth.PUT("/friends/:friend_id", controllers.FriendController.UpdateFriendStatus) // 更新好友狀態
	// auth.DELETE("/friends/:friend_id", controllers.FriendController.RemoveFriend)     // 刪除好友

	// dm room
	auth.GET("/dm_rooms", controllers.ChatController.GetDMRoomList)                   // 獲取聊天列表
	auth.PUT("/dm_rooms", controllers.ChatController.UpdateDMRoom)                    // 更新聊天列表狀態
	auth.POST("/dm_rooms", controllers.ChatController.CreateDMRoom)                   // 保存聊天列表
	auth.GET("/dm_rooms/:room_id/messages", controllers.ChatController.GetDMMessages) // 獲取私聊訊息

	// server
	auth.GET("/servers", controllers.ServerController.GetServerList)
	auth.POST("/servers", controllers.ServerController.CreateServer)
	// auth.DELETE("/servers/:server_id", controllers.ServerController.DeleteServer)

	// channel
	// auth.GET("/channels/:server_id", controllers.ChatController.GetChannelList)
}
