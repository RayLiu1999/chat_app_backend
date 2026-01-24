package utils

import (
	"log/slog"
	"os"
)

var Log *slog.Logger

// InitLogger 初始化日誌系統
// 根據環境變數 GO_ENV 決定輸出格式
// production: JSON 格式, Info Level
// development: Text 格式, Debug Level
func InitLogger() {
	env := os.Getenv("GO_ENV")

	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if env == "production" {
		// 生產環境使用 JSON 格式
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// 開發環境使用文字格式，並顯示 Debug 訊息
		opts.Level = slog.LevelDebug
		// opts.AddSource = true // 開發時可開啟顯示原始碼位置，但會影響效能
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	Log = slog.New(handler)

	// 設置為預設 logger
	slog.SetDefault(Log)
}

// 確保在包初始化時就可用，但建議在 main 中再次調用 InitLogger 以正確讀取 ENV
func init() {
	InitLogger()
}
