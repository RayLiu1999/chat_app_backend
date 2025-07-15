package main

import (
	"log"
	"os"
	"time"

	"chat_app_backend/config"
	"chat_app_backend/providers"
	"chat_app_backend/routes"

	"github.com/gin-gonic/gin"
)

var cfg *config.Config

func init() {
	// 取得全局配置
	cfg = config.GetConfig()

	// 設置時區
	location, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}
	time.Local = location
}

func main() {
	// 初始化資料庫連接
	mongodb, err := providers.DBConnect[*providers.MongoWrapper]("mongodb")
	if err != nil {
		mongodb.Close()
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	gin.SetMode(cfg.Server.Mode)

	// 初始化 Gin
	r := gin.Default()

	// 設置路由
	routes.SetupRoutes(r, cfg, mongodb)

	// 確保上傳目錄存在
	err = os.MkdirAll("uploads", 0755)
	if err != nil {
		log.Fatalf("無法創建上傳目錄: %v", err)
	}

	// 啟動服務器
	log.Println("Server starting on :" + cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
