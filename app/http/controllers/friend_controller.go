package controllers

import (
	"net/http"

	"chat_app_backend/app/models"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"chat_app_backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type FriendController struct {
	config        *config.Config
	mongoConnect  *mongo.Database
	friendService services.FriendService
}

func NewFriendController(cfg *config.Config, mongodb *mongo.Database, friendService services.FriendService) *FriendController {
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

// GetPendingRequests 獲取待處理好友請求
func (fc *FriendController) GetPendingRequests(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	pendingRequests, msgOpt := fc.friendService.GetPendingRequests(userID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, pendingRequests, "待處理請求獲取成功")
}

// GetBlockedUsers 獲取封鎖用戶列表
func (fc *FriendController) GetBlockedUsers(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	blockedUsers, msgOpt := fc.friendService.GetBlockedUsers(userID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, blockedUsers, "封鎖列表獲取成功")
}

// SendFriendRequest 發送好友請求
func (fc *FriendController) SendFriendRequest(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	var request struct {
		Username string `json:"username" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.AddFriendRequest(userID, request.Username)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "好友請求已發送")
}

// AcceptFriendRequest 接受好友請求
func (fc *FriendController) AcceptFriendRequest(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.AcceptFriendRequest(userID, requestID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "已接受好友請求")
}

// DeclineFriendRequest 拒絕好友請求
func (fc *FriendController) DeclineFriendRequest(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.DeclineFriendRequest(userID, requestID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "已拒絕好友請求")
}

// CancelFriendRequest 取消好友請求
func (fc *FriendController) CancelFriendRequest(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	requestID := c.Param("request_id")
	if requestID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.CancelFriendRequest(userID, requestID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "已取消好友請求")
}

// BlockUser 封鎖用戶
func (fc *FriendController) BlockUser(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.BlockUser(userID, targetUserID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "已封鎖用戶")
}

// UnblockUser 解除封鎖用戶
func (fc *FriendController) UnblockUser(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	targetUserID := c.Param("user_id")
	if targetUserID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.UnblockUser(userID, targetUserID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "已解除封鎖")
}

// RemoveFriend 刪除好友
func (fc *FriendController) RemoveFriend(c *gin.Context) {
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	friendID := c.Param("friend_id")
	if friendID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{Code: models.ErrInvalidParams})
		return
	}

	msgOpt := fc.friendService.RemoveFriend(userID, friendID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "已刪除好友")
}
