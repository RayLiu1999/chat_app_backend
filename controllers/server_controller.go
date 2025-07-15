package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/repositories"
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
	userRepo      repositories.UserRepositoryInterface
}

// 創建控制器的工廠函數
func NewServerController(cfg *config.Config, mongodb *mongo.Database, serverService services.ServerServiceInterface, userRepo repositories.UserRepositoryInterface) *ServerController {
	return &ServerController{
		config:        cfg,
		mongoConnect:  mongodb,
		serverService: serverService,
		userRepo:      userRepo,
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

	_, err = sc.userRepo.GetUserById(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound})
		return
	}

	servers, err := sc.serverService.GetServerListByUserId(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	// 定義響應
	serverResponses := make([]models.ServerResponse, len(servers))
	for i, server := range servers {
		serverResponses[i] = models.ServerResponse{
			ID:          server.BaseModel.ID,
			Name:        server.Name,
			PictureURL:  utils.GetUploadURL(server.Picture, nil),
			Description: server.Description,
		}
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponses, utils.MessageOptions{Message: "伺服器列表獲取成功"})
}

// 建立伺服器
func (sc *ServerController) CreateServer(c *gin.Context) {
	// 取得使用者ID
	userID, userObjectId, err := utils.GetUserIDFromHeader(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{Code: utils.ErrUnauthorized})
		return
	}

	_, err = sc.userRepo.GetUserById(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, utils.MessageOptions{Code: utils.ErrUserNotFound})
		return
	}

	// 從c取得name和file
	name := c.PostForm("name")
	picture, err := c.FormFile("picture")
	if err != nil {
		log.Printf("Error getting file: %v", err)
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{Code: utils.ErrInvalidParams})
		return
	}

	// 保存文件
	err = c.SaveUploadedFile(picture, "uploads/"+picture.Filename)
	if err != nil {
		log.Printf("Error saving file: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	server := &models.Server{
		Name:        name,
		Picture:     picture.Filename,
		Description: "This is a test server",
		OwnerID:     userObjectId,
		Members:     []models.Member{{UserID: userObjectId}},
	}

	// 建立伺服器
	createdServer, err := sc.serverService.CreateServer(server)
	if err != nil {
		log.Printf("Error creating server: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{Code: utils.ErrInternalServer})
		return
	}

	// 定義響應
	serverResponse := models.ServerResponse{
		ID:          createdServer.BaseModel.ID,
		Name:        createdServer.Name,
		PictureURL:  utils.GetUploadURL(createdServer.Picture, nil),
		Description: createdServer.Description,
	}

	// 返回響應
	utils.SuccessResponse(c, serverResponse, utils.MessageOptions{Message: "伺服器創建成功"})
}
