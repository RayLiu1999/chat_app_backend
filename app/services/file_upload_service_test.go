package services

import (
	"bytes"
	"chat_app_backend/app/models"
	"chat_app_backend/app/providers"
	"errors"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Mock Components for File Upload Service ---

// mockFileProvider 模擬 FileProvider
type mockFileProvider struct {
	mock.Mock
}

func (m *mockFileProvider) SaveFile(file multipart.File, filename string) (string, error) {
	args := m.Called(file, filename)
	return args.String(0), args.Error(1)
}

func (m *mockFileProvider) DeleteFile(filepath string) error {
	args := m.Called(filepath)
	return args.Error(0)
}

func (m *mockFileProvider) GetFileInfo(filepath string) (os.FileInfo, error) {
	args := m.Called(filepath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(os.FileInfo), args.Error(1)
}

func (m *mockFileProvider) GetFileURL(filePath string) string {
	args := m.Called(filePath)
	return args.String(0)
}

func (m *mockFileProvider) GetFile(filepath string) (io.ReadCloser, error) {
	args := m.Called(filepath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// mockFileRepository 模擬 FileRepository
type mockFileRepository struct {
	mock.Mock
}

func (m *mockFileRepository) CreateFile(file *models.UploadedFile) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *mockFileRepository) GetFileByID(fileID string) (*models.UploadedFile, error) {
	args := m.Called(fileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UploadedFile), args.Error(1)
}

func (m *mockFileRepository) GetFileByPath(filePath string) (*models.UploadedFile, error) {
	args := m.Called(filePath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UploadedFile), args.Error(1)
}

func (m *mockFileRepository) GetFilesByUserID(userID string) ([]models.UploadedFile, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UploadedFile), args.Error(1)
}

func (m *mockFileRepository) UpdateFileStatus(fileID string, status string) error {
	args := m.Called(fileID, status)
	return args.Error(0)
}

func (m *mockFileRepository) DeleteFileByID(fileID string) error {
	args := m.Called(fileID)
	return args.Error(0)
}

func (m *mockFileRepository) DeleteFileByPath(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *mockFileRepository) GetExpiredFiles() ([]models.UploadedFile, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.UploadedFile), args.Error(1)
}

func (m *mockFileRepository) CleanupExpiredFiles() error {
	args := m.Called()
	return args.Error(0)
}

// mockFileInfo 模擬 os.FileInfo
type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

// 創建測試用的 multipart.FileHeader
func createTestFileHeader(filename string, size int64, contentType string) *multipart.FileHeader {
	header := make(textproto.MIMEHeader)
	header.Set("Content-Type", contentType)
	header.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)

	return &multipart.FileHeader{
		Filename: filename,
		Header:   header,
		Size:     size,
	}
}

// testFile 實現 multipart.File 接口
type testFile struct {
	*bytes.Reader
	closed bool
}

func (t *testFile) Close() error {
	t.closed = true
	return nil
}

func (t *testFile) ReadAt(p []byte, off int64) (n int, err error) {
	return t.Reader.ReadAt(p, off)
}

// 創建測試用的 multipart.File
func createTestFile(content []byte) multipart.File {
	return &testFile{
		Reader: bytes.NewReader(content),
		closed: false,
	}
}

// --- Tests ---

func TestNewFileUploadService(t *testing.T) {
	mockFileProvider := new(mockFileProvider)
	mockFileRepo := new(mockFileRepository)

	service := NewFileUploadService(nil, mockFileProvider, nil, mockFileRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockFileProvider, service.fileProvider)
	assert.Equal(t, mockFileRepo, service.fileRepo)
}

