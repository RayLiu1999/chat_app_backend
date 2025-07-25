package controllers

import (
	"chat_app_backend/config"
	"chat_app_backend/services"
	"chat_app_backend/utils"
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
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{
			Code:        utils.ErrUnauthorized,
			Message:     "未找到用戶ID",
			Displayable: true,
		})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無法獲取上傳檔案: " + err.Error(),
			Displayable: false,
		})
		return
	}
	defer file.Close()

	// 上傳檔案
	result, err := fc.fileUploadService.UploadFile(file, header, userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "檔案上傳失敗: " + err.Error(),
			Displayable: true,
		})
		return
	}

	utils.SuccessResponse(c, result, utils.MessageOptions{
		Message:     "檔案上傳成功",
		Displayable: true,
	})
}

// UploadAvatar 頭像上傳
func (fc *FileController) UploadAvatar(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{
			Code:        utils.ErrUnauthorized,
			Message:     "未找到用戶ID",
			Displayable: true,
		})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無法獲取頭像檔案: " + err.Error(),
			Displayable: false,
		})
		return
	}
	defer file.Close()

	// 上傳頭像
	result, err := fc.fileUploadService.UploadAvatar(file, header, userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "頭像上傳失敗: " + err.Error(),
			Displayable: true,
		})
		return
	}

	utils.SuccessResponse(c, result, utils.MessageOptions{
		Message:     "頭像上傳成功",
		Displayable: true,
	})
}

// UploadDocument 文件上傳
func (fc *FileController) UploadDocument(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{
			Code:        utils.ErrUnauthorized,
			Message:     "未找到用戶ID",
			Displayable: true,
		})
		return
	}

	// 獲取上傳的檔案
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "無法獲取文件檔案: " + err.Error(),
			Displayable: false,
		})
		return
	}
	defer file.Close()

	// 上傳文件
	result, err := fc.fileUploadService.UploadDocument(file, header, userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "文件上傳失敗: " + err.Error(),
			Displayable: true,
		})
		return
	}

	utils.SuccessResponse(c, result, utils.MessageOptions{
		Message:     "文件上傳成功",
		Displayable: true,
	})
}

// GetUserFiles 獲取用戶檔案列表
func (fc *FileController) GetUserFiles(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{
			Code:        utils.ErrUnauthorized,
			Message:     "未找到用戶ID",
			Displayable: true,
		})
		return
	}

	// 獲取用戶檔案列表
	files, err := fc.fileUploadService.GetUserFiles(userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, utils.MessageOptions{
			Code:        utils.ErrInternalServer,
			Message:     "獲取檔案列表失敗: " + err.Error(),
			Displayable: true,
		})
		return
	}

	utils.SuccessResponse(c, files, utils.MessageOptions{
		Message:     "獲取檔案列表成功",
		Displayable: true,
	})
}

// DeleteFile 刪除檔案
func (fc *FileController) DeleteFile(c *gin.Context) {
	// 從 JWT 中獲取用戶ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, utils.MessageOptions{
			Code:        utils.ErrUnauthorized,
			Message:     "未找到用戶ID",
			Displayable: true,
		})
		return
	}

	// 獲取檔案ID
	fileID := c.Param("file_id")
	if fileID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "檔案ID不能為空",
			Displayable: true,
		})
		return
	}

	// 刪除檔案
	err := fc.fileUploadService.DeleteFileByID(fileID, userID.(string))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, utils.MessageOptions{
			Code:        utils.ErrInvalidParams,
			Message:     "檔案刪除失敗: " + err.Error(),
			Displayable: true,
		})
		return
	}

	utils.SuccessResponse(c, nil, utils.MessageOptions{
		Message:     "檔案刪除成功",
		Displayable: true,
	})
}
