package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	// 取得 userID
	userID, _, err := utils.GetUserFromToken(token)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrInvalidToken})
		return
	}

	// 升級 HTTP 連接為 WebSocket
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	// 使用聊天服務處理連接
	cc.chatService.HandleWebSocket(ws, userID)
}

// GetDMRoomList 獲取用戶的聊天列表
func (cc *ChatController) GetDMRoomList(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取聊天列表
	dmRoomResponseList, err := cc.chatService.GetDMRoomResponseList(userID, false)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: "獲取聊天列表失敗"})
		return
	}

	utils.SuccessResponse(c, dmRoomResponseList, utils.MessageOptions{Message: "獲取聊天列表成功"})
}

// UpdateDMRoom 更新聊天列表的狀態（標記為已刪除或取消刪除）
func (cc *ChatController) UpdateDMRoom(c *gin.Context) {
	_, userObjectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	var requestBody struct {
		RoomID   string `json:"room_id"`
		IsHidden bool   `json:"is_hidden"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	DMRoomObjectID, err := primitive.ObjectIDFromHex(requestBody.RoomID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 檢查room_id是否存在
	dmRoomCollection := cc.mongoConnect.Collection("dm_rooms")
	var dmRoom models.DMRoom
	err = dmRoomCollection.FindOne(context.Background(), bson.M{"room_id": DMRoomObjectID, "user_id": userObjectID}).Decode(&dmRoom)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrRoomNotFound})
		return
	}

	// 更新狀態
	dmRoomCollection.UpdateOne(context.Background(), bson.M{"room_id": DMRoomObjectID, "user_id": userObjectID}, bson.M{"$set": bson.M{"is_hidden": requestBody.IsHidden}})

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "聊天列表保存成功"})
}

// 建立私聊房間
func (cc *ChatController) CreateDMRoom(c *gin.Context) {
	_, userObjectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 從json取得chat_with_user_id
	var requestBody struct {
		ChatWithUserID string `json:"chat_with_user_id"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	ChatWithUserObjectID, err := primitive.ObjectIDFromHex(requestBody.ChatWithUserID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 檢查chat_with_user_id是否存在
	userCollection := cc.mongoConnect.Collection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": ChatWithUserObjectID}).Decode(&user)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound, Message: "對話對象不存在"})
		return
	}

	// 檢查房間是否存在
	roomCollection := cc.mongoConnect.Collection("dm_rooms")
	var roomList []models.DMRoom
	cursor, err := roomCollection.Find(context.Background(), bson.M{"$or": []bson.M{
		bson.M{
			"user_id":           ChatWithUserObjectID,
			"chat_with_user_id": userObjectID,
		},
		bson.M{
			"user_id":           userObjectID,
			"chat_with_user_id": ChatWithUserObjectID,
		},
	}})

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	cursor.All(context.Background(), &roomList)

	// 定義回傳格式
	var DMRoomResponse models.DMRoomResponse

	// 如果雙方都建立則直接回傳
	if len(roomList) == 2 {
		for _, room := range roomList {
			if room.UserID == userObjectID {
				DMRoomResponse = models.DMRoomResponse{
					RoomID:    room.RoomID,
					Nickname:  user.Nickname,
					Picture:   user.Picture,
					Timestamp: room.UpdatedAt.Unix(),
				}

				// 如果isHidden為true，則將isHidden設為false
				if room.IsHidden {
					roomCollection.UpdateOne(context.Background(), bson.M{"_id": room.ID}, bson.M{"$set": bson.M{"is_hidden": false}})
				}
			}
		}

		utils.SuccessResponse(c, DMRoomResponse, utils.MessageOptions{Message: "聊天列表已存在"})
		return
	}

	// 如果只有一邊建立過
	if len(roomList) == 1 {
		for _, room := range roomList {
			// 如果對方建立過，自己沒建過，則取得對方RoomID，並建立user_id為自己的房間
			if room.ChatWithUserID == userObjectID {
				// 取得對方RoomID
				roomID := room.RoomID

				// 建立user_id為自己的房間
				var dmRoom = models.DMRoom{
					RoomID:         roomID,
					UserID:         userObjectID,
					ChatWithUserID: ChatWithUserObjectID,
					IsHidden:       false,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}

				_, err := roomCollection.InsertOne(context.Background(), dmRoom)
				if err != nil {
					utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
					return
				}

				chatResponse := models.DMRoomResponse{
					RoomID:    roomID,
					Nickname:  user.Nickname,
					Picture:   user.Picture,
					Timestamp: dmRoom.UpdatedAt.Unix(),
				}

				utils.SuccessResponse(c, chatResponse, utils.MessageOptions{Message: "聊天列表保存成功"})
				return
			}

			// 如果自己建立過則直接回傳
			if room.UserID == userObjectID {
				chatResponse := models.DMRoomResponse{
					RoomID:    room.RoomID,
					Nickname:  user.Nickname,
					Picture:   user.Picture,
					Timestamp: room.UpdatedAt.Unix(),
				}

				utils.SuccessResponse(c, chatResponse, utils.MessageOptions{Message: "聊天列表已存在"})
				return
			}
		}
	}

	// 如果雙方都沒有建立過
	if len(roomList) == 0 {
		// 建立房間
		var dmRoom = models.DMRoom{
			RoomID:         primitive.NewObjectID(),
			UserID:         userObjectID,
			ChatWithUserID: ChatWithUserObjectID,
			IsHidden:       false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		_, err := roomCollection.InsertOne(context.Background(), dmRoom)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
			return
		}

		chatResponse := models.DMRoomResponse{
			RoomID:    dmRoom.RoomID,
			Nickname:  user.Nickname,
			Picture:   user.Picture,
			Timestamp: dmRoom.UpdatedAt.Unix(),
		}

		utils.SuccessResponse(c, chatResponse, utils.MessageOptions{Message: "聊天列表保存成功"})
		return
	}

	utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
}

// GetDMMessages 獲取私聊訊息
func (cc *ChatController) GetDMMessages(c *gin.Context) {
	_, userObjectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	RoomID := c.Param("room_id")
	Before := c.Query("before")
	After := c.Query("after")
	Limit := c.Query("limit")

	RoomObjectID, err := primitive.ObjectIDFromHex(RoomID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 檢查room_id是否存在
	roomCollection := cc.mongoConnect.Collection("dm_rooms")
	var room models.DMRoom
	err = roomCollection.FindOne(context.Background(), bson.M{"room_id": RoomObjectID, "user_id": userObjectID}).Decode(&room)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrRoomNotFound})
		return
	}

	// 獲取訊息
	messageCollection := cc.mongoConnect.Collection("messages")
	var messageList []models.Message
	filter := bson.M{"room_id": RoomObjectID}
	if Before != "" {
		BeforeObjectID, err := primitive.ObjectIDFromHex(Before)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
			return
		}
		filter["_id"] = bson.M{"$lt": BeforeObjectID}
	}
	if After != "" {
		AfterObjectID, err := primitive.ObjectIDFromHex(After)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
			return
		}
		filter["_id"] = bson.M{"$gt": AfterObjectID}
	}

	utils.PrettyPrint("filter", filter)

	opts := options.Find().SetSort(bson.D{{"_id", -1}})

	if Limit != "" {
		limitVal, err := strconv.ParseInt(Limit, 10, 64)
		if err == nil && limitVal > 0 {
			opts.SetLimit(limitVal)
		}
	}

	cursor, err := messageCollection.Find(context.Background(), filter, opts)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	cursor.All(context.Background(), &messageList)

	var messageResponse []models.MessageResponse

	for _, message := range messageList {
		messageResponse = append(messageResponse, models.MessageResponse{
			ID:        message.ID,
			RoomType:  models.RoomType(message.RoomType),
			RoomID:    message.RoomID.Hex(),
			SenderID:  message.SenderID.Hex(),
			Content:   message.Content,
			Timestamp: message.UpdatedAt.UnixMilli(),
		})
	}

	utils.SuccessResponse(c, messageResponse, utils.MessageOptions{Message: "獲取訊息成功"})
}