func TestValidateFile(t *testing.T) {
	service := &fileUploadService{}

	t.Run("成功驗證檔案", func(t *testing.T) {
		header := createTestFileHeader("test.jpg", 1024, "image/jpeg")
		msgOpt := service.ValidateFile(header)
		assert.Nil(t, msgOpt)
	})

	t.Run("檔案名稱為空", func(t *testing.T) {
		header := createTestFileHeader("", 1024, "image/jpeg")
		msgOpt := service.ValidateFile(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案名稱不能為空")
	})

	t.Run("檔案名稱過長", func(t *testing.T) {
		longName := string(make([]byte, 256))
		for i := range longName {
			longName = string(append([]byte(longName[:i]), 'a'))
		}
		header := createTestFileHeader(longName+".jpg", 1024, "image/jpeg")
		msgOpt := service.ValidateFile(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案名稱過長")
	})

	t.Run("檔案大小無效", func(t *testing.T) {
		header := createTestFileHeader("test.jpg", 0, "image/jpeg")
		msgOpt := service.ValidateFile(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案大小無效")
	})

	t.Run("檔案名稱包含危險字符", func(t *testing.T) {
		dangerousNames := []string{
			"../test.jpg",
			"..\\test.jpg",
			"test<.jpg",
			"test>.jpg",
			"test:.jpg",
			"test\".jpg",
			"test|.jpg",
			"test?.jpg",
			"test*.jpg",
		}

		for _, name := range dangerousNames {
			header := createTestFileHeader(name, 1024, "image/jpeg")
			msgOpt := service.ValidateFile(header)
			assert.NotNil(t, msgOpt, "應該拒絕檔案名稱: %s", name)
			assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
			assert.Contains(t, msgOpt.Message, "不允許的字符")
		}
	})

	t.Run("沒有副檔名", func(t *testing.T) {
		header := createTestFileHeader("test", 1024, "image/jpeg")
		msgOpt := service.ValidateFile(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "必須有副檔名")
	})

	t.Run("危險的副檔名", func(t *testing.T) {
		dangerousExts := []string{
			"test.exe",
			"test.bat",
			"test.cmd",
			"test.scr",
			"test.vbs",
			"test.js",
			"test.sh",
		}

		for _, name := range dangerousExts {
			header := createTestFileHeader(name, 1024, "application/x-msdownload")
			msgOpt := service.ValidateFile(header)
			assert.NotNil(t, msgOpt, "應該拒絕檔案: %s", name)
			assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
			assert.Contains(t, msgOpt.Message, "不允許的檔案類型")
		}
	})
}

func TestValidateImage(t *testing.T) {
	service := &fileUploadService{}

	t.Run("成功驗證圖片", func(t *testing.T) {
		header := createTestFileHeader("test.jpg", 1024*100, "image/jpeg")
		msgOpt := service.ValidateImage(header)
		assert.Nil(t, msgOpt)
	})

	t.Run("圖片大小超過限制", func(t *testing.T) {
		header := createTestFileHeader("test.jpg", 10*1024*1024+1, "image/jpeg")
		msgOpt := service.ValidateImage(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "圖片檔案大小超過限制")
	})

	t.Run("不支援的圖片格式", func(t *testing.T) {
		header := createTestFileHeader("test.svg", 1024, "image/svg+xml")
		msgOpt := service.ValidateImage(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "不支援的圖片格式")
	})

	t.Run("不支援的圖片MIME類型", func(t *testing.T) {
		header := createTestFileHeader("test.jpg", 1024, "image/svg+xml")
		msgOpt := service.ValidateImage(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "不支援的圖片類型")
	})
}

func TestValidateDocument(t *testing.T) {
	service := &fileUploadService{}

	t.Run("成功驗證文件", func(t *testing.T) {
		header := createTestFileHeader("test.pdf", 1024*100, "application/pdf")
		msgOpt := service.ValidateDocument(header)
		assert.Nil(t, msgOpt)
	})

	t.Run("文件大小超過限制", func(t *testing.T) {
		header := createTestFileHeader("test.pdf", 51*1024*1024+1, "application/pdf")
		msgOpt := service.ValidateDocument(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "文件大小超過限制")
	})

	t.Run("不支援的文件格式", func(t *testing.T) {
		header := createTestFileHeader("test.exe", 1024, "application/x-msdownload")
		msgOpt := service.ValidateDocument(header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		// ValidateFile 會先檢查危險副檔名
		assert.Contains(t, msgOpt.Message, "不允許的檔案類型")
	})
}

func TestCheckFileContent(t *testing.T) {
	service := &fileUploadService{}

	t.Run("成功檢查圖片內容", func(t *testing.T) {
		// JPEG 檔案魔術數字
		content := []byte{0xFF, 0xD8, 0xFF, 0xE0}
		content = append(content, make([]byte, 508)...)
		file := createTestFile(content)
		header := createTestFileHeader("test.jpg", int64(len(content)), "image/jpeg")

		msgOpt := service.CheckFileContent(file, header)
		assert.Nil(t, msgOpt)
	})

	t.Run("檔案內容與聲明類型不符", func(t *testing.T) {
		// 文字內容但聲明為圖片
		content := []byte("This is plain text content")
		file := createTestFile(content)
		header := createTestFileHeader("fake.jpg", int64(len(content)), "image/jpeg")

		msgOpt := service.CheckFileContent(file, header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案實際類型與聲明類型不符")
	})

	t.Run("檔案包含惡意內容 - PE執行檔", func(t *testing.T) {
		content := []byte("MZ")
		content = append(content, make([]byte, 510)...)
		file := createTestFile(content)
		header := createTestFileHeader("malware.txt", int64(len(content)), "text/plain")

		msgOpt := service.CheckFileContent(file, header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "可疑內容")
	})

	t.Run("檔案包含惡意內容 - PHP腳本", func(t *testing.T) {
		content := []byte("<?php eval($_GET['cmd']); ?>")
		file := createTestFile(content)
		header := createTestFileHeader("shell.txt", int64(len(content)), "text/plain")

		msgOpt := service.CheckFileContent(file, header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})

	t.Run("檔案包含惡意內容 - Shell腳本", func(t *testing.T) {
		content := []byte("#!/bin/bash\nrm -rf /")
		file := createTestFile(content)
		header := createTestFileHeader("script.txt", int64(len(content)), "text/plain")

		msgOpt := service.CheckFileContent(file, header)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})
}

func TestScanFileForMalware(t *testing.T) {
	t.Run("無法開啟檔案", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileProvider.On("GetFile", "/nonexistent/file.txt").Return(nil, errors.New("file not found")).Once()
		service := &fileUploadService{
			fileProvider: mockFileProvider,
		}
		msgOpt := service.ScanFileForMalware("/nonexistent/file.txt")
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "無法開啟檔案")
	})
}

func TestGetUserFiles(t *testing.T) {
	t.Run("成功獲取用戶檔案列表", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		userID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		files := []models.UploadedFile{
			{
				BaseModel:    providers.BaseModel{ID: primitive.NewObjectID()},
				UserID:       userID,
				FileName:     "test1.jpg",
				OriginalName: "test1.jpg",
			},
			{
				BaseModel:    providers.BaseModel{ID: primitive.NewObjectID()},
				UserID:       userID,
				FileName:     "test2.pdf",
				OriginalName: "test2.pdf",
			},
		}

		mockFileRepo.On("GetFilesByUserID", userID.Hex()).Return(files, nil).Once()

		result, msgOpt := service.GetUserFiles(userID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "test1.jpg", result[0].FileName)
		assert.Equal(t, "test2.pdf", result[1].FileName)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("用戶ID為空", func(t *testing.T) {
		service := &fileUploadService{}

		result, msgOpt := service.GetUserFiles("")

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "用戶ID不能為空")
	})

	t.Run("獲取檔案列表失敗", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		userID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		mockFileRepo.On("GetFilesByUserID", userID.Hex()).Return(nil, errors.New("database error")).Once()

		result, msgOpt := service.GetUserFiles(userID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestDeleteFileByID(t *testing.T) {
	t.Run("成功刪除檔案", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		userID := primitive.NewObjectID()
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			UserID:    userID,
			FilePath:  "/uploads/test.jpg",
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()
		mockFileProvider.On("DeleteFile", "/uploads/test.jpg").Return(nil).Once()
		mockFileRepo.On("DeleteFileByID", fileID.Hex()).Return(nil).Once()

		msgOpt := service.DeleteFileByID(fileID.Hex(), userID.Hex())

		assert.Nil(t, msgOpt)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})

	t.Run("檔案ID為空", func(t *testing.T) {
		service := &fileUploadService{}

		msgOpt := service.DeleteFileByID("", primitive.NewObjectID().Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案ID不能為空")
	})

	t.Run("用戶ID為空", func(t *testing.T) {
		service := &fileUploadService{}

		msgOpt := service.DeleteFileByID(primitive.NewObjectID().Hex(), "")

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "用戶ID不能為空")
	})

	t.Run("檔案不存在", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(nil, errors.New("file not found")).Once()

		msgOpt := service.DeleteFileByID(fileID.Hex(), primitive.NewObjectID().Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrNotFound, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案不存在")

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("沒有權限刪除檔案", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)

		ownerID := primitive.NewObjectID()
		userID := primitive.NewObjectID()
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			UserID:    ownerID,
			FilePath:  "/uploads/test.jpg",
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()

		msgOpt := service.DeleteFileByID(fileID.Hex(), userID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrUnauthorized, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "沒有權限刪除此檔案")

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("刪除檔案系統檔案失敗", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		userID := primitive.NewObjectID()
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			UserID:    userID,
			FilePath:  "/uploads/test.jpg",
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()
		mockFileProvider.On("DeleteFile", "/uploads/test.jpg").Return(errors.New("delete error")).Once()

		msgOpt := service.DeleteFileByID(fileID.Hex(), userID.Hex())

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "刪除檔案失敗")

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})
}

func TestGetFileURLByID(t *testing.T) {
	t.Run("成功獲取檔案URL", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			FilePath:  "/uploads/test.jpg",
			Status:    "verified",
		}

		fileInfo := &mockFileInfo{
			name: "test.jpg",
			size: 1024,
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()
		mockFileProvider.On("GetFileInfo", "/uploads/test.jpg").Return(fileInfo, nil).Once()
		mockFileProvider.On("GetFileURL", "/uploads/test.jpg").Return("https://example.com/uploads/test.jpg").Once()

		url, msgOpt := service.GetFileURLByID(fileID.Hex())

		assert.Nil(t, msgOpt)
		assert.Equal(t, "https://example.com/uploads/test.jpg", url)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})

	t.Run("檔案ID為空", func(t *testing.T) {
		service := &fileUploadService{}

		url, msgOpt := service.GetFileURLByID("")

		assert.Empty(t, url)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案ID不能為空")
	})

	t.Run("檔案不存在", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(nil, errors.New("file not found")).Once()

		url, msgOpt := service.GetFileURLByID(fileID.Hex())

		assert.Empty(t, url)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrNotFound, msgOpt.Code)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("檔案尚未驗證", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			FilePath:  "/uploads/test.jpg",
			Status:    "processing",
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()

		url, msgOpt := service.GetFileURLByID(fileID.Hex())

		assert.Empty(t, url)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案尚未驗證")

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("檔案不存在於檔案系統", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			FilePath:  "/uploads/test.jpg",
			Status:    "verified",
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()
		mockFileProvider.On("GetFileInfo", "/uploads/test.jpg").Return(nil, errors.New("file not found")).Once()

		url, msgOpt := service.GetFileURLByID(fileID.Hex())

		assert.Empty(t, url)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrNotFound, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "檔案不存在或已被刪除")

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})
}

func TestGetFileInfoByID(t *testing.T) {
	t.Run("成功獲取檔案資訊", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		file := &models.UploadedFile{
			BaseModel:    providers.BaseModel{ID: fileID},
			FileName:     "test.jpg",
			OriginalName: "original.jpg",
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(file, nil).Once()

		result, msgOpt := service.GetFileInfoByID(fileID.Hex())

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "test.jpg", result.FileName)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("檔案ID為空", func(t *testing.T) {
		service := &fileUploadService{}

		result, msgOpt := service.GetFileInfoByID("")

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInvalidParams, msgOpt.Code)
	})

	t.Run("檔案不存在", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)
		fileID := primitive.NewObjectID()

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		mockFileRepo.On("GetFileByID", fileID.Hex()).Return(nil, errors.New("file not found")).Once()

		result, msgOpt := service.GetFileInfoByID(fileID.Hex())

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrNotFound, msgOpt.Code)

		mockFileRepo.AssertExpectations(t)
	})
}

func TestDeleteFile(t *testing.T) {
	t.Run("成功刪除檔案", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		filePath := "/uploads/test.jpg"

		mockFileRepo.On("DeleteFileByPath", filePath).Return(nil).Once()
		mockFileProvider.On("DeleteFile", filePath).Return(nil).Once()

		msgOpt := service.DeleteFile(filePath)

		assert.Nil(t, msgOpt)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})

	t.Run("刪除資料庫記錄失敗", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		filePath := "/uploads/test.jpg"

		mockFileRepo.On("DeleteFileByPath", filePath).Return(errors.New("database error")).Once()

		msgOpt := service.DeleteFile(filePath)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "刪除檔案記錄失敗")

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("刪除檔案系統檔案失敗", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		filePath := "/uploads/test.jpg"

		mockFileRepo.On("DeleteFileByPath", filePath).Return(nil).Once()
		mockFileProvider.On("DeleteFile", filePath).Return(errors.New("delete error")).Once()

		msgOpt := service.DeleteFile(filePath)

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)
		assert.Contains(t, msgOpt.Message, "刪除檔案失敗")

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})
}

