package providers

import (
	"chat_app_backend/config"
	"crypto/sha256"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BaseUploadPath 上傳檔案的基礎路徑
const BaseUploadPath = "uploads/"

// fileProvider 本地檔案系統提供者
type fileProvider struct {
	cfg *config.Config
}

// NewFileProvider 創建新的本地檔案提供者
func NewFileProvider(cfg *config.Config) *fileProvider {
	return &fileProvider{
		cfg: cfg,
	}
}

// SaveFile 儲存檔案到本地檔案系統
func (fp *fileProvider) SaveFile(file multipart.File, filename string) (string, error) {
	// fullPath 是實際存在檔案系統上的絕對路徑或相對於專案目錄的完整路徑
	fullPath := filepath.Join(BaseUploadPath, filename)

	// 確保目錄存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("無法創建目錄: %w", err)
	}

	// 創建目標檔案
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("無法創建檔案: %w", err)
	}
	defer dst.Close()

	// 重置檔案指針到開始
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("無法重置檔案指針: %w", err)
	}

	// 複製檔案內容
	if _, err := io.Copy(dst, file); err != nil {
		// 如果複製失敗，清理部分寫入的檔案
		os.Remove(fullPath)
		return "", fmt.Errorf("無法複製檔案內容: %w", err)
	}

	// 回傳相對路徑，儲存於資料庫
	return filename, nil
}

// DeleteFile 刪除檔案
func (fp *fileProvider) DeleteFile(filepathStr string) error {
	// 若傳入的是相對路徑，先補上 BaseUploadPath
	fullPath := filepathStr
	if !strings.HasPrefix(fullPath, BaseUploadPath) {
		fullPath = filepath.Join(BaseUploadPath, filepathStr)
	}

	// 確保檔案路徑在允許的基礎路徑內（防止路徑遍歷攻擊）
	if !strings.HasPrefix(fullPath, BaseUploadPath) {
		return fmt.Errorf("檔案路徑不在允許範圍內")
	}

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("檔案不存在: %s", fullPath)
		}
		return fmt.Errorf("無法刪除檔案: %w", err)
	}

	return nil
}

// GetFileInfo 取得檔案資訊
func (fp *fileProvider) GetFileInfo(filepathStr string) (os.FileInfo, error) {
	// 若傳入的是相對路徑，先補上 BaseUploadPath
	fullPath := filepathStr
	if !strings.HasPrefix(fullPath, BaseUploadPath) {
		fullPath = filepath.Join(BaseUploadPath, filepathStr)
	}

	// 確保檔案路徑在允許的基礎路徑內
	if !strings.HasPrefix(fullPath, BaseUploadPath) {
		return nil, fmt.Errorf("檔案路徑不在允許範圍內")
	}

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("檔案不存在: %s", fullPath)
		}
		return nil, fmt.Errorf("無法取得檔案資訊: %w", err)
	}

	return info, nil
}

// GetFileURL 生成檔案的URL
func (fp *fileProvider) GetFileURL(filePath string) string {
	baseURL := fp.cfg.Server.BaseURL
	// 移除末尾的斜線
	baseURL = strings.TrimRight(baseURL, "/")

	if (baseURL == "" || baseURL == "http://localhost") && fp.cfg.Server.Port != "" {
		baseURL = fmt.Sprintf("http://localhost:%s", fp.cfg.Server.Port)
	}

	// 將反斜線替換為正斜線 (處理 Windows 路徑)
	cleanFilePath := filepath.ToSlash(filePath)

	// 如果已經包含 uploads/ 前綴，確保不會重複
	if !strings.HasPrefix(cleanFilePath, filepath.ToSlash(BaseUploadPath)) {
		cleanFilePath = filepath.ToSlash(filepath.Join(BaseUploadPath, cleanFilePath))
	}

	// 返回完整的檔案URL
	return fmt.Sprintf("%s/%s", baseURL, cleanFilePath)
}

// GetFile 取得檔案內容
func (fp *fileProvider) GetFile(filepathStr string) (io.ReadCloser, error) {
	// 若傳入的是相對路徑，先補上 BaseUploadPath
	fullPath := filepathStr
	if !strings.HasPrefix(fullPath, BaseUploadPath) {
		fullPath = filepath.Join(BaseUploadPath, filepathStr)
	}

	// 確保檔案路徑在允許的基礎路徑內
	if !strings.HasPrefix(fullPath, BaseUploadPath) {
		return nil, fmt.Errorf("檔案路徑不在允許範圍內")
	}

	return os.Open(fullPath)
}

// GenerateSecureFileName 生成安全的檔案名稱
func GenerateSecureFileName(originalName, userID string) string {
	// 取得副檔名
	ext := strings.ToLower(filepath.Ext(originalName))

	// 生成時間戳
	timestamp := time.Now().Unix()

	// 生成隨機雜湊
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s_%d_%s", userID, timestamp, originalName)))
	hash := fmt.Sprintf("%x", h.Sum(nil))[:16]

	return fmt.Sprintf("%d_%s%s", timestamp, hash, ext)
}

// GenerateFileHash 生成檔案SHA256雜湊值
func GenerateFileHash(file multipart.File) (string, error) {
	// 重置檔案指針到開始
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("無法重置檔案指針: %w", err)
	}

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", fmt.Errorf("無法計算檔案雜湊: %w", err)
	}

	// 重置檔案指針到開始以供後續使用
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("無法重置檔案指針: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// ValidateFilePath 驗證檔案路徑安全性
func ValidateFilePath(basePath, filePath string) error {
	// 清理路徑
	cleanPath := filepath.Clean(filePath)
	cleanBasePath := filepath.Clean(basePath)

	// 檢查是否包含相對路徑組件
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("檔案路徑包含無效字符")
	}

	// 檢查絕對路徑是否在基礎路徑內
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("無法解析檔案路徑: %w", err)
	}

	absBasePath, err := filepath.Abs(cleanBasePath)
	if err != nil {
		return fmt.Errorf("無法解析基礎路徑: %w", err)
	}

	if !strings.HasPrefix(absPath, absBasePath) {
		return fmt.Errorf("檔案路徑超出允許範圍")
	}

	return nil
}

// CreateTempFile 創建臨時檔案
func CreateTempFile(file multipart.File, tempDir string) (string, error) {
	// 確保臨時目錄存在
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("無法創建臨時目錄: %w", err)
	}

	// 創建臨時檔案
	tempFile, err := os.CreateTemp(tempDir, "upload_*")
	if err != nil {
		return "", fmt.Errorf("無法創建臨時檔案: %w", err)
	}
	defer tempFile.Close()

	// 重置檔案指針
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("無法重置檔案指針: %w", err)
	}

	// 複製內容到臨時檔案
	if _, err := io.Copy(tempFile, file); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("無法複製檔案到臨時位置: %w", err)
	}

	// 重置原檔案指針
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("無法重置原檔案指針: %w", err)
	}

	return tempFile.Name(), nil
}
