package services

import (
	"chat_app_backend/config"
	"chat_app_backend/models"
	"chat_app_backend/providers"
	"chat_app_backend/repositories"
	"chat_app_backend/utils"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FileUploadServiceImpl 檔案上傳服務實現
type FileUploadServiceImpl struct {
	config       *config.Config
	fileProvider providers.FileProviderInterface
	odm          *providers.ODM
	fileRepo     repositories.FileRepositoryInterface
}

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

// NewFileUploadService 創建新的檔案上傳服務
func NewFileUploadService(cfg *config.Config, fileProvider providers.FileProviderInterface, odm *providers.ODM, fileRepo repositories.FileRepositoryInterface) FileUploadServiceInterface {
	return &FileUploadServiceImpl{
		config:       cfg,
		fileProvider: fileProvider,
		odm:          odm,
		fileRepo:     fileRepo,
	}
}

// UploadFileWithConfig 統一檔案上傳函數，使用配置參數
func (fs *FileUploadServiceImpl) UploadFileWithConfig(file multipart.File, header *multipart.FileHeader, userID string, config *models.FileUploadConfig) (*FileResult, error) {
	// 基本驗證
	if err := fs.ValidateFile(header); err != nil {
		return nil, fmt.Errorf("檔案驗證失敗: %w", err)
	}

	// 檔案大小檢查
	if header.Size > config.MaxFileSize {
		return nil, fmt.Errorf("檔案大小超過限制 %d bytes", config.MaxFileSize)
	}

	// 檢查副檔名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !slices.Contains(config.AllowedExtensions, ext) {
		return nil, fmt.Errorf("不支援的檔案格式: %s", ext)
	}

	// 檢查MIME類型
	mimeType := header.Header.Get("Content-Type")
	if !slices.Contains(config.AllowedMimeTypes, mimeType) {
		return nil, fmt.Errorf("不支援的檔案類型: %s", mimeType)
	}

	// 內容安全檢查
	if err := fs.CheckFileContent(file, header); err != nil {
		return nil, fmt.Errorf("檔案內容檢查失敗: %w", err)
	}

	// 生成安全的檔案名稱
	secureFileName := providers.GenerateSecureFileName(header.Filename, userID)

	relativePath := filepath.Join(config.FileType, secureFileName)

	// 儲存檔案
	fullPath, err := fs.fileProvider.SaveFile(file, relativePath)
	if err != nil {
		return nil, fmt.Errorf("檔案儲存失敗: %w", err)
	}

	utils.PrettyPrint("filepath", fullPath)

	// 生成檔案雜湊
	fileHash, err := providers.GenerateFileHash(file)
	if err != nil {
		// 清理已儲存的檔案
		fs.fileProvider.DeleteFile(fullPath)
		return nil, fmt.Errorf("檔案雜湊計算失敗: %w", err)
	}

	// 惡意軟體掃描（如果配置要求）
	if config.ScanMalware {
		if err := fs.ScanFileForMalware(fullPath); err != nil {
			fs.fileProvider.DeleteFile(fullPath)
			return nil, fmt.Errorf("檔案安全掃描失敗: %w", err)
		}
	}

	// 創建資料庫記錄
	userObjID, _ := primitive.ObjectIDFromHex(userID)
	uploadedFile := &models.UploadedFile{
		UserID:       userObjID,
		OriginalName: header.Filename,
		FileName:     secureFileName,
		FilePath:     fullPath,
		FileSize:     header.Size,
		MimeType:     mimeType,
		FileType:     config.FileType,
		Status:       "verified",
		Hash:         fileHash,
	}

	if err := fs.fileRepo.CreateFile(uploadedFile); err != nil {
		fs.fileProvider.DeleteFile(fullPath)
		return nil, fmt.Errorf("資料庫記錄創建失敗: %w", err)
	}

	return &FileResult{
		ID:         uploadedFile.BaseModel.GetID(),
		FileName:   secureFileName,
		FilePath:   fullPath,
		FileURL:    fs.fileProvider.GetFileURL(fullPath),
		FileSize:   header.Size,
		MimeType:   mimeType,
		UploadedAt: time.Now().Unix(),
		UserID:     userID,
	}, nil
}

// UploadFile 通用檔案上傳 - 使用統一函數
func (fs *FileUploadServiceImpl) UploadFile(file multipart.File, header *multipart.FileHeader, userID string) (*FileResult, error) {
	config := models.GetGeneralUploadConfig()
	return fs.UploadFileWithConfig(file, header, userID, config)
}

