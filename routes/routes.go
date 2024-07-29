package routes

import (
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	auth := r.Group("/")
	auth.Use(middlewares.Auth())
	{
		auth.GET("/ws", controllers.HandleWebSocket)
	}

	r.StaticFile("/", "./static/index.html")
}
