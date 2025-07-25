package repositories

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileRepository struct {
	config *config.Config
	odm    *providers.ODM
}

func NewFileRepository(cfg *config.Config, odm *providers.ODM) FileRepositoryInterface {
	return &FileRepository{
		config: cfg,
		odm:    odm,
	}
}

// CreateFile 創建檔案記錄
func (fr *FileRepository) CreateFile(file *models.UploadedFile) error {
	return fr.odm.Create(context.Background(), file)
}

// GetFileByID 根據檔案ID獲取檔案
func (fr *FileRepository) GetFileByID(fileID string) (*models.UploadedFile, error) {
	var file models.UploadedFile
	err := fr.odm.FindByID(context.Background(), fileID, &file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetFileByPath 根據檔案路徑獲取檔案
func (fr *FileRepository) GetFileByPath(filePath string) (*models.UploadedFile, error) {
	qb := providers.NewQueryBuilder()
	qb.Where("file_path", filePath)

	var file models.UploadedFile
	err := fr.odm.FindOne(context.Background(), qb.GetFilter(), &file)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

// GetFilesByUserID 根據用戶ID獲取檔案列表
func (fr *FileRepository) GetFilesByUserID(userID string) ([]models.UploadedFile, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	qb := providers.NewQueryBuilder()
	qb.Where("user_id", userObjID)
	qb.SortDesc("created_at")

	var files []models.UploadedFile
	err = fr.odm.Find(context.Background(), qb.GetFilter(), &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// UpdateFileStatus 更新檔案狀態
func (fr *FileRepository) UpdateFileStatus(fileID string, status string) error {
	var file models.UploadedFile
	err := fr.odm.FindByID(context.Background(), fileID, &file)
	if err != nil {
		return err
	}

	updates := bson.M{"status": status}
	return fr.odm.UpdateFields(context.Background(), &file, updates)
}

// DeleteFileByID 根據檔案ID刪除檔案記錄
func (fr *FileRepository) DeleteFileByID(fileID string) error {
	var file models.UploadedFile
	return fr.odm.DeleteByID(context.Background(), fileID, &file)
}

// DeleteFileByPath 根據檔案路徑刪除檔案記錄
func (fr *FileRepository) DeleteFileByPath(filePath string) error {
	qb := providers.NewQueryBuilder()
	qb.Where("file_path", filePath)

	var file models.UploadedFile
	err := fr.odm.FindOne(context.Background(), qb.GetFilter(), &file)
	if err != nil {
		return err
	}

	return fr.odm.Delete(context.Background(), &file)
}

// GetExpiredFiles 獲取過期檔案列表
func (fr *FileRepository) GetExpiredFiles() ([]models.UploadedFile, error) {
	now := time.Now()

	qb := providers.NewQueryBuilder()
	qb.WhereLt("expires_at", now)
	qb.WhereNotNull("expires_at")

	var files []models.UploadedFile
	err := fr.odm.Find(context.Background(), qb.GetFilter(), &files)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// CleanupExpiredFiles 清理過期檔案記錄
func (fr *FileRepository) CleanupExpiredFiles() error {
	now := time.Now()

	qb := providers.NewQueryBuilder()
	qb.WhereLt("expires_at", now)
	qb.WhereNotNull("expires_at")

	var file models.UploadedFile
	return fr.odm.DeleteMany(context.Background(), &file, qb.GetFilter())
}
