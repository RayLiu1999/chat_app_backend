package services

import (
	"log"
	"time"
)

// BackgroundTasks 管理後台任務
type BackgroundTasks struct {
	userService UserServiceInterface
}

// NewBackgroundTasks 創建後台任務管理器
func NewBackgroundTasks(userService UserServiceInterface) *BackgroundTasks {
	return &BackgroundTasks{
		userService: userService,
	}
}

// StartOfflineUserChecker 啟動離線用戶檢查任務
// 每隔指定時間檢查並設置超時未活動的用戶為離線狀態
func (bt *BackgroundTasks) StartOfflineUserChecker(intervalMinutes, offlineThresholdMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	log.Printf("Started offline user checker: checking every %d minutes, threshold %d minutes", intervalMinutes, offlineThresholdMinutes)

	for range ticker.C {
		if err := bt.userService.CheckAndSetOfflineUsers(offlineThresholdMinutes); err != nil {
			log.Printf("Error checking offline users: %v", err)
		} else {
			log.Printf("Offline user check completed")
		}
	}
}

// StartAllBackgroundTasks 啟動所有後台任務
func (bt *BackgroundTasks) StartAllBackgroundTasks() {
	// 啟動離線用戶檢查任務 - 每5分鐘檢查一次，15分鐘未活動則設為離線
	go bt.StartOfflineUserChecker(5, 15)

	log.Println("All background tasks started")
}
