package controllers

import (
	"bytes"
	"chat_app_backend/app/mocks"
	"chat_app_backend/app/models"
	"chat_app_backend/config"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestNewFileController 測試創建 FileController
func TestNewFileController(t *testing.T) {
	setupTestConfig()
	cfg := &config.Config{}
	mockFileService := new(mocks.FileUploadService)

	controller := NewFileController(cfg, nil, mockFileService)

	assert.NotNil(t, controller)
	assert.Equal(t, cfg, controller.config)
	assert.Equal(t, mockFileService, controller.fileUploadService)
}

// TestFileController_UploadFile 測試通用檔案上傳
func TestFileController_UploadFile(t *testing.T) {
	t.Run("成功上傳檔案", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		fileID := primitive.NewObjectID()
		expectedResult := &models.FileResult{
			ID:       fileID,
			FileName: "test.txt",
			FileURL:  "/uploads/test.txt",
			FileSize: 1024,
		}

		mockFileService.On("UploadFile", mock.Anything, mock.Anything, "user123").
			Return(expectedResult, (*models.MessageOptions)(nil))

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.POST("/files/upload", controller.UploadFile)

		// 創建 multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("test content"))
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/files/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFileService.AssertExpectations(t)
	})

	t.Run("未授權用戶", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)
		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.POST("/files/upload", controller.UploadFile)

		req, _ := http.NewRequest(http.MethodPost, "/files/upload", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("缺少檔案", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)
		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.POST("/files/upload", controller.UploadFile)

		req, _ := http.NewRequest(http.MethodPost, "/files/upload", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestFileController_UploadAvatar 測試頭像上傳
func TestFileController_UploadAvatar(t *testing.T) {
	t.Run("成功上傳頭像", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		fileID := primitive.NewObjectID()
		expectedResult := &models.FileResult{
			ID:       fileID,
			FileName: "avatar.jpg",
			FileURL:  "/uploads/avatars/avatar.jpg",
			FileSize: 2048,
		}

		mockFileService.On("UploadAvatar", mock.Anything, mock.Anything, "user123").
			Return(expectedResult, (*models.MessageOptions)(nil))

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.POST("/files/avatar", controller.UploadAvatar)

		// 創建 multipart form
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "avatar.jpg")
		part.Write([]byte("avatar content"))
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/files/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "頭像上傳成功", response.Message)

		mockFileService.AssertExpectations(t)
	})

	t.Run("上傳頭像失敗", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		mockFileService.On("UploadAvatar", mock.Anything, mock.Anything, "user123").
			Return(nil, &models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Message: "檔案格式不正確",
			})

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.POST("/files/avatar", controller.UploadAvatar)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("avatar", "avatar.txt")
		part.Write([]byte("invalid content"))
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/files/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockFileService.AssertExpectations(t)
	})
}

// TestFileController_UploadDocument 測試文件上傳
func TestFileController_UploadDocument(t *testing.T) {
	t.Run("成功上傳文件", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		fileID := primitive.NewObjectID()
		expectedResult := &models.FileResult{
			ID:       fileID,
			FileName: "document.pdf",
			FileURL:  "/uploads/documents/document.pdf",
			FileSize: 4096,
		}

		mockFileService.On("UploadDocument", mock.Anything, mock.Anything, "user123").
			Return(expectedResult, (*models.MessageOptions)(nil))

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.POST("/files/document", controller.UploadDocument)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("document", "document.pdf")
		part.Write([]byte("pdf content"))
		writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/files/document", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFileService.AssertExpectations(t)
	})
}

// TestFileController_GetUserFiles 測試獲取用戶檔案列表
func TestFileController_GetUserFiles(t *testing.T) {
	t.Run("成功獲取檔案列表", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		fileID1 := primitive.NewObjectID()
		fileID2 := primitive.NewObjectID()
		expectedFiles := []*models.UploadedFile{
			{
				OriginalName: "test1.txt",
				FileName:     "test1.txt",
				FilePath:     "/uploads/test1.txt",
			},
			{
				OriginalName: "test2.pdf",
				FileName:     "test2.pdf",
				FilePath:     "/uploads/test2.pdf",
			},
		}
		expectedFiles[0].ID = fileID1
		expectedFiles[1].ID = fileID2

		mockFileService.On("GetUserFiles", "user123").Return(expectedFiles, (*models.MessageOptions)(nil))

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.GET("/files", controller.GetUserFiles)

		req, _ := http.NewRequest(http.MethodGet, "/files", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)

		mockFileService.AssertExpectations(t)
	})

	t.Run("獲取檔案列表失敗", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		mockFileService.On("GetUserFiles", "user123").Return(
			nil,
			&models.MessageOptions{
				Code:    models.ErrInternalServer,
				Message: "資料庫查詢失敗",
			},
		)

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.GET("/files", controller.GetUserFiles)

		req, _ := http.NewRequest(http.MethodGet, "/files", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		mockFileService.AssertExpectations(t)
	})
}

// TestFileController_DeleteFile 測試刪除檔案
func TestFileController_DeleteFile(t *testing.T) {
	t.Run("成功刪除檔案", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		mockFileService.On("DeleteFileByID", "file123", "user123").Return((*models.MessageOptions)(nil))

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.DELETE("/files/:file_id", controller.DeleteFile)

		req, _ := http.NewRequest(http.MethodDelete, "/files/file123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response models.APIResponse
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "檔案刪除成功", response.Message)

		mockFileService.AssertExpectations(t)
	})

	t.Run("檔案ID為空", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)
		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.DELETE("/files/:file_id", controller.DeleteFile)

		req, _ := http.NewRequest(http.MethodDelete, "/files/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("刪除檔案失敗", func(t *testing.T) {
		mockFileService := new(mocks.FileUploadService)

		mockFileService.On("DeleteFileByID", "file123", "user123").Return(
			&models.MessageOptions{
				Code:    models.ErrInvalidParams,
				Message: "檔案不存在或無權限",
			},
		)

		controller := NewFileController(&config.Config{}, nil, mockFileService)

		router := setupTestRouter()
		router.Use(func(c *gin.Context) {
			c.Set("userID", "user123")
			c.Next()
		})
		router.DELETE("/files/:file_id", controller.DeleteFile)

		req, _ := http.NewRequest(http.MethodDelete, "/files/file123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		mockFileService.AssertExpectations(t)
	})
}
