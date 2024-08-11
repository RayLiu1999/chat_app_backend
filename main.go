package main

import (
	"log"

	"chat_app_backend/config"
	"chat_app_backend/routes"
	"chat_app_backend/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化數據庫連接
	db, err := services.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Client().Disconnect(db.Context())

	// 初始化 Gin
	r := gin.Default()

	// 設置路由
	routes.SetupRoutes(r, db)

	// 啟動服務器
	log.Println("Server starting on :" + config.Port)
	r.Run(":" + config.Port)
}