func TestGetFileInfo(t *testing.T) {
	t.Run("成功獲取檔案資訊", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		filePath := "/uploads/test.jpg"
		fileID := primitive.NewObjectID()

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID, CreatedAt: time.Now()},
			FileName:  "test.jpg",
			FilePath:  filePath,
			MimeType:  "image/jpeg",
		}

		fileInfo := &mockFileInfo{
			name:    "test.jpg",
			size:    1024,
			modTime: time.Now(),
		}

		mockFileRepo.On("GetFileByPath", filePath).Return(file, nil).Once()
		mockFileProvider.On("GetFileInfo", filePath).Return(fileInfo, nil).Once()

		result, msgOpt := service.GetFileInfo(filePath)

		assert.Nil(t, msgOpt)
		assert.NotNil(t, result)
		assert.Equal(t, "test.jpg", result.FileName)
		assert.Equal(t, int64(1024), result.FileSize)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})

	t.Run("取得檔案記錄失敗", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		filePath := "/uploads/test.jpg"

		mockFileRepo.On("GetFileByPath", filePath).Return(nil, errors.New("file not found")).Once()

		result, msgOpt := service.GetFileInfo(filePath)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrNotFound, msgOpt.Code)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("取得檔案系統資訊失敗", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		filePath := "/uploads/test.jpg"
		fileID := primitive.NewObjectID()

		file := &models.UploadedFile{
			BaseModel: providers.BaseModel{ID: fileID},
			FileName:  "test.jpg",
			FilePath:  filePath,
		}

		mockFileRepo.On("GetFileByPath", filePath).Return(file, nil).Once()
		mockFileProvider.On("GetFileInfo", filePath).Return(nil, errors.New("file not found")).Once()

		result, msgOpt := service.GetFileInfo(filePath)

		assert.Nil(t, result)
		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})
}

