package routes

import (
	"chat_app_backend/controllers"
	"chat_app_backend/middlewares"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, db *mongo.Database) {
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)

	auth := r.Group("/")
	auth.Use(middlewares.Auth())
	{
		auth.GET("/ws", controllers.HandleConnections(db))
	}

	r.StaticFile("/", "./static/index.html")
}
