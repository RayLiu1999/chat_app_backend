package routes

import (
	"chat_app_backend/app/http/middlewares"
	"chat_app_backend/app/providers"
	"chat_app_backend/config"
	"chat_app_backend/di"
	"chat_app_backend/version"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func SetupRoutes(r *gin.Engine, cfg *config.Config, redis *providers.RedisWrapper, controllers *di.ControllerContainer) {
	// 初始化 Prometheus 監控
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	// 使用 JSON 日誌中間件 (配合 Loki)
	r.Use(middlewares.JSONLoggerMiddleware())

	// 健康檢查
	r.GET("/health", controllers.HealthController.HealthCheck)
	r.GET("/health/proxy", middlewares.PublicHealthCheckAuth(cfg), controllers.HealthController.ProxyCheck)
	r.GET("/health/detailed", middlewares.HealthCheckAuth(cfg), controllers.HealthController.DetailedHealthCheck)

	// 版本資訊 API
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, version.GetInfo())
	})

	// --- 以下路由套用全域請求超時設定 (30秒) ---
	// WebSocket 連線不套用此 timeout（長連線由 Ping/Pong 管理）
	withTimeout := r.Group("/")
	withTimeout.Use(middlewares.Timeout(30 * time.Second))

	// 設定靜態文件服務 (此處也可以視需求決定是否套用 timeout，Static 通常不需要)
	uploadsAbsPath := filepath.Join(".", "uploads")
	r.Static("/uploads", uploadsAbsPath)

	// 驗證前端來源
	if cfg.Server.Mode == config.ProductionMode {
		withTimeout.Use(middlewares.VerifyOrigin(cfg.Server.AllowedOrigins))
	}

	// 設定 CORS 中介軟體
	withTimeout.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.Server.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-CSRF-NAME", "X-CSRF-TOKEN"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 未認證的路由
	public := withTimeout.Group("/")
	public.POST("/register",
		middlewares.RateLimiter(redis.Client, "register", 3, time.Minute),
		controllers.UserController.Register,
	)
	public.POST("/login",
		middlewares.RateLimiter(redis.Client, "login", 5, time.Minute),
		controllers.UserController.Login,
	)
	public.POST("/logout", middlewares.VerifyCSRFToken(), controllers.UserController.Logout)
	public.POST("/refresh_token",
		middlewares.RateLimiter(redis.Client, "refresh_token", 10, 5*time.Minute),
		middlewares.VerifyCSRFToken(),
		controllers.UserController.RefreshToken,
	)

	// 測試用 API
	if cfg.Server.Mode != config.ProductionMode {
		public.GET("/test/user", controllers.UserController.GetUserByUsername)
	}

	// 需要認證的路由
	auth := withTimeout.Group("/")
	auth.Use(middlewares.Auth())

	// WebSocket 特殊處理：需要認證，但不要 Timeout
	// 注意：這裡使用 auth.Group("/") 但排除 timeout 是比較困難的，
	// 所以我們建立一個獨立的 wsAuth 組，只包含 Auth 但不包含 Timeout。
	wsAuth := r.Group("/")
	wsAuth.Use(middlewares.Auth())
	wsAuth.GET("/ws", controllers.ChatController.HandleConnections)

	authWithCSRF := auth.Group("/")
	authWithCSRF.Use(middlewares.VerifyCSRFToken())

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
	// 上傳路由獨立群組，覆蓋全域的 30s timeout，改為 120s（大型檔案上傳需要更長時間）
	uploadGroup := authWithCSRF.Group("/")
	uploadGroup.Use(middlewares.Timeout(120 * time.Second))
	uploadGroup.POST("/upload/file", controllers.FileController.UploadFile)         // 通用檔案上傳
	uploadGroup.POST("/upload/avatar", controllers.FileController.UploadAvatar)     // 頭像上傳
	uploadGroup.POST("/upload/document", controllers.FileController.UploadDocument) // 文件上傳
	auth.GET("/files", controllers.FileController.GetUserFiles)                     // 獲取用戶檔案列表
	authWithCSRF.DELETE("/files/:file_id", controllers.FileController.DeleteFile)   // 刪除檔案
}
