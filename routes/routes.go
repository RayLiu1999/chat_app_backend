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

func SetupRoutes(r *gin.Engine, cfg *config.Config, mongodb *providers.MongoWrapper, controllers *di.ControllerContainer) {
	// 設定靜態文件服務
	// 使用絕對路徑，確保在任何環境下都可以正確訪問上傳的文件
	uploadsAbsPath := filepath.Join(".", "uploads")
	r.Static("/uploads", uploadsAbsPath)

	// 健康檢查 - 使用中介軟體保護
	r.GET("/health", middlewares.HealthCheckAuth(cfg), controllers.HealthController.HealthCheck)
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
	public.POST("/logout", controllers.UserController.Logout)
	public.POST("/refresh_token", controllers.UserController.RefreshToken)

	// 需要認證的路由
	auth := r.Group("/")
	auth.Use(middlewares.Auth())

	// WebSocket
	auth.GET("/ws", controllers.ChatController.HandleConnections)

	// user
	auth.GET("/user", controllers.UserController.GetUser)

	// User Profile APIs
	auth.GET("/user/profile", controllers.UserController.GetUserProfile)
	auth.PUT("/user/profile", controllers.UserController.UpdateUserProfile)
	auth.POST("/user/upload-image", controllers.UserController.UploadUserImage)
	auth.DELETE("/user/avatar", controllers.UserController.DeleteUserAvatar)
	auth.DELETE("/user/banner", controllers.UserController.DeleteUserBanner)
	auth.PUT("/user/password", controllers.UserController.UpdateUserPassword)
	// auth.GET("/user/two-factor-status", controllers.UserController.GetTwoFactorStatus)
	// auth.PUT("/user/two-factor", controllers.UserController.UpdateTwoFactorStatus)
	auth.PUT("/user/deactivate", controllers.UserController.DeactivateAccount)
	auth.DELETE("/user/delete", controllers.UserController.DeleteAccount)

	// auth.GET("/users/:id/online-status", controllers.UserController.CheckUserOnlineStatus) // 檢查特定用戶在線狀態

	// friend
	auth.GET("/friends", controllers.FriendController.GetFriendList)

	// 新增的 API
	auth.GET("/friends/pending", controllers.FriendController.GetPendingRequests) // 獲取待處理請求
	auth.GET("/friends/blocked", controllers.FriendController.GetBlockedUsers)    // 獲取封鎖列表

	// 好友請求管理
	auth.POST("/friends/requests", controllers.FriendController.SendFriendRequest)                       // 發送好友請求
	auth.PUT("/friends/requests/:request_id/accept", controllers.FriendController.AcceptFriendRequest)   // 接受請求
	auth.PUT("/friends/requests/:request_id/decline", controllers.FriendController.DeclineFriendRequest) // 拒絕請求
	auth.DELETE("/friends/requests/:request_id", controllers.FriendController.CancelFriendRequest)       // 取消請求

	// 封鎖管理
	auth.POST("/friends/:user_id/block", controllers.FriendController.BlockUser)     // 封鎖用戶
	auth.DELETE("/friends/:user_id/block", controllers.FriendController.UnblockUser) // 解除封鎖

	// 刪除好友
	auth.DELETE("/friends/remove/:friend_id", controllers.FriendController.RemoveFriend) // 刪除好友

	// dm room
	auth.GET("/dm_rooms", controllers.ChatController.GetDMRoomList)                   // 獲取聊天列表
	auth.PUT("/dm_rooms", controllers.ChatController.UpdateDMRoom)                    // 更新聊天列表狀態
	auth.POST("/dm_rooms", controllers.ChatController.CreateDMRoom)                   // 保存聊天列表
	auth.GET("/dm_rooms/:room_id/messages", controllers.ChatController.GetDMMessages) // 獲取私聊訊息

	// server
	auth.GET("/servers", controllers.ServerController.GetServerList)
	auth.POST("/servers", controllers.ServerController.CreateServer)
	auth.GET("/servers/search", controllers.ServerController.SearchPublicServers)            // 搜尋公開伺服器
	auth.GET("/servers/:server_id", controllers.ServerController.GetServerByID)              // 獲取單個伺服器信息
	auth.GET("/servers/:server_id/detail", controllers.ServerController.GetServerDetailByID) // 獲取伺服器詳細信息（含成員）
	auth.PUT("/servers/:server_id", controllers.ServerController.UpdateServer)               // 更新伺服器信息
	auth.DELETE("/servers/:server_id", controllers.ServerController.DeleteServer)            // 刪除伺服器
	auth.POST("/servers/:server_id/join", controllers.ServerController.JoinServer)           // 請求加入伺服器
	auth.POST("/servers/:server_id/leave", controllers.ServerController.LeaveServer)         // 離開伺服器

	// channel
	auth.GET("/servers/:server_id/channels", controllers.ChannelController.GetChannelsByServerID) // 獲取伺服器頻道列表
	auth.GET("/channels/:channel_id", controllers.ChannelController.GetChannelByID)               // 獲取單個頻道信息
	auth.POST("/servers/:server_id/channels", controllers.ChannelController.CreateChannel)        // 創建新頻道
	auth.PUT("/channels/:channel_id", controllers.ChannelController.UpdateChannel)                // 更新頻道信息
	auth.DELETE("/channels/:channel_id", controllers.ChannelController.DeleteChannel)             // 刪除頻道
	auth.GET("/channels/:channel_id/messages", controllers.ChatController.GetChannelMessages)     // 獲取頻道訊息

	// file upload
	auth.POST("/upload/file", controllers.FileController.UploadFile)         // 通用檔案上傳
	auth.POST("/upload/avatar", controllers.FileController.UploadAvatar)     // 頭像上傳
	auth.POST("/upload/document", controllers.FileController.UploadDocument) // 文件上傳
	auth.GET("/files", controllers.FileController.GetUserFiles)              // 獲取用戶檔案列表
	auth.DELETE("/files/:file_id", controllers.FileController.DeleteFile)    // 刪除檔案
}
