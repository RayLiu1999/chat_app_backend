package controllers

import (
	"chat_app_backend/models"
	"chat_app_backend/services"
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 定義用戶結構
type User struct {
	ID     string
	Conn   *websocket.Conn
	Status string // `online` 或 `offline`
}

// 各伺服器包含的用戶列表(用於判斷用戶發送訊息後，該訊息要發送到哪些用戶)
type Server struct {
	mu        sync.Mutex
	ID        string
	Broadcast chan Message
	Users     map[string]User
}

// 定義私聊房間結構(用於1對1聊天)
type DMRoom struct {
	mu        sync.Mutex
	ID        string
	Broadcast chan Message // 用於廣播訊息的通道
	Users     map[string]User
}

// 定義消息結構
type Message struct {
	Type      string `json:"type"` // 可選，消息類型，如 text, image, video, audio, file 等
	RoomID    string `json:"room_id"`
	ServerID  string `json:"server_id"` // 可選，在伺服器時，server_id 用於識別伺服器
	UserID    string `json:"user_id"`   // 新增 UserID
	Text      string `json:"text"`      // 消息文本
	Timestamp int64  `json:"timestamp"` // 時間戳
}

// 伺服器管理
var servers = make(map[string]*Server)

// 房間管理
var rooms = make(map[string]*DMRoom)

// 定義 WebSocket 升級器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (bc *BaseController) HandleConnections(c *gin.Context) {
	// 解析聊天室ID和用戶ID
	roomID := c.Query("room_id")
	userID := c.Query("user_id")

	log.Println("User ID:", userID)
	log.Println("Room ID:", roomID)

	// 升級初始 HTTP 連接為 WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// 创建用户
	user := &User{
		ID:     userID,
		Conn:   ws,
		Status: "online",
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

		// 發送消息
		SendMessage(msg)
	}

	// 移除用戶
	// room.Lock.Lock()
	// delete(room.Clients, user)
	// room.Lock.Unlock()
}

// func (bc *BaseController) SendMessage(c *gin.Context) {
func SendMessage(msg Message) {
	// 判斷為伺服器發話or私聊
	if msg.Type == "server" {
		// 伺服器發話
		// sendToServer(msg)
	} else if msg.Type == "dm" {
		// 私聊
		sendToDM(msg)
	}

	// room := rooms[msg.RoomID]

	// // 將訊息發送到 broadcast 通道
	// room.Broadcast <- msg
}

func sendToServer(msg Message) {
	// 取得伺服器
	server := servers[msg.ServerID]

	// 發送訊息到伺服器的所有用戶
	server.Broadcast <- msg

	// 發送訊息到伺服器的所有用戶
	// for _, user := range server.Users {
	// 	err := user.Conn.WriteJSON(msg)
	// 	if err != nil {
	// 		log.Println("Error sending message to server:", err)
	// 		user.Conn.Close()
	// 	}
	// }
}

func sendToDM(msg Message) {
	// 取得私聊房間
	room := rooms[msg.RoomID]

	// 發送訊息到私聊房間的所有用戶
	// room.Broadcast <- msg
	room.mu.Lock()
	defer room.mu.Unlock()
	for _, user := range room.Users {
		err := user.Conn.WriteJSON(msg)
		if err != nil {
			log.Printf("Error sending message to user %s: %v", user.ID, err)
			user.Conn.Close()
			delete(room.Users, user.ID)
		}
	}
}

// joinRoom 让用户加入房间
func joinRoom(roomID string, user *User) *DMRoom {
	var room *DMRoom
	if r, ok := rooms[roomID]; ok {
		room = r
	} else {
		room = &DMRoom{
			ID:        roomID,
			Broadcast: make(chan Message),
			Users:     make(map[string]User),
		}

		rooms[roomID] = room
		// go room.start()
	}

	room.mu.Lock()
	room.Users[user.ID] = *user
	room.mu.Unlock()

	return room
}

// leaveRoom 让用户离开房间
func (room *DMRoom) leaveRoom(user *User) {
	// room.Lock.Lock()
	// defer room.Lock.Unlock()
	// delete(room.Clients, user)
}

// start 开始监听并广播消息到房间内的所有用户
func (room *DMRoom) start() {
	for {
		msg := <-room.Broadcast
		room.mu.Lock()
		for _, user := range room.Users {
			err := user.Conn.WriteJSON(msg)
			if err != nil {
				log.Println("Error broadcasting message:", err)
				user.Conn.Close()
				delete(room.Users, user.ID)
			}
		}
		room.mu.Unlock()
	}
}

// 取得伺服器列表
func (bc *BaseController) GetServerList(c *gin.Context) {
	// 取得使用者ID
	_, objectID, err := services.GetUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	_, err = bc.service.GetUserById(objectID)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse{Error: "User not found"})
		return
	}

	// var members = []models.Member{
	// 	{
	// 		UserID: objectID,
	// 	},
	// }

	// // 新建測試伺服器
	// server := &models.Server{
	// 	ID:          primitive.NewObjectID(),
	// 	Name:        "server2",
	// 	Picture:     "https://via.placeholder.com/150",
	// 	Description: "This is a test server",
	// 	OwnerID:     objectID,
	// 	Channels:    []primitive.ObjectID{},
	// 	Members:     members,
	// 	CreatedAt:   time.Now(),
	// 	UpdateAt:    time.Now(),
	// }

	// _, err = bc.service.AddServer(server)
	// if err != nil {
	// 	log.Println(err)
	// }

	servers, err := bc.service.GetServerListByUserId(objectID)
	log.Println("Servers:", servers)
	if err != nil {
		log.Println(err)
	}
	log.Println(servers)
	c.JSON(http.StatusOK, servers)
	return

	_, objectID, err = services.GetUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	// 先把資料加到server和user_server表中
	// serverCollection := bc.MongoConnect.Collection("servers")
	// userServerCollection := bc.MongoConnect.Collection("user_server")

	// var server models.Server
	// server.ID = primitive.NewObjectID()
	// server.Name = "server1"
	// server.Picture = "https://via.placeholder.com/150"
	// _, err = serverCollection.InsertOne(context.Background(), server)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting server"})
	// 	return
	// }

	// var userServer models.UserServer
	// userServer.ID = primitive.NewObjectID()
	// userServer.UserID = objectID
	// userServer.ServerID = server.ID
	// _, err = userServerCollection.InsertOne(context.Background(), userServer)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting user server"})
	// 	return
	// }

	// 取得伺服器ID清單
	userServerCollection := bc.MongoConnect.Collection("user_server")
	var userServerList []models.UserServer
	cursor, err := userServerCollection.Find(context.Background(), bson.M{"user_id": objectID})
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Error fetching user servers"})
		return
	}
	if err := cursor.All(context.Background(), &userServerList); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Error decoding user servers"})
		return
	}

	// 取得伺服器列表
	serverCollection := bc.MongoConnect.Collection("servers")

	// 提取所有的 server_id
	var serverIDs []primitive.ObjectID
	for _, userServer := range userServerList {
		serverIDs = append(serverIDs, userServer.ServerID)
	}

	// 使用 $in 運算符一次性查詢所有的伺服器
	filter := bson.M{"_id": bson.M{"$in": serverIDs}}
	cursor, err = serverCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error fetching servers"})
		return
	}
	defer cursor.Close(context.Background())

	var serverList []models.Server
	if err = cursor.All(context.Background(), &serverList); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Error decoding servers"})
		return
	}

	c.JSON(http.StatusOK, serverList)
}
