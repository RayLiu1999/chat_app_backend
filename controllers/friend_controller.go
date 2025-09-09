package controllers

import (
	"net/http"

	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendController struct {
	config        *config.Config
	mongoConnect  *mongo.Database
	friendService services.FriendServiceInterface
}

func NewFriendController(cfg *config.Config, mongodb *mongo.Database, friendService services.FriendServiceInterface) *FriendController {
	return &FriendController{
		config:        cfg,
		mongoConnect:  mongodb,
		friendService: friendService,
	}
}

// 更新好友請求結構
type FriendRequestStatus struct {
	Status string `json:"status" binding:"required,oneof=pending accepted rejected"`
}

// 定義好友請求狀態的可能值
const (
	FriendStatusPending  = "pending"
	FriendStatusAccepted = "accepted"
	FriendStatusRejected = "rejected"
)

// 取得用戶資訊
func (uc *FriendController) GetFriendList(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 使用service層的業務邏輯
	friends, msgOpt := uc.friendService.GetFriendList(userID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, friends, "好友列表獲取成功")
}

// 建立好友請求
func (uc *FriendController) AddFriendRequest(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 綁定請求參數
	var friendRequest models.FriendRequest
	if err := c.ShouldBindJSON(&friendRequest); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	username := friendRequest.Username

	// 使用service層的業務邏輯
	res := uc.friendService.AddFriendRequest(userID, username)
	if res.Code != "" {
		ErrorResponse(c, http.StatusBadRequest, *res)
		return
	}

	SuccessResponse(c, nil, "好友請求已發送")
}

// 更新好友狀態
func (uc *FriendController) UpdateFriendStatus(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	friendID := c.Param("friend_id")

	// 取得put中status資料
	var requestBody FriendRequestStatus
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	// 使用service層的業務邏輯
	msgOpt := uc.friendService.UpdateFriendStatus(userID, friendID, requestBody.Status)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	var message string
	if requestBody.Status == FriendStatusAccepted {
		message = "已接受好友請求"
	} else {
		message = "已拒絕好友請求"
	}

	SuccessResponse(c, nil, message)
}
