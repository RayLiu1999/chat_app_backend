package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
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

// GetChatList 獲取用戶的聊天列表
func (cc *ChatController) GetChatList(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 解析是否包含已刪除的記錄
	includeDeletedStr := c.DefaultQuery("include_deleted", "false")
	includeDeleted := includeDeletedStr == "true"

	// 獲取聊天列表
	chatResponseList, err := cc.chatService.GetChatResponseList(objectID, includeDeleted)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: "獲取聊天列表失敗"})
		return
	}

	utils.SuccessResponse(c, chatResponseList, utils.MessageOptions{Message: "獲取聊天列表成功"})
}

// UpdateChat 更新聊天列表的狀態（標記為已刪除或取消刪除）
func (cc *ChatController) UpdateChat(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 從json取得chat_with_user_id
	var requestBody struct {
		UserID    primitive.ObjectID `json:"user_id"`
		IsDeleted bool               `json:"is_deleted"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 檢查user_id是否存在
	userCollection := cc.mongoConnect.Collection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": requestBody.UserID}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound, Message: "對話對象不存在"})
		return
	}

	var chat = models.Chat{
		UserID:         objectID,
		ChatWithUserID: requestBody.UserID,
		IsDeleted:      requestBody.IsDeleted,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存聊天列表
	chatResponse, err := cc.chatService.SaveChat(chat)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	utils.SuccessResponse(c, chatResponse, utils.MessageOptions{Message: "聊天列表保存成功"})
}

func (cc *ChatController) SaveChat(c *gin.Context) {
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 從json取得chat_with_user_id
	var requestBody struct {
		UserID primitive.ObjectID `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 檢查user_id是否存在
	userCollection := cc.mongoConnect.Collection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": requestBody.UserID}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound, Message: "對話對象不存在"})
		return
	}

	var chat = models.Chat{
		UserID:         objectID,
		ChatWithUserID: requestBody.UserID,
		IsDeleted:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 保存聊天列表
	chatResponse, err := cc.chatService.SaveChat(chat)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	utils.SuccessResponse(c, chatResponse, utils.MessageOptions{Message: "聊天列表保存成功"})
}
