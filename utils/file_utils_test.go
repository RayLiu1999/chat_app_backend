package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupTestFile 建立一個用於測試的暫時檔案。
func setupTestFile(t *testing.T, content string) (path string, cleanup func()) {
	t.Helper()
	tmpFile, err := os.CreateTemp(t.TempDir(), "testfile-*.txt")
	assert.NoError(t, err)

	if content != "" {
		_, err = tmpFile.WriteString(content)
		assert.NoError(t, err)
	}
	tmpFile.Close()

	return tmpFile.Name(), func() {
		os.Remove(tmpFile.Name())
	}
}

func TestFileExists(t *testing.T) {
	filePath, cleanup := setupTestFile(t, "hello")
	defer cleanup()

	assert.True(t, FileExists(filePath), "檔案應該存在")
	assert.False(t, FileExists(filePath+".nonexistent"), "檔案不應該存在")
	assert.False(t, FileExists(t.TempDir()), "目錄不應該被視為檔案")
}

func TestReadFileContent(t *testing.T) {
	content := "hello world\nthis is a test"
	filePath, cleanup := setupTestFile(t, content)
	defer cleanup()

	readContent, err := ReadFileContent(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, readContent)

	_, err = ReadFileContent(filePath + ".nonexistent")
	assert.Error(t, err, "讀取不存在的檔案應該產生錯誤")
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name string
		size int64
		want string
	}{
		{"位元組", 500, "500 B"},
		{"千位元組", 1536, "1.5 KB"},
		{"百萬位元組", 1572864, "1.5 MB"},
		{"十億位元組", 1610612736, "1.5 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatFileSize(tt.size))
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir := t.TempDir()
	newDirPath := filepath.Join(tmpDir, "new_dir")

	// 目錄不存在，應該被建立
	err := EnsureDir(newDirPath)
	assert.NoError(t, err)
	info, err := os.Stat(newDirPath)
	assert.NoError(t, err)
	assert.True(t, info.IsDir(), "EnsureDir 應該建立一個目錄")

	// 目錄已存在，應該什麼都不做
	err = EnsureDir(newDirPath)
	assert.NoError(t, err)
}

func TestCopyFile(t *testing.T) {
	content := "copy this content"
	srcPath, cleanup := setupTestFile(t, content)
	defer cleanup()

	dstPath := filepath.Join(t.TempDir(), "destination.txt")

	bytesCopied, err := CopyFile(srcPath, dstPath)
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), bytesCopied)

	copiedContent, err := os.ReadFile(dstPath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))

	// 測試複製不存在的檔案
	_, err = CopyFile("nonexistent.txt", dstPath)
	assert.Error(t, err)
}
