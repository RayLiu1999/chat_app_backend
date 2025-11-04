package mocks

import (
	"chat_app_backend/app/models"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

// FileUploadService 是 services.FileUploadService 介面的 mock 實作
type FileUploadService struct {
	mock.Mock
}

// 業務方法
func (m *FileUploadService) UploadFile(file multipart.File, header *multipart.FileHeader, userID string) (*models.FileResult, *models.MessageOptions) {
	args := m.Called(file, header, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.FileResult), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) UploadAvatar(file multipart.File, header *multipart.FileHeader, userID string) (*models.FileResult, *models.MessageOptions) {
	args := m.Called(file, header, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.FileResult), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) UploadDocument(file multipart.File, header *multipart.FileHeader, userID string) (*models.FileResult, *models.MessageOptions) {
	args := m.Called(file, header, userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.FileResult), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) UploadFileWithConfig(file multipart.File, header *multipart.FileHeader, userID string, config *models.FileUploadConfig) (*models.FileResult, *models.MessageOptions) {
	args := m.Called(file, header, userID, config)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.FileResult), args.Get(1).(*models.MessageOptions)
}

// 驗證方法
func (m *FileUploadService) ValidateFile(header *multipart.FileHeader) *models.MessageOptions {
	args := m.Called(header)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

func (m *FileUploadService) ValidateImage(header *multipart.FileHeader) *models.MessageOptions {
	args := m.Called(header)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

func (m *FileUploadService) ValidateDocument(header *multipart.FileHeader) *models.MessageOptions {
	args := m.Called(header)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// 安全檢查方法
func (m *FileUploadService) ScanFileForMalware(filePath string) *models.MessageOptions {
	args := m.Called(filePath)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

func (m *FileUploadService) CheckFileContent(file multipart.File, header *multipart.FileHeader) *models.MessageOptions {
	args := m.Called(file, header)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

// 檔案管理方法
func (m *FileUploadService) DeleteFile(filePath string) *models.MessageOptions {
	args := m.Called(filePath)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

func (m *FileUploadService) DeleteFileByID(fileID string, userID string) *models.MessageOptions {
	args := m.Called(fileID, userID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}

func (m *FileUploadService) GetFileInfo(filePath string) (*models.FileInfo, *models.MessageOptions) {
	args := m.Called(filePath)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.FileInfo), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) GetFileURLByID(fileID string) (string, *models.MessageOptions) {
	args := m.Called(fileID)
	if args.Get(1) == nil {
		return args.String(0), nil
	}
	return args.String(0), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) GetFileInfoByID(fileID string) (*models.UploadedFile, *models.MessageOptions) {
	args := m.Called(fileID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).(*models.UploadedFile), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) GetUserFiles(userID string) ([]*models.UploadedFile, *models.MessageOptions) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Get(1).(*models.MessageOptions)
	}
	return args.Get(0).([]*models.UploadedFile), args.Get(1).(*models.MessageOptions)
}

func (m *FileUploadService) CleanupExpiredFiles() *models.MessageOptions {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*models.MessageOptions)
}
