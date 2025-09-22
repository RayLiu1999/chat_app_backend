package models

import (
	"chat_app_backend/app/providers"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UploadedFile 上傳檔案記錄
type UploadedFile struct {
	providers.BaseModel `bson:",inline"`
	UserID              primitive.ObjectID `json:"user_id" bson:"user_id"`
	OriginalName        string             `json:"original_name" bson:"original_name"`
	FileName            string             `json:"file_name" bson:"file_name"`
	FilePath            string             `json:"file_path" bson:"file_path"`
	FileSize            int64              `json:"file_size" bson:"file_size"`
	MimeType            string             `json:"mime_type" bson:"mime_type"`
	FileType            string             `json:"file_type" bson:"file_type"` // "avatar", "document", "image", "general"
	Status              string             `json:"status" bson:"status"`       // "uploaded", "processing", "verified", "failed"
	Hash                string             `json:"hash" bson:"hash"`           // 檔案SHA256雜湊值
	ExpiresAt           *time.Time         `json:"expires_at" bson:"expires_at,omitempty"`
}

func (u *UploadedFile) GetCollectionName() string {
	return "uploaded_files"
}

// FileUploadConfig 檔案上傳配置
type FileUploadConfig struct {
	FileType          string   `json:"file_type"`          // 檔案類型，例如 "avatar", "document", "image", "general"
	MaxFileSize       int64    `json:"max_file_size"`      // 最大檔案大小 (bytes)
	AllowedMimeTypes  []string `json:"allowed_mime_types"` // 允許的MIME類型
	AllowedExtensions []string `json:"allowed_extensions"` // 允許的副檔名
	UploadPath        string   `json:"upload_path"`        // 上傳路徑
	TempPath          string   `json:"temp_path"`          // 臨時檔案路徑
	RequireAuth       bool     `json:"require_auth"`       // 是否需要認證
	ScanMalware       bool     `json:"scan_malware"`       // 是否掃描惡意軟體
}

// GetServerUploadConfig 取得伺服器上傳配置
func GetServerUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		FileType:    "server",
		MaxFileSize: 5 * 1024 * 1024, // 3MB
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		UploadPath:        "uploads/servers",
		TempPath:          "uploads/temp",
		RequireAuth:       true,
		ScanMalware:       true,
	}
}

// GetAvatarUploadConfig 取得頭像上傳配置
func GetAvatarUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		FileType:    "avatar",
		MaxFileSize: 5 * 1024 * 1024, // 3MB
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
		},
		AllowedExtensions: []string{".jpg", ".jpeg", ".png"},
		UploadPath:        "uploads/avatars",
		TempPath:          "uploads/temp",
		RequireAuth:       true,
		ScanMalware:       true,
	}
}

// 使用者橫幅上傳配置
func GetBannerUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		FileType:    "banner",
		MaxFileSize: 5 * 1024 * 1024, // 5MB
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		UploadPath:        "uploads/banners",
		TempPath:          "uploads/temp",
		RequireAuth:       true,
		ScanMalware:       true,
	}
}

// GetDocumentUploadConfig 取得文件上傳配置
func GetDocumentUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		FileType:    "document",
		MaxFileSize: 50 * 1024 * 1024, // 50MB
		AllowedMimeTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"text/plain",
		},
		AllowedExtensions: []string{".pdf", ".doc", ".docx", ".txt"},
		UploadPath:        "uploads/documents",
		TempPath:          "uploads/temp",
		RequireAuth:       true,
		ScanMalware:       true,
	}
}

// GetImageUploadConfig 取得圖片上傳配置
func GetImageUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		FileType:    "image",
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		AllowedMimeTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
			"image/bmp",
			"image/tiff",
		},
		AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tiff"},
		UploadPath:        "uploads/images",
		TempPath:          "uploads/temp",
		RequireAuth:       true,
		ScanMalware:       true,
	}
}

// GetGeneralUploadConfig 取得通用檔案上傳配置
func GetGeneralUploadConfig() *FileUploadConfig {
	return &FileUploadConfig{
		FileType:    "general",
		MaxFileSize: 100 * 1024 * 1024, // 100MB
		AllowedMimeTypes: []string{
			"image/jpeg", "image/png", "image/gif", "image/webp",
			"application/pdf", "application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"text/plain", "application/zip",
		},
		AllowedExtensions: []string{
			".jpg", ".jpeg", ".png", ".gif", ".webp",
			".pdf", ".doc", ".docx", ".txt", ".zip",
		},
		UploadPath:  "uploads/files",
		TempPath:    "uploads/temp",
		RequireAuth: true,
		ScanMalware: true,
	}
}
