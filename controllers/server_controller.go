package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/services"
	"chat_app_backend/utils"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// 定義專門的控制器結構體
type ServerController struct {
	config        *config.Config
	mongoConnect  *mongo.Database
	serverService services.ServerServiceInterface
	userService   services.UserServiceInterface
}

// 創建控制器的工廠函數
func NewServerController(cfg *config.Config, mongodb *mongo.Database, serverService services.ServerServiceInterface, userService services.UserServiceInterface) *ServerController {
	return &ServerController{
		config:        cfg,
		mongoConnect:  mongodb,
		serverService: serverService,
		userService:   userService,
	}
}

// 獲取用戶的伺服器列表
func (sc *ServerController) GetServerList(c *gin.Context) {
	// 取得使用者ID
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "未授權的請求")
		return
	}

	_, err = sc.userService.GetUserById(objectID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.ErrUserNotFound, "使用者不存在")
		return
	}

	servers, err := sc.serverService.GetServerListByUserId(objectID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	// 定義響應
	serverResponses := make([]models.ServerResponse, len(servers))
	for i, server := range servers {
		serverResponses[i] = models.ServerResponse{
			ID:          server.ID,
			Name:        server.Name,
			PictureURL:  utils.GetUploadURL(server.Picture, nil),
			Description: server.Description,
		}
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponses, "伺服器列表獲取成功", 0)
}

// 建立伺服器
func (sc *ServerController) CreateServer(c *gin.Context) {
	// 取得使用者ID
	_, objectID, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.ErrUnauthorized, "未授權的請求")
		return
	}

	_, err = sc.userService.GetUserById(objectID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.ErrUserNotFound, "使用者不存在")
		return
	}

	// 從c取得name和file
	name := c.PostForm("name")
	picture, err := c.FormFile("picture")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, utils.ErrInvalidParams, "請求參數錯誤")
		return
	}

	// 保存文件
	err = c.SaveUploadedFile(picture, "uploads/"+picture.Filename)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	server := &models.Server{
		ID:          primitive.NewObjectID(),
		Name:        name,
		Picture:     picture.Filename,
		Description: "This is a test server",
		OwnerID:     objectID,
		Members:     []models.Member{{UserID: objectID}},
		CreatedAt:   time.Now(),
		UpdateAt:    time.Now(),
	}

	// 建立伺服器
	createdServer, err := sc.serverService.CreateServer(server)
	if err != nil {
		log.Printf("Error creating server: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.ErrInternalServer, "伺服器內部錯誤")
		return
	}

	// 定義響應
	serverResponse := models.ServerResponse{
		ID:          createdServer.ID,
		Name:        createdServer.Name,
		PictureURL:  utils.GetUploadURL(createdServer.Picture, nil),
		Description: createdServer.Description,
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponse, "伺服器創建成功", 0, true)
}
