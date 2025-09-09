package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"log"
	"mime/multipart"
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
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 透過Service獲取伺服器列表
	serverResponses, msgOpt := sc.serverService.GetServerListResponse(userID)
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, serverResponses, "伺服器列表獲取成功")
}

// 建立伺服器
func (sc *ServerController) CreateServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 從c取得name和file
	name := c.PostForm("name")
	if name == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器名稱不能為空",
		})
		return
	}

	picture, _ := c.FormFile("picture")
	var file multipart.File
	if picture != nil {
		// 開啟檔案
		file, err = picture.Open()
		if err != nil {
			log.Printf("Error opening file: %v", err)
			ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Message: "無法開啟檔案",
			})
			return
		}
		defer file.Close()
	}

	// 透過Service創建伺服器（包含檔案上傳）
	serverResponse, msgOpt := sc.serverService.CreateServer(userID, name, file, picture)
	if msgOpt != nil {
		log.Printf("Error creating server: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, serverResponse, "伺服器創建成功")
}

// SearchPublicServers 搜尋公開伺服器
func (sc *ServerController) SearchPublicServers(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 解析查詢參數
	var request models.ServerSearchRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "查詢參數格式錯誤",
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
	searchResults, msgOpt := sc.serverService.SearchPublicServers(userID, request)
	if msgOpt != nil {
		log.Printf("Error searching servers: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, searchResults, "搜尋完成")
}

// UpdateServer 更新伺服器信息
func (sc *ServerController) UpdateServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 解析請求體
	var updateRequest map[string]interface{}
	if err := c.ShouldBindJSON(&updateRequest); err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "請求格式錯誤",
			Details: err.Error(),
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
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "沒有有效的更新欄位",
		})
		return
	}

	// 透過Service更新伺服器
	serverResponse, msgOpt := sc.serverService.UpdateServer(userID, serverID, updates)
	if msgOpt != nil {
		log.Printf("Error updating server: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, serverResponse, "伺服器更新成功")
}

// DeleteServer 刪除伺服器
func (sc *ServerController) DeleteServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 透過Service刪除伺服器
	msgOpt := sc.serverService.DeleteServer(userID, serverID)
	if msgOpt != nil {
		log.Printf("Error deleting server: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, nil, "伺服器刪除成功")
}

// GetServerByID 根據ID獲取伺服器信息
func (sc *ServerController) GetServerByID(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 透過Service獲取伺服器
	serverResponse, msgOpt := sc.serverService.GetServerByID(userID, serverID)
	if msgOpt != nil {
		log.Printf("Error getting server: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, serverResponse, "獲取伺服器成功")
}

// GetServerDetailByID 獲取伺服器詳細信息（包含成員和頻道列表）
func (sc *ServerController) GetServerDetailByID(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 透過Service獲取伺服器詳細信息
	serverDetailResponse, msgOpt := sc.serverService.GetServerDetailByID(userID, serverID)
	if msgOpt != nil {
		log.Printf("Error getting server detail: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, serverDetailResponse, "獲取伺服器詳細信息成功")
}

// JoinServer 請求加入伺服器
func (sc *ServerController) JoinServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 透過Service加入伺服器
	msgOpt := sc.serverService.JoinServer(userID, serverID)
	if msgOpt != nil {
		log.Printf("Error joining server: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, nil, "加入伺服器成功")
}

// LeaveServer 離開伺服器
func (sc *ServerController) LeaveServer(c *gin.Context) {
	// 取得使用者ID
	userID, _, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{Code: models.ErrUnauthorized})
		return
	}

	// 取得伺服器ID
	serverID := c.Param("server_id")
	if serverID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "伺服器ID不能為空",
		})
		return
	}

	// 透過Service離開伺服器
	msgOpt := sc.serverService.LeaveServer(userID, serverID)
	if msgOpt != nil {
		log.Printf("Error leaving server: %v", msgOpt.Details)
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	// 返回響應
	SuccessResponse(c, nil, "成功離開伺服器")
}