// UploadAvatar 頭像上傳 - 使用統一函數
func (fs *FileUploadServiceImpl) UploadAvatar(file multipart.File, header *multipart.FileHeader, userID string) (*FileResult, error) {
	config := models.GetAvatarUploadConfig()
	return fs.UploadFileWithConfig(file, header, userID, config)
}

// UploadDocument 文件上傳 - 使用統一函數
func (fs *FileUploadServiceImpl) UploadDocument(file multipart.File, header *multipart.FileHeader, userID string) (*FileResult, error) {
	config := models.GetDocumentUploadConfig()
	return fs.UploadFileWithConfig(file, header, userID, config)
}

// ValidateFile 基本檔案驗證
func (fs *FileUploadServiceImpl) ValidateFile(header *multipart.FileHeader) error {
	// 檢查檔案名稱
	if header.Filename == "" {
		return fmt.Errorf("檔案名稱不能為空")
	}

	// 檢查檔案名稱長度
	if len(header.Filename) > 255 {
		return fmt.Errorf("檔案名稱過長")
	}

	// 檢查檔案大小
	if header.Size <= 0 {
		return fmt.Errorf("檔案大小無效")
	}

	// 檢查檔案名稱中的危險字符
	dangerousChars := []string{"../", "..\\", "<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range dangerousChars {
		if strings.Contains(header.Filename, char) {
			return fmt.Errorf("檔案名稱包含不允許的字符: %s", char)
		}
	}

	// 檢查副檔名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		return fmt.Errorf("檔案必須有副檔名")
	}

	// 檢查危險的副檔名
	dangerousExts := []string{".exe", ".bat", ".cmd", ".scr", ".pif", ".com", ".vbs", ".js", ".jar", ".sh"}
	if slices.Contains(dangerousExts, ext) {
		return fmt.Errorf("不允許的檔案類型: %s", ext)
	}

	return nil
}

// ValidateImage 圖片檔案驗證
func (fs *FileUploadServiceImpl) ValidateImage(header *multipart.FileHeader) error {
	if err := fs.ValidateFile(header); err != nil {
		return err
	}

	config := models.GetImageUploadConfig()

	// 檢查檔案大小
	if header.Size > config.MaxFileSize {
		return fmt.Errorf("圖片檔案大小超過限制 %d bytes", config.MaxFileSize)
	}

	// 檢查副檔名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !slices.Contains(config.AllowedExtensions, ext) {
		return fmt.Errorf("不支援的圖片格式: %s", ext)
	}

	// 檢查MIME類型
	mimeType := header.Header.Get("Content-Type")
	if !slices.Contains(config.AllowedMimeTypes, mimeType) {
		return fmt.Errorf("不支援的圖片類型: %s", mimeType)
	}

	return nil
}

// ValidateDocument 文件檔案驗證
func (fs *FileUploadServiceImpl) ValidateDocument(header *multipart.FileHeader) error {
	if err := fs.ValidateFile(header); err != nil {
		return err
	}

	config := models.GetDocumentUploadConfig()

	// 檢查檔案大小
	if header.Size > config.MaxFileSize {
		return fmt.Errorf("文件大小超過限制 %d bytes", config.MaxFileSize)
	}

	// 檢查副檔名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !slices.Contains(config.AllowedExtensions, ext) {
		return fmt.Errorf("不支援的文件格式: %s", ext)
	}

	// 檢查MIME類型
	mimeType := header.Header.Get("Content-Type")
	if !slices.Contains(config.AllowedMimeTypes, mimeType) {
		return fmt.Errorf("不支援的文件類型: %s", mimeType)
	}

	return nil
}

// CheckFileContent 檢查檔案內容安全性
func (fs *FileUploadServiceImpl) CheckFileContent(file multipart.File, header *multipart.FileHeader) error {
	// 重置檔案指針
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("無法重置檔案指針: %w", err)
	}

	// 讀取檔案前512位元組進行MIME類型檢測
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("無法讀取檔案內容: %w", err)
	}

	// 重置檔案指針
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("無法重置檔案指針: %w", err)
	}

	// 檢測真實的MIME類型
	detectedMimeType := http.DetectContentType(buffer[:n])
	declaredMimeType := header.Header.Get("Content-Type")

	// 驗證MIME類型一致性（允許一些變化）
	if !fs.isMimeTypeCompatible(detectedMimeType, declaredMimeType) {
		return fmt.Errorf("檔案實際類型 (%s) 與聲明類型 (%s) 不符", detectedMimeType, declaredMimeType)
	}

	// 檢查是否包含惡意內容標誌
	if fs.containsMaliciousContent(buffer[:n]) {
		return fmt.Errorf("檔案包含可疑內容")
	}

	return nil
}

