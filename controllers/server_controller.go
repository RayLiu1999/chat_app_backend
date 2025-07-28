package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// 定義專門的控制器結構體
type ServerController struct {
	config        *config.Config
	mongoConnect  *mongo.Database
	serverService services.ServerServiceInterface
}

// 創建控制器的工廠函數
func NewServerController(cfg *config.Config, mongodb *mongo.Database, serverService services.ServerServiceInterface) *ServerController {
	return &ServerController{
		config:        cfg,
		mongoConnect:  mongodb,
		serverService: serverService,
	}
}

// 獲取用戶的伺服器列表
func (sc *ServerController) GetServerList(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 透過Service獲取伺服器列表
	serverResponses, err := sc.serverService.GetServerListResponse(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponses, utils.MessageOptions{Message: "伺服器列表獲取成功"})
}

// 建立伺服器
func (sc *ServerController) CreateServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 從c取得name和file
	name := c.PostForm("name")
	if name == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器名稱不能為空",
			Displayable: false,
		})
		return
	}

	picture, err := c.FormFile("picture")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無法獲取圖片檔案",
			Displayable: false,
		})
		return
	}

	// 開啟檔案
	file, err := picture.Open()
	if err != nil {
		log.Printf("Error opening file: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無法開啟檔案",
			Displayable: false,
		})
		return
	}
	defer file.Close()

	// 透過Service創建伺服器（包含檔案上傳）
	serverResponse, err := sc.serverService.CreateServer(userID, name, file, picture)
	if err != nil {
		log.Printf("Error creating server: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "創建伺服器失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponse, utils.MessageOptions{Message: "伺服器創建成功"})
}

// SearchPublicServers 搜尋公開伺服器
func (sc *ServerController) SearchPublicServers(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 解析查詢參數
	var request models.ServerSearchRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "查詢參數格式錯誤",
			Displayable: false,
		})
		return
	}

	// 設定預設值
	if request.Page <= 0 {
		request.Page = 1
	}
	if request.Limit <= 0 {
		request.Limit = 20
	}
	if request.SortBy == "" {
		request.SortBy = "created_at"
	}
	if request.SortOrder == "" {
		request.SortOrder = "desc"
	}

	// 透過Service搜尋伺服器
	searchResults, err := sc.serverService.SearchPublicServers(userID, request)
	if err != nil {
		log.Printf("Error searching servers: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "搜尋伺服器失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, searchResults, utils.MessageOptions{Message: "搜尋完成"})
}

// UpdateServer 更新伺服器信息
func (sc *ServerController) UpdateServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: false,
		})
		return
	}

	// 解析請求體
	var updateRequest map[string]interface{}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "請求格式錯誤",
			Displayable: false,
		})
		return
	}

	// 過濾允許更新的欄位
	allowedFields := map[string]bool{
		"name":        true,
		"description": true,
		"is_public":   true,
		"tags":        true,
		"region":      true,
		"max_members": true,
	}

	updates := make(map[string]interface{})
	for key, value := range updateRequest {
		if allowedFields[key] {
			updates[key] = value
		}
	}

	if len(updates) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "沒有有效的更新欄位",
			Displayable: false,
		})
		return
	}

	// 透過Service更新伺服器
	serverResponse, err := sc.serverService.UpdateServer(userID, serverID, updates)
	if err != nil {
		log.Printf("Error updating server: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "更新伺服器失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponse, utils.MessageOptions{Message: "伺服器更新成功"})
}

// DeleteServer 刪除伺服器
func (sc *ServerController) DeleteServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: false,
		})
		return
	}

	// 透過Service刪除伺服器
	err = sc.serverService.DeleteServer(userID, serverID)
	if err != nil {
		log.Printf("Error deleting server: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "刪除伺服器失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "伺服器刪除成功"})
}

// GetServerByID 根據ID獲取伺服器信息
func (sc *ServerController) GetServerByID(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: false,
		})
		return
	}

	// 透過Service獲取伺服器
	serverResponse, err := sc.serverService.GetServerByID(userID, serverID)
	if err != nil {
		log.Printf("Error getting server: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "獲取伺服器失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponse, utils.MessageOptions{Message: "獲取伺服器成功"})
}

// GetServerDetailByID 獲取伺服器詳細信息（包含成員和頻道列表）
func (sc *ServerController) GetServerDetailByID(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: false,
		})
		return
	}

	// 透過Service獲取伺服器詳細信息
	serverDetailResponse, err := sc.serverService.GetServerDetailByID(userID, serverID)
	if err != nil {
		log.Printf("Error getting server detail: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "獲取伺服器詳細信息失敗: " + err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, serverDetailResponse, utils.MessageOptions{Message: "獲取伺服器詳細信息成功"})
}

// JoinServer 請求加入伺服器
func (sc *ServerController) JoinServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: false,
		})
		return
	}

	// 透過Service加入伺服器
	err = sc.serverService.JoinServer(userID, serverID)
	if err != nil {
		log.Printf("Error joining server: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "成功加入伺服器"})
}

// LeaveServer 離開伺服器
func (sc *ServerController) LeaveServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "伺服器ID不能為空",
			Displayable: false,
		})
		return
	}

	// 透過Service離開伺服器
	err = sc.serverService.LeaveServer(userID, serverID)
	if err != nil {
		log.Printf("Error leaving server: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     err.Error(),
			Displayable: false,
		})
		return
	}

	// 返回響應
	utils.SuccessResponse(c, nil, utils.MessageOptions{Message: "成功離開伺服器"})
}
