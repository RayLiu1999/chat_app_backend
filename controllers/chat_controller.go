package controllers

import (
	"log"
	"net/http"

	"chat_app_backend/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]string)
var broadcast = make(chan models.Message)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 升級初始 HTTP 連接為 WebSocket
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer ws.Close()

		// 獲取用戶名並將其添加到 clients map 中
		username := c.Query("username")
		log.Printf("User %s connected", username)
		clients[ws] = username

		// 持續接收來自前端的訊息
		for {
			var msg models.Message
			err := ws.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(clients, ws)
				break
			}
			broadcast <- msg
		}
	}
}

func SendMessage(c *gin.Context) {
	var msg models.Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 將訊息發送到 broadcast 通道
	broadcast <- msg
}

func HandleMessages() {
	for {
		// 從 broadcast 通道中讀取訊息
		msg := <-broadcast
		log.Printf("Message received: %v", msg)
	}
}