// isMimeTypeCompatible 檢查MIME類型相容性
func (fs *FileUploadServiceImpl) isMimeTypeCompatible(detected, declared string) bool {
	// 如果完全相同
	if detected == declared {
		return true
	}

	// 一些MIME類型的相容性對應
	compatibleTypes := map[string][]string{
		"image/jpeg": {"image/jpeg", "image/jpg"},
		"image/jpg":  {"image/jpeg", "image/jpg"},
		"image/png":  {"image/png"},
		"image/gif":  {"image/gif"},
		"image/webp": {"image/webp"},
		"text/plain": {"text/plain", "application/octet-stream"},
	}

	if allowed, exists := compatibleTypes[declared]; exists {
		return slices.Contains(allowed, detected)
	}

	return false
}

// containsMaliciousContent 檢查是否包含惡意內容
func (fs *FileUploadServiceImpl) containsMaliciousContent(content []byte) bool {
	// 檢查常見的惡意軟體標誌
	maliciousSignatures := [][]byte{
		[]byte("MZ"),             // PE執行檔標誌
		[]byte("<!DOCTYPE html"), // HTML檔案
		[]byte("<script"),        // JavaScript
		[]byte("<?php"),          // PHP腳本
		[]byte("<%"),             // ASP/JSP
		[]byte("#!/bin/sh"),      // Shell腳本
		[]byte("#!/bin/bash"),    // Bash腳本
		[]byte("\x7fELF"),        // ELF執行檔（Linux）
		[]byte("PK\x03\x04"),     // ZIP檔案（可能包含惡意軟體）
	}

	contentStr := strings.ToLower(string(content))

	// 檢查二進位標誌
	for _, signature := range maliciousSignatures {
		if len(content) >= len(signature) {
			for i := 0; i <= len(content)-len(signature); i++ {
				match := true
				for j := 0; j < len(signature); j++ {
					if content[i+j] != signature[j] {
						match = false
						break
					}
				}
				if match {
					return true
				}
			}
		}
	}

	// 檢查可疑的文字內容
	suspiciousStrings := []string{
		"eval(",
		"base64_decode",
		"shell_exec",
		"system(",
		"exec(",
		"passthru(",
		"file_get_contents",
		"fopen(",
		"javascript:",
		"vbscript:",
	}

	for _, suspicious := range suspiciousStrings {
		if strings.Contains(contentStr, suspicious) {
			return true
		}
	}

	return false
}

// ScanFileForMalware 惡意軟體掃描（簡化實現）
func (fs *FileUploadServiceImpl) ScanFileForMalware(filePath string) error {
	// 這裡可以整合第三方防毒引擎，如 ClamAV
	// 目前實現基本的檔案檢查

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("無法開啟檔案進行掃描: %w", err)
	}
	defer file.Close()

	// 讀取檔案開頭進行基本檢查
	buffer := make([]byte, 1024)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Errorf("無法讀取檔案進行掃描: %w", err)
	}

	// 檢查是否包含惡意內容
	if fs.containsMaliciousContent(buffer[:n]) {
		return fmt.Errorf("檔案掃描發現可疑內容")
	}

	return nil
}

// DeleteFile 刪除檔案
func (fs *FileUploadServiceImpl) DeleteFile(filePath string) error {
	// 從資料庫刪除記錄
	if err := fs.fileRepo.DeleteFileByPath(filePath); err != nil {
		return fmt.Errorf("刪除檔案記錄失敗: %w", err)
	}

	// 從檔案系統刪除檔案
	if err := fs.fileProvider.DeleteFile(filePath); err != nil {
		return fmt.Errorf("刪除檔案失敗: %w", err)
	}

	return nil
}

