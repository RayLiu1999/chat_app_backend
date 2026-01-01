package routes

import (
	"chat_app_backend/app/http/middlewares"
	"chat_app_backend/config"
	"chat_app_backend/di"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, controllers *di.ControllerContainer) {
	// 初始化 Prometheus 監控
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	// 使用 JSON 日誌中間件 (配合 Loki)
	r.Use(middlewares.JSONLoggerMiddleware())

	// 設定靜態文件服務
	// 使用絕對路徑，確保在任何環境下都可以正確訪問上傳的文件
	uploadsAbsPath := filepath.Join(".", "uploads")
	r.Static("/uploads", uploadsAbsPath)

	// 健康檢查 - 使用中介軟體保護
	// r.GET("/health", middlewares.HealthCheckAuth(cfg), controllers.HealthController.HealthCheck)
	r.GET("/health", controllers.HealthController.HealthCheck)
	r.GET("/health/proxy", middlewares.PublicHealthCheckAuth(cfg), controllers.HealthController.ProxyCheck)
	r.GET("/health/detailed", middlewares.HealthCheckAuth(cfg), controllers.HealthController.DetailedHealthCheck)

	// 驗證前端來源
	if cfg.Server.Mode == config.ProductionMode {
		r.Use(middlewares.VerifyOrigin(cfg.Server.AllowedOrigins))
	}

	// 設定 CORS 中介軟體
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.AllowedOrigins,                                                          // 允許的來源
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},                                           // 允許的方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-NAME", "X-CSRF-TOKEN"}, // 允許的標頭
		ExposeHeaders:    []string{"Content-Length"},                                                         // 允許暴露的標頭
		AllowCredentials: true,                                                                               // 是否允許憑證
		MaxAge:           12 * time.Hour,                                                                     // 預檢請求的緩存時間
	}))

	// 未認證的路由，只需要 CSRF 驗證
	public := r.Group("/")
	public.POST("/register", controllers.UserController.Register)
	public.POST("/login", controllers.UserController.Login)
	public.POST("/logout", middlewares.VerifyCSRFToken(), controllers.UserController.Logout)
	public.POST("/refresh_token", middlewares.VerifyCSRFToken(), controllers.UserController.RefreshToken)

	// 測試用 API
	if cfg.Server.Mode != config.ProductionMode {
		// 透過用戶名取得使用者資訊（不需要認證）
		public.GET("/test/user", controllers.UserController.GetUserByUsername)
	}

	// 需要認證的路由
	auth := r.Group("/")
	auth.Use(middlewares.Auth())

	authWithCSRF := auth.Group("/")
	authWithCSRF.Use(middlewares.VerifyCSRFToken())

	// WebSocket
	auth.GET("/ws", controllers.ChatController.HandleConnections)

	// user
	auth.GET("/user", controllers.UserController.GetUser)

	// User Profile APIs
	auth.GET("/user/profile", controllers.UserController.GetUserProfile)
	authWithCSRF.PUT("/user/profile", controllers.UserController.UpdateUserProfile)
	authWithCSRF.POST("/user/upload-image", controllers.UserController.UploadUserImage)
	authWithCSRF.DELETE("/user/avatar", controllers.UserController.DeleteUserAvatar)
	authWithCSRF.DELETE("/user/banner", controllers.UserController.DeleteUserBanner)
	authWithCSRF.PUT("/user/password", controllers.UserController.UpdateUserPassword)
	authWithCSRF.PUT("/user/deactivate", controllers.UserController.DeactivateAccount)
	authWithCSRF.DELETE("/user/delete", controllers.UserController.DeleteAccount)

	// auth.GET("/users/:id/online-status", controllers.UserController.CheckUserOnlineStatus) // 檢查特定用戶在線狀態

	// friend
	auth.GET("/friends", controllers.FriendController.GetFriendList)

	// 新增的 API
	auth.GET("/friends/pending", controllers.FriendController.GetPendingRequests) // 獲取待處理請求
	auth.GET("/friends/blocked", controllers.FriendController.GetBlockedUsers)    // 獲取封鎖列表

	// 好友請求管理
	authWithCSRF.POST("/friends/requests", controllers.FriendController.SendFriendRequest)                       // 發送好友請求
	authWithCSRF.PUT("/friends/requests/:request_id/accept", controllers.FriendController.AcceptFriendRequest)   // 接受請求
	authWithCSRF.PUT("/friends/requests/:request_id/decline", controllers.FriendController.DeclineFriendRequest) // 拒絕請求
	authWithCSRF.DELETE("/friends/requests/:request_id", controllers.FriendController.CancelFriendRequest)       // 取消請求

	// 封鎖管理
	authWithCSRF.POST("/friends/:user_id/block", controllers.FriendController.BlockUser)     // 封鎖用戶
	authWithCSRF.DELETE("/friends/:user_id/block", controllers.FriendController.UnblockUser) // 解除封鎖

	// 刪除好友
	authWithCSRF.DELETE("/friends/remove/:friend_id", controllers.FriendController.RemoveFriend) // 刪除好友

	// dm room
	auth.GET("/dm_rooms", controllers.ChatController.GetDMRoomList)                   // 獲取聊天列表
	authWithCSRF.PUT("/dm_rooms", controllers.ChatController.UpdateDMRoom)            // 更新聊天列表狀態
	authWithCSRF.POST("/dm_rooms", controllers.ChatController.CreateDMRoom)           // 保存聊天列表
	auth.GET("/dm_rooms/:room_id/messages", controllers.ChatController.GetDMMessages) // 獲取私聊訊息

	// server
	auth.GET("/servers", controllers.ServerController.GetServerList)
	authWithCSRF.POST("/servers", controllers.ServerController.CreateServer)
	auth.GET("/servers/search", controllers.ServerController.SearchPublicServers)            // 搜尋公開伺服器
	auth.GET("/servers/:server_id", controllers.ServerController.GetServerByID)              // 獲取單個伺服器信息
	auth.GET("/servers/:server_id/detail", controllers.ServerController.GetServerDetailByID) // 獲取伺服器詳細信息（含成員）
	authWithCSRF.PUT("/servers/:server_id", controllers.ServerController.UpdateServer)       // 更新伺服器信息
	authWithCSRF.DELETE("/servers/:server_id", controllers.ServerController.DeleteServer)    // 刪除伺服器
	authWithCSRF.POST("/servers/:server_id/join", controllers.ServerController.JoinServer)   // 請求加入伺服器
	authWithCSRF.POST("/servers/:server_id/leave", controllers.ServerController.LeaveServer) // 離開伺服器

	// channel
	auth.GET("/servers/:server_id/channels", controllers.ChannelController.GetChannelsByServerID)  // 獲取伺服器頻道列表
	auth.GET("/channels/:channel_id", controllers.ChannelController.GetChannelByID)                // 獲取單個頻道信息
	authWithCSRF.POST("/servers/:server_id/channels", controllers.ChannelController.CreateChannel) // 創建新頻道
	authWithCSRF.PUT("/channels/:channel_id", controllers.ChannelController.UpdateChannel)         // 更新頻道信息
	authWithCSRF.DELETE("/channels/:channel_id", controllers.ChannelController.DeleteChannel)      // 刪除頻道
	auth.GET("/channels/:channel_id/messages", controllers.ChatController.GetChannelMessages)      // 獲取頻道訊息

	// file upload
	authWithCSRF.POST("/upload/file", controllers.FileController.UploadFile)         // 通用檔案上傳
	authWithCSRF.POST("/upload/avatar", controllers.FileController.UploadAvatar)     // 頭像上傳
	authWithCSRF.POST("/upload/document", controllers.FileController.UploadDocument) // 文件上傳
	auth.GET("/files", controllers.FileController.GetUserFiles)                      // 獲取用戶檔案列表
	authWithCSRF.DELETE("/files/:file_id", controllers.FileController.DeleteFile)    // 刪除檔案
}
