package main

import (
	"log"

	"chat_app_backend/config"
	"chat_app_backend/database"
	"chat_app_backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 获取全局配置
	cfg := config.GetConfig()

	gin.SetMode(cfg.Server.Mode)

	// 初始化數據庫連接
	_, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 初始化 Gin
	r := gin.Default()

	// 設置路由
	routes.SetupRoutes(r)

	// 啟動服務器
	log.Println("Server starting on :" + cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
