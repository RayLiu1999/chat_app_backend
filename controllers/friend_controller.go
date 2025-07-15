package controllers

import (
	"net/http"

	"chat_app_backend/config"
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
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized, Displayable: true})
		return
	}

	// 使用service層的業務邏輯
	friends, err := uc.friendService.GetFriendList(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, friends, utils.MessageOptions{Message: "好友資訊獲取成功"})
}

// 建立好友請求
func (uc *FriendController) AddFriendRequest(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	username := c.Param("username")

	// 使用service層的業務邏輯
	err = uc.friendService.AddFriendRequest(userID, username)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: err.Error()})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "好友請求已發送"})
}

// 更新好友狀態
func (uc *FriendController) UpdateFriendStatus(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	friendID := c.Param("friend_id")

	// 取得put中status資料
	var requestBody FriendRequestStatus
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 使用service層的業務邏輯
	err = uc.friendService.UpdateFriendStatus(userID, friendID, requestBody.Status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams, Message: err.Error()})
		return
	}

	var message string
	if requestBody.Status == FriendStatusAccepted {
		message = "已接受好友請求"
	} else {
		message = "已拒絕好友請求"
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: message})
}
