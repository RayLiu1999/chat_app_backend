package main

import (
	"context"
	"log"

	"chat_app_backend/config"
	"chat_app_backend/database"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// 获取全局配置
	cfg := config.GetConfig()

	gin.SetMode(cfg.Server.Mode)

	// 初始化數據庫連接
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 插入一筆數據
	mongoDatabase, ok := db.(*mongo.Database)
	if !ok {
		log.Fatalf("Failed to cast db to *mongo.Client")
	}
	collection := mongoDatabase.Collection("users")
	_, err = collection.InsertOne(context.TODO(), bson.M{"name": "test"})
	if err != nil {
		log.Fatalf("Failed to insert document: %v", err)
	}

	// // 根据需要处理数据库连接对象 db
	// // db 可以是 *mongo.Client, *gorm.DB depending on the config
	// log.Println("Database connection initialized:", db)

	// // 初始化 Gin
	// r := gin.Default()

	// // 設置路由
	// routes.SetupRoutes(r)

	// // 啟動服務器
	// log.Println("Server starting on :" + cfg.Server.Port)
	// r.Run(":" + cfg.Server.Port)
}
