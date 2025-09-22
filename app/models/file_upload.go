package models

import "go.mongodb.org/mongo-driver/bson/primitive"

// FileResult 上傳結果結構
type FileResult struct {
	ID         primitive.ObjectID `json:"id"`
	FileName   string             `json:"file_name"`
	FilePath   string             `json:"file_path"`
	FileURL    string             `json:"file_url"`
	FileSize   int64              `json:"file_size"`
	MimeType   string             `json:"mime_type"`
	UploadedAt int64              `json:"uploaded_at"`
	UserID     string             `json:"user_id"`
}

// FileInfo 檔案資訊結構
type FileInfo struct {
	ID         primitive.ObjectID `json:"id"`
	FileName   string             `json:"file_name"`
	FilePath   string             `json:"file_path"`
	FileSize   int64              `json:"file_size"`
	MimeType   string             `json:"mime_type"`
	CreatedAt  int64              `json:"created_at"`
	ModifiedAt int64              `json:"modified_at"`
}
