package services

import (
	"chat_app_backend/utils"
	"log"
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
// 每隔指定時間檢查並設置超時未活動的用戶為離線狀態
func (bt *BackgroundTasks) StartOfflineUserChecker(intervalMinutes, offlineThresholdMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	utils.Log.Info("離線用戶檢查任務已啟動", "interval_minutes", intervalMinutes, "threshold_minutes", offlineThresholdMinutes)

	for range ticker.C {
		if err := bt.userService.CheckAndSetOfflineUsers(offlineThresholdMinutes); err != nil {
			utils.Log.Error("離線用戶檢查失敗", "error", err)
		}
	}
}

// StartAllBackgroundTasks 啟動所有後台任務
func (bt *BackgroundTasks) StartAllBackgroundTasks() {
	// 啟動離線用戶檢查任務 - 每5分鐘檢查一次，15分鐘未活動則設為離線
	go bt.StartOfflineUserChecker(5, 15)

	log.Println("所有後台任務已啟動")
}
