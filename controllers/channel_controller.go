package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
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
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: true,
		})
		return
	}

	// 獲取頻道列表
	channels, err := cc.channelService.GetChannelsByServerID(userID, serverID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "獲取頻道列表失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	utils.SuccessResponse(c, channels, utils.MessageOptions{
		Message:     "獲取頻道列表成功",
		Displayable: false,
	})
}

// GetChannelByID 獲取單個頻道信息
func (cc *ChannelController) GetChannelByID(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取頻道ID
	channelID := c.Param("channel_id")
	if channelID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "頻道ID不能為空",
			Displayable: true,
		})
		return
	}

	// 驗證頻道ID格式
	_, err = primitive.ObjectIDFromHex(channelID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的頻道ID格式",
			Displayable: true,
		})
		return
	}

	// 獲取頻道信息
	channel, err := cc.channelService.GetChannelByID(userID, channelID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "獲取頻道信息失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	utils.SuccessResponse(c, channel, utils.MessageOptions{
		Message:     "獲取頻道信息成功",
		Displayable: false,
	})
}

// CreateChannel 創建新頻道
func (cc *ChannelController) CreateChannel(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: true,
		})
		return
	}

	// 驗證伺服器ID格式
	_, err = primitive.ObjectIDFromHex(serverID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的伺服器ID格式",
			Displayable: true,
		})
		return
	}

	// 解析請求體
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的請求格式: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 驗證請求數據
	if req.Name == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "頻道名稱不能為空",
			Displayable: true,
		})
		return
	}

	if req.Type == "" {
		req.Type = "text" // 默認為文字頻道
	}

	// 轉換伺服器ID
	serverObjID, err := primitive.ObjectIDFromHex(serverID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的伺服器ID: " + err.Error(),
			Displayable: false,
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
			utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
				Code:        utils.ErrInvalidParams,
				Message:     "無效的分類ID: " + err.Error(),
				Displayable: false,
			})
			return
		}
		channel.CategoryID = categoryObjID
	}

	// 創建頻道
	createdChannel, err := cc.channelService.CreateChannel(userID, channel)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "創建頻道失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	c.JSON(http.StatusCreated, utils.APIResponse{
		Status:      "success",
		Message:     "創建頻道成功",
		Displayable: true,
		Data:        createdChannel,
	})
}

// UpdateChannel 更新頻道信息
func (cc *ChannelController) UpdateChannel(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取頻道ID
	channelID := c.Param("channel_id")
	if channelID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "頻道ID不能為空",
			Displayable: true,
		})
		return
	}

	// 驗證頻道ID格式
	_, err = primitive.ObjectIDFromHex(channelID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的頻道ID格式",
			Displayable: true,
		})
		return
	}

	// 解析請求體
	var req UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的請求格式: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 構建更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.CategoryID != "" {
		categoryObjID, err := primitive.ObjectIDFromHex(req.CategoryID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
				Code:        utils.ErrInvalidParams,
				Message:     "無效的分類ID: " + err.Error(),
				Displayable: false,
			})
			return
		}
		updates["category_id"] = categoryObjID
	}

	// 檢查是否有字段需要更新
	if len(updates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "沒有提供要更新的字段",
			Displayable: true,
		})
		return
	}

	// 更新頻道
	updatedChannel, err := cc.channelService.UpdateChannel(userID, channelID, updates)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "更新頻道失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	utils.SuccessResponse(c, updatedChannel, utils.MessageOptions{
		Message:     "更新頻道成功",
		Displayable: true,
	})
}

// DeleteChannel 刪除頻道
func (cc *ChannelController) DeleteChannel(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 獲取頻道ID
	channelID := c.Param("channel_id")
	if channelID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "頻道ID不能為空",
			Displayable: true,
		})
		return
	}

	// 驗證頻道ID格式
	_, err = primitive.ObjectIDFromHex(channelID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無效的頻道ID格式",
			Displayable: true,
		})
		return
	}

	// 刪除頻道
	err = cc.channelService.DeleteChannel(userID, channelID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "刪除頻道失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{
		Message:     "刪除頻道成功",
		Displayable: true,
	})
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
