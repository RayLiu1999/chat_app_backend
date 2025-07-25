package providers

import (
	"mime/multipart"
	"os"
)

// FileProviderInterface - 負責底層文件操作
type FileProviderInterface interface {
	SaveFile(file multipart.File, filename string) (string, error)
	DeleteFile(filepath string) error
	GetFileInfo(filepath string) (os.FileInfo, error)
	GetFileURL(filePath string) string
}