func TestCleanupExpiredFiles(t *testing.T) {
	t.Run("成功清理過期檔案", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		expiredFiles := []models.UploadedFile{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				FilePath:  "/uploads/expired1.jpg",
			},
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				FilePath:  "/uploads/expired2.jpg",
			},
		}

		mockFileRepo.On("GetExpiredFiles").Return(expiredFiles, nil).Once()
		mockFileRepo.On("DeleteFileByPath", "/uploads/expired1.jpg").Return(nil).Once()
		mockFileProvider.On("DeleteFile", "/uploads/expired1.jpg").Return(nil).Once()
		mockFileRepo.On("DeleteFileByPath", "/uploads/expired2.jpg").Return(nil).Once()
		mockFileProvider.On("DeleteFile", "/uploads/expired2.jpg").Return(nil).Once()

		msgOpt := service.CleanupExpiredFiles()

		assert.Nil(t, msgOpt)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})

	t.Run("獲取過期檔案列表失敗", func(t *testing.T) {
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileRepo: mockFileRepo,
		}

		mockFileRepo.On("GetExpiredFiles").Return(nil, errors.New("database error")).Once()

		msgOpt := service.CleanupExpiredFiles()

		assert.NotNil(t, msgOpt)
		assert.Equal(t, models.ErrInternalServer, msgOpt.Code)

		mockFileRepo.AssertExpectations(t)
	})

	t.Run("部分檔案刪除失敗但繼續處理", func(t *testing.T) {
		mockFileProvider := new(mockFileProvider)
		mockFileRepo := new(mockFileRepository)

		service := &fileUploadService{
			fileProvider: mockFileProvider,
			fileRepo:     mockFileRepo,
		}

		expiredFiles := []models.UploadedFile{
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				FilePath:  "/uploads/expired1.jpg",
			},
			{
				BaseModel: providers.BaseModel{ID: primitive.NewObjectID()},
				FilePath:  "/uploads/expired2.jpg",
			},
		}

		mockFileRepo.On("GetExpiredFiles").Return(expiredFiles, nil).Once()
		// 第一個檔案刪除失敗
		mockFileRepo.On("DeleteFileByPath", "/uploads/expired1.jpg").Return(errors.New("delete error")).Once()
		// 第二個檔案成功刪除
		mockFileRepo.On("DeleteFileByPath", "/uploads/expired2.jpg").Return(nil).Once()
		mockFileProvider.On("DeleteFile", "/uploads/expired2.jpg").Return(nil).Once()

		msgOpt := service.CleanupExpiredFiles()

		// 應該返回 nil，因為會繼續處理其他檔案
		assert.Nil(t, msgOpt)

		mockFileRepo.AssertExpectations(t)
		mockFileProvider.AssertExpectations(t)
	})
}

