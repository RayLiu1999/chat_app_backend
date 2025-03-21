package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 定義專門的控制器結構體
type ChatController struct {
	config       *config.Config
	mongoConnect *mongo.Database
	chatService  services.ChatServiceInterface
	userService  services.UserServiceInterface
}

// 創建控制器的工廠函數
func NewChatController(cfg *config.Config, mongodb *mongo.Database, chatService services.ChatServiceInterface, userService services.UserServiceInterface) *ChatController {
	return &ChatController{
		config:       cfg,
		mongoConnect: mongodb,
		chatService:  chatService,
		userService:  userService,
	}
}

// 定義用戶結構
type User struct {
	ID     string
	Conn   *websocket.Conn
	Status string // `online` 或 `offline`
}

// 定義 WebSocket 升級器
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 處理 WebSocket 連接
func (cc *ChatController) HandleConnections(c *gin.Context) {
	// 解析參數
	userID := c.Query("user_id")
	// roomID := c.Query("room_id")
	// serverID := c.Query("server_id")
	// username := c.Query("username")

	// 轉換 userID 為 ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		return
	}

	// 升級 HTTP 連接為 WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// 使用聊天服務處理連接
	cc.chatService.HandleConnection(objectID, ws)
}

// 取得伺服器列表
func (cc *ChatController) GetServerList(c *gin.Context) {
	// 取得使用者ID
	_, objectID, err := services.GetUserIDFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	_, err = cc.userService.GetUserById(objectID)
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

	servers, err := cc.chatService.GetServerListByUserId(objectID)
	// log.Println("Servers:", servers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	utils.SuccessResponse(c, servers, "Success")
}
