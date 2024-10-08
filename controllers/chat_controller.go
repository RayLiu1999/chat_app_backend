package controllers

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 定義消息結構
type Message struct {
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`   // 新增 UserID
	Username  string `json:"username"`  // 用戶名
	Text      string `json:"text"`      // 消息文本
	Timestamp int64  `json:"timestamp"` // 時間戳
}

// 定義用戶結構
type User struct {
	ID   string
	Conn *websocket.Conn
}

// 定義房間結構
type Room struct {
	ID        string
	Broadcast chan Message   // 用於廣播訊息的通道
	Clients   map[*User]bool // 將 Clients 變更為 User
	Lock      sync.Mutex
}

// 房間管理
var rooms = make(map[string]*Room)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (bc *BaseController) HandleConnections(c *gin.Context) {
	// 解析聊天室ID和用戶ID
	roomID := c.Query("room_id")
	userID := c.Query("user_id")

	// 升級初始 HTTP 連接為 WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// 创建用户
	user := &User{
		ID:   userID,
		Conn: ws,
	}

	// 加入房间
	room := joinRoom(roomID, user)

	defer func() {
		room.leaveRoom(user)
		ws.Close()
	}()

	// 監聽用戶發送的消息
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Println("Error reading json:", err)
			break
		}

		msg.UserID = user.ID

		room.Broadcast <- msg // 廣播消息到該房間的所有用戶
	}

	// 移除用戶
	room.Lock.Lock()
	delete(room.Clients, user)
	room.Lock.Unlock()
}

func (bc *BaseController) SendMessage(c *gin.Context) {
	var msg Message
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := rooms[msg.RoomID]

	// 將訊息發送到 broadcast 通道
	room.Broadcast <- msg
}

// joinRoom 让用户加入房间
func joinRoom(roomID string, user *User) *Room {
	var room *Room
	if r, ok := rooms[roomID]; ok {
		room = r
	} else {
		room = &Room{
			ID:        roomID,
			Broadcast: make(chan Message),
			Clients:   make(map[*User]bool),
		}
		rooms[roomID] = room
		go room.start()
	}

	room.Lock.Lock()
	room.Clients[user] = true
	room.Lock.Unlock()

	return room
}

// leaveRoom 让用户离开房间
func (room *Room) leaveRoom(user *User) {
	room.Lock.Lock()
	defer room.Lock.Unlock()
	delete(room.Clients, user)
}

// start 开始监听并广播消息到房间内的所有用户
func (room *Room) start() {
	for {
		msg := <-room.Broadcast
		room.Lock.Lock()
		for client := range room.Clients {
			err := client.Conn.WriteJSON(msg)
			if err != nil {
				fmt.Println("Error broadcasting message:", err)
				client.Conn.Close()
				delete(room.Clients, client)
			}
		}
		room.Lock.Unlock()
	}
}
