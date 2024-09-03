package main

import (
	"context"
	"log"

	"chat_app_backend/config"
	"chat_app_backend/providers"
	"chat_app_backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化數據庫連接
	db, err := providers.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Client().Disconnect(context.TODO())

	// 初始化 Gin
	r := gin.Default()

	// 設置路由
	routes.SetupRoutes(r)

	// 啟動服務器
	log.Println("Server starting on :" + config.Port)
	r.Run(":" + config.Port)
}
