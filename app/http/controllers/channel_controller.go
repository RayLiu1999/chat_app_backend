package controllers

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"chat_app_backend/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ChannelController struct {
	config         *config.Config
	mongoConnect   *mongo.Database
	channelService services.ChannelServiceInterface
}

func NewChannelController(cfg *config.Config, mongodb *mongo.Database, channelService services.ChannelServiceInterface) *ChannelController {
	return &ChannelController{
		config:         cfg,
		mongoConnect:   mongodb,
		channelService: channelService,
	}
}

// GetChannelsByServerID 獲取伺服器的頻道列表
func (cc *ChannelController) GetChannelsByServerID(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized, Message: "需要認證"})
		return
	}

	// 獲取伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 獲取頻道列表
	channels, msgOpt := cc.channelService.GetChannelsByServerID(userID, serverID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, channels, "獲取頻道列表成功")
}

// GetChannelByID 獲取單個頻道信息
func (cc *ChannelController) GetChannelByID(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized, Message: "需要認證"})
		return
	}

	// 獲取頻道ID
	channelID := c.Param("channel_id")
	if channelID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "頻道ID不能為空",
		})
		return
	}

	// 驗證頻道ID格式
	_, err = primitive.ObjectIDFromHex(channelID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的頻道ID格式",
		})
		return
	}

	// 獲取頻道信息
	channel, msgOpt := cc.channelService.GetChannelByID(userID, channelID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, channel, "獲取頻道信息成功")
}

// CreateChannel 創建新頻道
func (cc *ChannelController) CreateChannel(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 獲取伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 驗證伺服器ID格式
	_, err = primitive.ObjectIDFromHex(serverID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的伺服器ID格式",
		})
		return
	}

	// 解析請求體
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的請求格式",
			Details: err.Error(),
		})
		return
	}

	// 驗證請求數據
	if req.Name == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "頻道名稱不能為空",
		})
		return
	}

	if req.Type == "" {
		req.Type = "text" // 默認為文字頻道
	}

	// 轉換伺服器ID
	serverObjID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的伺服器ID",
			Details: err.Error(),
		})
		return
	}

	// 創建頻道模型
	channel := &models.Channel{
		Name:     req.Name,
		ServerID: serverObjID,
		Type:     req.Type,
	}

	// 如果提供了分類ID，則設置分類ID
	if req.CategoryID != "" {
		categoryObjID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Message: "無效的分類ID",
				Details: err.Error(),
			})
			return
		}
		channel.CategoryID = categoryObjID
	}

	// 創建頻道
	createdChannel, msgOpt := cc.channelService.CreateChannel(userID, channel)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Status:  "success",
		Message: "創建頻道成功",
		Data:    createdChannel,
	})
}

// UpdateChannel 更新頻道信息
func (cc *ChannelController) UpdateChannel(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 獲取頻道ID
	channelID := c.Param("channel_id")
	if channelID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "頻道ID不能為空",
		})
		return
	}

	// 驗證頻道ID格式
	_, err = primitive.ObjectIDFromHex(channelID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的頻道ID格式",
			Details: err.Error(),
		})
		return
	}

	// 解析請求體
	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的請求格式",
			Details: err.Error(),
		})
		return
	}

	// 構建更新字段
	updates := make(map[string]any)
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.CategoryID != "" {
		categoryObjID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Message: "無效的分類ID",
				Details: err.Error(),
			})
			return
		}
		updates["category_id"] = categoryObjID
	}

	// 檢查是否有字段需要更新
	if len(updates) == 0 {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "沒有提供要更新的字段",
		})
		return
	}

	// 更新頻道
	updatedChannel, msgOpt := cc.channelService.UpdateChannel(userID, channelID, updates)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, updatedChannel, "更新頻道成功")
}

// DeleteChannel 刪除頻道
func (cc *ChannelController) DeleteChannel(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 獲取頻道ID
	channelID := c.Param("channel_id")
	if channelID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "頻道ID不能為空",
		})
		return
	}

	// 驗證頻道ID格式
	_, err = primitive.ObjectIDFromHex(channelID)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無效的頻道ID格式",
			Details: err.Error(),
		})
		return
	}

	// 刪除頻道
	msgOpt := cc.channelService.DeleteChannel(userID, channelID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "刪除頻道成功")
}

// Request DTOs for Channel operations
type CreateChannelRequest struct {
	Name       string `json:"name" binding:"required" example:"一般"`
	Type       string `json:"type" example:"text"`    // "text" or "voice"
	CategoryID string `json:"category_id" example:""` // 可選的分類ID
}

type UpdateChannelRequest struct {
	Name       string `json:"name" example:"一般"`
	Type       string `json:"type" example:"text"`    // "text" or "voice"
	CategoryID string `json:"category_id" example:""` // 可選的分類ID
}
