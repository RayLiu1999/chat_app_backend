package controllers

import (
	"log"
	"net/http"

	"chat_app_backend/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
)

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan models.Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer ws.Close()

		username := c.GetString("username")
		clients[ws] = username

		for {
			var msg models.Message
			err := ws.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(clients, ws)
				break
			}
			msg.Username = username
			broadcast <- msg
		}
	}
}

func HandleMessages(db *mongo.Database) {
	for {
		msg := <-broadcast
		msg.Save(db)

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
