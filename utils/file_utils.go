package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileExists 檢查指定路徑的檔案是否存在
// 參數：
//   - path: 檔案路徑
//
// 返回：
//   - 檔案存在返回 true，否則返回 false
func FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// ReadFileContent 讀取整個檔案內容為字串
// 參數：
//   - path: 檔案路徑
//
// 返回：
//   - 檔案內容和錯誤信息
func ReadFileContent(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatFileSize 將檔案大小（字節）轉換為可讀的格式
// 參數：
//   - size: 檔案大小（字節）
//
// 返回：
//   - 可讀的檔案大小字串，如 "1.2 MB"
func FormatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// EnsureDir 確保目錄存在，如果不存在則創建
// 參數：
//   - dirPath: 目錄路徑
//   - perm: 權限設置，默認為 0755
//
// 返回：
//   - 錯誤訊息
func EnsureDir(dirPath string, perm ...os.FileMode) error {
	permission := os.FileMode(0755)
	if len(perm) > 0 {
		permission = perm[0]
	}

	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, permission)
	}

	return nil
}

// CopyFile 複製檔案內容從源路徑到目標路徑
// 參數：
//   - src: 源檔案路徑
//   - dst: 目標檔案路徑
//
// 返回：
//   - 複製的位元組數和錯誤訊息
func CopyFile(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	// 確保目標目錄存在
	dstDir := filepath.Dir(dst)
	if err := EnsureDir(dstDir); err != nil {
		return 0, err
	}

	destinationFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destinationFile.Close()

	return io.Copy(destinationFile, sourceFile)
}
