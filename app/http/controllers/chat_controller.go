package controllers

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 處理 WebSocket 連接
func (cc *ChatController) HandleConnections(c *gin.Context) {
	// 解析參數
	token := c.Query("token")
	utils.PrettyPrint("WebSocket token:", token)

	// 取得 userID
	userID, _, err := utils.GetUserFromToken(token)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrInvalidToken})
		return
	}

	// 升級 HTTP 連接為 WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, models.MessageOptions{Code: models.ErrInternalServer})
		return
	}

	utils.PrettyPrint("WebSocket connection established for user:", userID)
	// 使用聊天服務處理連接
	cc.chatService.HandleWebSocket(ws, userID)
}

// GetDMRoomList 獲取用戶的聊天列表
func (cc *ChatController) GetDMRoomList(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 獲取聊天列表
	dmRoomResponseList, msgOpt := cc.chatService.GetDMRoomResponseList(userID, false)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, dmRoomResponseList, "獲取聊天列表成功")
}

// UpdateDMRoom 更新聊天列表的狀態（標記為已刪除或取消刪除）
func (cc *ChatController) UpdateDMRoom(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	var requestBody struct {
		RoomID   string `json:"room_id"`
		IsHidden bool   `json:"is_hidden"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	// 使用service層的業務邏輯
	msgOpt := cc.chatService.UpdateDMRoom(userID, requestBody.RoomID, requestBody.IsHidden)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "聊天列表保存成功")
}

// 建立私聊房間
func (cc *ChatController) CreateDMRoom(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 從json取得chat_with_user_id
	var requestBody struct {
		ChatWithUserID string `json:"chat_with_user_id"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	// 使用service層的業務邏輯
	response, msgOpt := cc.chatService.CreateDMRoom(userID, requestBody.ChatWithUserID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, response, "聊天列表已創建")
}

// GetDMMessages 獲取私聊訊息
func (cc *ChatController) GetDMMessages(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	roomID := c.Param("room_id")
	before := c.Query("before")
	after := c.Query("after")
	limit := c.Query("limit")

	// 使用service層的業務邏輯
	messages, msgOpt := cc.chatService.GetDMMessages(userID, roomID, before, after, limit)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, messages, "獲取訊息成功")
}

// GetChannelMessages 獲取頻道訊息
func (cc *ChatController) GetChannelMessages(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	channelID := c.Param("channel_id")
	if channelID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "頻道ID不能為空",
		})
		return
	}

	before := c.Query("before")
	after := c.Query("after")
	limit := c.Query("limit")

	// 使用service層的業務邏輯
	messages, msgOpt := cc.chatService.GetChannelMessages(userID, channelID, before, after, limit)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, messages, "獲取頻道訊息成功")
}
