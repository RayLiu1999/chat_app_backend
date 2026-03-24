package services

import (
	"context"
	"log"
	"log/slog"
	"time"
)

// BackgroundTasks 管理後台任務
type BackgroundTasks struct {
	userService UserService
}

// NewBackgroundTasks 創建後台任務管理器
func NewBackgroundTasks(userService UserService) *BackgroundTasks {
	return &BackgroundTasks{
		userService: userService,
	}
}

// StartOfflineUserChecker 啟動離線用戶檢查任務
// 每隔指定時間檢查並設置超時未活動的用戶為離線狀態 (支援優雅停機)
func (bt *BackgroundTasks) StartOfflineUserChecker(ctx context.Context, intervalMinutes, offlineThresholdMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	slog.Info("離線用戶檢查任務已啟動", "interval_minutes", intervalMinutes, "threshold_minutes", offlineThresholdMinutes)

	for {
		select {
		case <-ctx.Done():
			slog.Info("收到關閉信號，停止離線用戶檢查任務")
			return
		case <-ticker.C:
			if err := bt.userService.CheckAndSetOfflineUsers(offlineThresholdMinutes); err != nil {
				slog.Error("離線用戶檢查失敗", "error", err)
			}
		}
	}
}

// StartAllBackgroundTasks 啟動所有後台任務 (支援優雅停機)
func (bt *BackgroundTasks) StartAllBackgroundTasks(ctx context.Context) {
	// 啟動離線用戶檢查任務 - 每5分鐘檢查一次，15分鐘未活動則設為離線
	go bt.StartOfflineUserChecker(ctx, 5, 15)

	// 啟動過期令牌清理任務 - 每10分鐘檢查一次
	go bt.StartExpiredTokenCleaner(ctx, 10)

	log.Println("所有後台任務已啟動")
}

// StartExpiredTokenCleaner 啟動過期令牌清理任務
func (bt *BackgroundTasks) StartExpiredTokenCleaner(ctx context.Context, intervalMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	slog.Info("過期令牌清理任務已啟動", "interval_minutes", intervalMinutes)

	for {
		select {
		case <-ctx.Done():
			slog.Info("收到關閉信號，停止過期令牌清理任務")
			return
		case <-ticker.C:
			if err := bt.userService.ClearExpiredRefreshTokens(); err != nil {
				slog.Error("清理過期令牌失敗", "error", err)
			}
		}
	}
}
