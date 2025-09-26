package utils

import "fmt"

// UserProfileCacheKey 生成用戶個人資料的快取鍵
func UserProfileCacheKey(userID string) string {
	return fmt.Sprintf("user:%s:profile", userID)
}

// UserStatusCacheKey 生成用戶在線狀態的快取鍵
func UserStatusCacheKey(userID string) string {
	return fmt.Sprintf("user:%s:status", userID)
}

// UserActivityThrottleCacheKey 生成用戶活動更新的節流閥快取鍵
func UserActivityThrottleCacheKey(userID string) string {
	return fmt.Sprintf("user:%s:active:throttle", userID)
}
