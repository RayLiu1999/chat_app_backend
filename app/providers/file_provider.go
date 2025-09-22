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

// FileProvider 本地檔案系統提供者
type FileProvider struct {
	cfg *config.Config
}

// NewFileProvider 創建新的本地檔案提供者
func NewFileProvider(cfg *config.Config) FileProviderInterface {
	return &FileProvider{
		cfg: cfg,
	}
}

// SaveFile 儲存檔案到本地檔案系統
func (fp *FileProvider) SaveFile(file multipart.File, filename string) (string, error) {
	// 確保目錄存在
	dir := filepath.Dir(filepath.Join(BaseUploadPath, filename))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("無法創建目錄: %w", err)
	}

	// 創建目標檔案
	fullPath := filepath.Join(BaseUploadPath, filename)
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

	return fullPath, nil
}

// DeleteFile 刪除檔案
func (fp *FileProvider) DeleteFile(filepath string) error {
	// 確保檔案路徑在允許的基礎路徑內（防止路徑遍歷攻擊）
	if !strings.HasPrefix(filepath, BaseUploadPath) {
		return fmt.Errorf("檔案路徑不在允許範圍內")
	}

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("檔案不存在: %s", filepath)
		}
		return fmt.Errorf("無法刪除檔案: %w", err)
	}

	return nil
}

// GetFileInfo 取得檔案資訊
func (fp *FileProvider) GetFileInfo(filepath string) (os.FileInfo, error) {
	// 確保檔案路徑在允許的基礎路徑內
	if !strings.HasPrefix(filepath, BaseUploadPath) {
		return nil, fmt.Errorf("檔案路徑不在允許範圍內")
	}

	info, err := os.Stat(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("檔案不存在: %s", filepath)
		}
		return nil, fmt.Errorf("無法取得檔案資訊: %w", err)
	}

	return info, nil
}

// GetFileURL 生成檔案的URL
func (fp *FileProvider) GetFileURL(filePath string) string {
	baseURL := fp.cfg.Server.BaseURL
	if (baseURL == "" || baseURL == "http://localhost") && fp.cfg.Server.Port != "" {
		baseURL = fmt.Sprintf("http://localhost:%s", fp.cfg.Server.Port)
	}

	// 確保路徑格式正確
	if !strings.HasPrefix(filePath, BaseUploadPath) {
		filePath = filepath.Join(BaseUploadPath, filePath)
	}

	// 返回完整的檔案URL
	return fmt.Sprintf("%s/%s", baseURL, filePath)
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
