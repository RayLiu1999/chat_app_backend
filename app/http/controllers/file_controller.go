package controllers

import (
	"chat_app_backend/app/models"
	"chat_app_backend/app/services"
	"chat_app_backend/config"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type FileController struct {
	config            *config.Config
	mongoConnect      *mongo.Database
	fileUploadService services.FileUploadServiceInterface
}

func NewFileController(cfg *config.Config, mongodb *mongo.Database, fileUploadService services.FileUploadServiceInterface) *FileController {
	return &FileController{
		config:            cfg,
		mongoConnect:      mongodb,
		fileUploadService: fileUploadService,
	}
}

// UploadFile 通用檔案上傳
func (fc *FileController) UploadFile(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "未找到用戶ID",
		})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無法獲取上傳檔案: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 上傳檔案
	result, msgOpt := fc.fileUploadService.UploadFile(file, header, userID.(string))
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, result, "檔案上傳成功")
}

// UploadAvatar 頭像上傳
func (fc *FileController) UploadAvatar(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "未找到用戶ID",
		})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無法獲取頭像檔案",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// 上傳頭像
	result, msgOpt := fc.fileUploadService.UploadAvatar(file, header, userID.(string))
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, result, "頭像上傳成功")
}

// UploadDocument 文件上傳
func (fc *FileController) UploadDocument(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "未找到用戶ID",
		})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "無法獲取文件檔案",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()

	// 上傳文件
	result, msgOpt := fc.fileUploadService.UploadDocument(file, header, userID.(string))
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, result, "文件上傳成功")
}

// GetUserFiles 獲取用戶檔案列表
func (fc *FileController) GetUserFiles(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "未找到用戶ID",
		})
		return
	}

	// 獲取用戶檔案列表
	files, msgOpt := fc.fileUploadService.GetUserFiles(userID.(string))
	if msgOpt != nil {
		ErrorResponse(c, http.StatusInternalServerError, *msgOpt)
		return
	}

	SuccessResponse(c, files, "獲取檔案列表成功")
}

// DeleteFile 刪除檔案
func (fc *FileController) DeleteFile(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, models.MessageOptions{
			Code:    models.ErrUnauthorized,
			Message: "未找到用戶ID",
		})
		return
	}

	// 獲取檔案ID
	fileID := c.Param("file_id")
	if fileID == "" {
		ErrorResponse(c, http.StatusBadRequest, models.MessageOptions{
			Code:    models.ErrInvalidParams,
			Message: "檔案ID不能為空",
		})
		return
	}

	// 刪除檔案
	msgOpt := fc.fileUploadService.DeleteFileByID(fileID, userID.(string))
	if msgOpt != nil {
		ErrorResponse(c, http.StatusBadRequest, *msgOpt)
		return
	}

	SuccessResponse(c, nil, "檔案刪除成功")
}
