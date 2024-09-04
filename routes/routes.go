package routes

import (
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"

	"chat_app_backend/providers"

	"github.com/gin-gonic/gin"
)

var MongoConnect = providers.DBConnect()

func SetupRoutes(r *gin.Engine) {
	// 创建 Controller 实例并传入数据库连接
	baseController := &controllers.BaseController{}

	r.POST("/register", baseController.Register)
	r.POST("/login", baseController.Login)

	auth := r.Group("/")
	auth.Use(middlewares.Auth())
	{
		auth.GET("/ws", controllers.HandleConnections())
	}
}
