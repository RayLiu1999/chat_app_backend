package providers

import (
	"context"
	"io"
	"mime/multipart"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// FileProvider - 負責底層文件操作
type FileProvider interface {
	SaveFile(file multipart.File, filename string) (string, error)
	DeleteFile(filepath string) error
	GetFileInfo(filepath string) (os.FileInfo, error)
	GetFileURL(filePath string) string
	GetFile(filepath string) (io.ReadCloser, error)
}

// ODM - 提供對模型的資料庫操作介面
type ODM interface {
	// ===== 基礎工具方法 =====

	// GetDatabase 返回數據庫連接
	GetDatabase() *mongo.Database

	// Collection 獲取模型對應的集合
	Collection(model Model) *mongo.Collection

	// ===== 創建操作 =====

	// Create 創建新文檔
	Create(ctx context.Context, model Model) error

	// InsertMany 插入多個文檔
	InsertMany(ctx context.Context, models []Model) error

	// ===== 查詢操作 =====

	// FindByID 通過ID查找文檔
	FindByID(ctx context.Context, ID string, model Model) error

	// FindOne 查找單個文檔
	FindOne(ctx context.Context, filter bson.M, model Model) error

	// Find 查找多個文檔
	Find(ctx context.Context, filter bson.M, models any) error

	// ===== 更新操作 =====

	// Update 更新文檔
	Update(ctx context.Context, model Model) error

	// UpdateFields 更新文檔的特定欄位
	UpdateFields(ctx context.Context, model Model, fields bson.M) error

	// UpdateMany 更新多個文檔
	UpdateMany(ctx context.Context, model Model, filter bson.M, update bson.M) error

	// ===== 刪除操作 =====

	// Delete 刪除文檔
	Delete(ctx context.Context, model Model) error

	// DeleteMany 刪除多個文檔
	DeleteMany(ctx context.Context, model Model, filter bson.M) error

	// DeleteByID 通過ID刪除文檔
	DeleteByID(ctx context.Context, ID string, model Model) error

	// ===== 統計和工具方法 =====

	// Count 計算符合條件的文檔數量
	Count(ctx context.Context, filter bson.M, model Model) (int64, error)

	// ===== 存在性檢查 =====

	// Exists 檢查文檔是否存在
	Exists(ctx context.Context, filter bson.M, model Model) (bool, error)

	// ExistsByID 通過ID檢查文檔是否存在
	ExistsByID(ctx context.Context, ID string, model Model) (bool, error)

	// ===== 高級查詢操作 =====

	// FindWithOptions 使用自定義選項查找文檔
	FindWithOptions(ctx context.Context, filter bson.M, models any, options *QueryOptions) error

	// Aggregate 執行聚合查詢
	Aggregate(ctx context.Context, pipeline any, models any, model Model) error

	// ===== 批量操作 =====

	// BulkWrite 執行批量寫入操作
	BulkWrite(ctx context.Context, operations []mongo.WriteModel, model Model) (*mongo.BulkWriteResult, error)
}
