package routes

import (
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, db *mongo.Database) {
	// 创建 Controller 实例并传入数据库连接
	baseController := &controllers.BaseController{DB: db}

	r.POST("/register", baseController.Register)
	r.POST("/login", baseController.Login)

	auth := r.Group("/")
	auth.Use(middlewares.Auth())
	{
		auth.GET("/ws", controllers.HandleConnections(db))
	}
}
