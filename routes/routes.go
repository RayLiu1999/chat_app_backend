package routes

import (
	"chat_app_backend/config"
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"
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

	// 初始化控制器
	baseController := controllers.NewBaseController(cfg, mongodb)

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
	csrf.POST("/register", baseController.Register)
	csrf.POST("/login", baseController.Login)
	csrf.POST("/refresh_token", baseController.Refresh)

	auth := r.Group("/")
	auth.Use(middlewares.Auth())
	auth.GET("/ws", baseController.HandleConnections)
	auth.GET("/user", baseController.GetUser)
	auth.GET("/servers", baseController.GetServerList)

	// GET以外的請求需要驗證 CSRF Token
	auth.Use(middlewares.VerifyCsrfToken())
	auth.POST("/message", baseController.SendMessage)
}
