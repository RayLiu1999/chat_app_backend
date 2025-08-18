package main

import (
	"log"
	"os"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/di"
	"chat_app_backend/providers"
	"chat_app_backend/routes"

	"github.com/gin-gonic/gin"
)

func init() {
	// 載入環境變數配置
	config.LoadConfig()

	// 設置時區
	location, err := time.LoadLocation(config.AppConfig.Server.Timezone)
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}
	time.Local = location
}

func main() {
	// 初始化資料庫連接
	mongodb, err := providers.DBConnect[*providers.MongoWrapper]("mongodb")
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongodb.Close()

	gin.SetMode(config.AppConfig.Server.Mode)

	// 初始化 Gin
	r := gin.Default()

	// 構建依賴
	deps := di.BuildDependencies(config.AppConfig, mongodb)

	// // 使用依賴容器中的 UserService 來啟動後台任務
	// backgroundTasks := services.NewBackgroundTasks(deps.Services.UserService)
	// go backgroundTasks.StartAllBackgroundTasks()

	// 註冊 pprof
	// pprof.Register(r)

	// 設置路由
	routes.SetupRoutes(r, config.AppConfig, mongodb, deps.Controllers)

	// 確保上傳目錄存在
	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		log.Fatalf("無法創建上傳目錄: %v", err)
	}

	// 啟動服務器
	log.Println("Server starting on :" + config.AppConfig.Server.Port)
	r.Run(":" + config.AppConfig.Server.Port)
}