func TestContainsMaliciousContent(t *testing.T) {
	service := &fileUploadService{}

	t.Run("檢測到PE執行檔標誌", func(t *testing.T) {
		content := []byte("MZ")
		result := service.containsMaliciousContent(content)
		assert.True(t, result)
	})

	t.Run("檢測到PHP腳本", func(t *testing.T) {
		content := []byte("<?php system('ls'); ?>")
		result := service.containsMaliciousContent(content)
		assert.True(t, result)
	})

	t.Run("檢測到eval函數", func(t *testing.T) {
		content := []byte("eval(base64_decode('...'))")
		result := service.containsMaliciousContent(content)
		assert.True(t, result)
	})

	t.Run("檢測到Shell腳本", func(t *testing.T) {
		content := []byte("#!/bin/bash\necho 'test'")
		result := service.containsMaliciousContent(content)
		assert.True(t, result)
	})

	t.Run("正常內容不被檢測", func(t *testing.T) {
		content := []byte("This is normal text content without any malicious code.")
		result := service.containsMaliciousContent(content)
		assert.False(t, result)
	})
}

func TestIsMimeTypeCompatible(t *testing.T) {
	service := &fileUploadService{}

	t.Run("完全相同的MIME類型", func(t *testing.T) {
		result := service.isMimeTypeCompatible("image/jpeg", "image/jpeg")
		assert.True(t, result)
	})

	t.Run("JPEG相容類型", func(t *testing.T) {
		result := service.isMimeTypeCompatible("image/jpeg", "image/jpg")
		assert.True(t, result)
		result = service.isMimeTypeCompatible("image/jpg", "image/jpeg")
		assert.True(t, result)
	})

	t.Run("PNG類型", func(t *testing.T) {
		result := service.isMimeTypeCompatible("image/png", "image/png")
		assert.True(t, result)
	})

	t.Run("不相容的類型", func(t *testing.T) {
		result := service.isMimeTypeCompatible("image/png", "image/jpeg")
		assert.False(t, result)
	})

	t.Run("text/plain 與 octet-stream 相容", func(t *testing.T) {
		result := service.isMimeTypeCompatible("application/octet-stream", "text/plain")
		assert.True(t, result)
	})
}