// GetFileInfo 取得檔案資訊
func (fs *FileUploadServiceImpl) GetFileInfo(filePath string) (*FileInfo, error) {
	// 從資料庫取得檔案資訊
	uploadedFile, err := fs.fileRepo.GetFileByPath(filePath)
	if err != nil {
		return nil, fmt.Errorf("取得檔案記錄失敗: %w", err)
	}

	// 從檔案系統取得檔案資訊
	fileInfo, err := fs.fileProvider.GetFileInfo(filePath)
	if err != nil {
		return nil, fmt.Errorf("取得檔案系統資訊失敗: %w", err)
	}

	return &FileInfo{
		ID:         uploadedFile.BaseModel.GetID(),
		FileName:   uploadedFile.FileName,
		FilePath:   uploadedFile.FilePath,
		FileSize:   fileInfo.Size(),
		MimeType:   uploadedFile.MimeType,
		CreatedAt:  uploadedFile.BaseModel.CreatedAt.Unix(),
		ModifiedAt: fileInfo.ModTime().Unix(),
	}, nil
}

// CleanupExpiredFiles 清理過期檔案
func (fs *FileUploadServiceImpl) CleanupExpiredFiles() error {
	// 取得過期的檔案列表
	expiredFiles, err := fs.fileRepo.GetExpiredFiles()
	if err != nil {
		return fmt.Errorf("取得過期檔案列表失敗: %w", err)
	}

	// 刪除過期檔案
	for _, file := range expiredFiles {
		if err := fs.DeleteFile(file.FilePath); err != nil {
			// 記錄錯誤但繼續處理其他檔案
			fmt.Printf("刪除過期檔案失敗 %s: %v\n", file.FilePath, err)
		}
	}

	return nil
}

// GetUserFiles 獲取用戶的檔案列表
func (fs *FileUploadServiceImpl) GetUserFiles(userID string) ([]*models.UploadedFile, error) {
	if userID == "" {
		return nil, fmt.Errorf("用戶ID不能為空")
	}

	files, err := fs.fileRepo.GetFilesByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 轉換為指針切片
	var result []*models.UploadedFile
	for i := range files {
		result = append(result, &files[i])
	}

	return result, nil
}

// DeleteFileByID 根據檔案ID刪除檔案
func (fs *FileUploadServiceImpl) DeleteFileByID(fileID string, userID string) error {
	if fileID == "" {
		return fmt.Errorf("檔案ID不能為空")
	}
	if userID == "" {
		return fmt.Errorf("用戶ID不能為空")
	}

	// 獲取檔案信息
	file, err := fs.fileRepo.GetFileByID(fileID)
	if err != nil {
		return fmt.Errorf("檔案不存在: %v", err)
	}

	// 檢查權限 - 用戶只能刪除自己的檔案
	if file.UserID.Hex() != userID {
		return fmt.Errorf("沒有權限刪除此檔案")
	}

	// 從檔案系統刪除檔案
	if err := fs.fileProvider.DeleteFile(file.FilePath); err != nil {
		return fmt.Errorf("刪除檔案失敗: %v", err)
	}

	// 從資料庫刪除記錄
	return fs.fileRepo.DeleteFileByID(fileID)
}

// GetFileURLByID 根據檔案ID獲取檔案連結
func (fs *FileUploadServiceImpl) GetFileURLByID(fileID string) (string, error) {
	if fileID == "" {
		return "", fmt.Errorf("檔案ID不能為空")
	}

	// 從資料庫獲取檔案記錄
	file, err := fs.fileRepo.GetFileByID(fileID)
	if err != nil {
		return "", fmt.Errorf("檔案不存在: %v", err)
	}

	// 檢查檔案狀態
	if file.Status != "verified" {
		return "", fmt.Errorf("檔案尚未驗證或已損壞")
	}

	// 檢查檔案是否存在
	if _, err := fs.fileProvider.GetFileInfo(file.FilePath); err != nil {
		return "", fmt.Errorf("檔案不存在或已被刪除: %v", err)
	}

	// 返回檔案URL
	return fs.fileProvider.GetFileURL(file.FilePath), nil
}

// GetFileInfoByID 根據檔案ID獲取完整檔案資訊
func (fs *FileUploadServiceImpl) GetFileInfoByID(fileID string) (*models.UploadedFile, error) {
	if fileID == "" {
		return nil, fmt.Errorf("檔案ID不能為空")
	}

	// 從資料庫獲取檔案記錄
	file, err := fs.fileRepo.GetFileByID(fileID)
	if err != nil {
		return nil, fmt.Errorf("檔案不存在: %v", err)
	}

	return file, nil
}
