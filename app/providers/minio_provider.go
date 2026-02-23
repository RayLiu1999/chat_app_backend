package providers

import (
	"chat_app_backend/config"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioProvider struct {
	cfg    *config.Config
	client *minio.Client
	bucket string
}

func NewMinIOProvider(cfg *config.Config) (*minioProvider, error) {
	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKeyID, cfg.MinIO.SecretAccessKey, ""),
		Secure: cfg.MinIO.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("初始化 MinIO 客戶端失敗: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.MinIO.BucketName)
	if err != nil {
		return nil, fmt.Errorf("檢查 Bucket 狀態失敗: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.MinIO.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("建立 Bucket 失敗: %w", err)
		}
	}

	return &minioProvider{
		cfg:    cfg,
		client: client,
		bucket: cfg.MinIO.BucketName,
	}, nil
}

func (mp *minioProvider) SaveFile(file multipart.File, filename string) (string, error) {
	ctx := context.Background()

	// 確保檔案指標在起點
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("重置檔案指針失敗: %w", err)
	}

	// -1 表示由 minio SDK 自動處理大小
	_, err := mp.client.PutObject(ctx, mp.bucket, filename, file, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream", // SDK 會盡力猜測或使用預設
	})
	if err != nil {
		return "", fmt.Errorf("上傳至 MinIO 失敗: %w", err)
	}

	// 回傳相對路徑
	return filename, nil
}

func (mp *minioProvider) DeleteFile(filepath string) error {
	ctx := context.Background()
	err := mp.client.RemoveObject(ctx, mp.bucket, filepath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("從 MinIO 刪除檔案失敗: %w", err)
	}
	return nil
}

func (mp *minioProvider) GetFileInfo(filepath string) (os.FileInfo, error) {
	ctx := context.Background()
	info, err := mp.client.StatObject(ctx, mp.bucket, filepath, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("獲取 MinIO 檔案資訊失敗: %w", err)
	}
	return &minioFileInfo{info: info}, nil
}

func (mp *minioProvider) GetFileURL(filePath string) string {
	baseURL := mp.cfg.MinIO.PublicURL
	baseURL = strings.TrimRight(baseURL, "/")
	cleanFilePath := strings.TrimLeft(filePath, "/")
	return fmt.Sprintf("%s/%s/%s", baseURL, mp.bucket, cleanFilePath)
}

func (mp *minioProvider) GetFile(filepath string) (io.ReadCloser, error) {
	ctx := context.Background()
	object, err := mp.client.GetObject(ctx, mp.bucket, filepath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("取得 MinIO 檔案失敗: %w", err)
	}
	return object, nil
}

// minioFileInfo 實作 os.FileInfo 介面
type minioFileInfo struct {
	info minio.ObjectInfo
}

func (m *minioFileInfo) Name() string       { return m.info.Key }
func (m *minioFileInfo) Size() int64        { return m.info.Size }
func (m *minioFileInfo) Mode() os.FileMode  { return 0644 }
func (m *minioFileInfo) ModTime() time.Time { return m.info.LastModified }
func (m *minioFileInfo) IsDir() bool        { return false }
func (m *minioFileInfo) Sys() any           { return m.info }
