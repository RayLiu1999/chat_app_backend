package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"chat_app_backend/app/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// slidingWindowScript 使用 Lua script 確保 Sliding Window 計數的原子性
// KEYS[1] = rate limit key（e.g. rate_limit:/login:127.0.0.1）
// ARGV[1] = 視窗大小（秒）
// ARGV[2] = 目前時間戳（毫秒）
// ARGV[3] = 允許的最大請求數
var slidingWindowScript = redis.NewScript(`
local key = KEYS[1]
local window = tonumber(ARGV[1]) * 1000
local now = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])

-- 移除視窗外的舊請求記錄
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- 計算目前視窗內的請求數
local count = redis.call('ZCARD', key)

if count >= limit then
    return 0
end

-- 記錄此次請求（score 與 member 皆使用時間戳，確保唯一性）
redis.call('ZADD', key, now, now)
redis.call('PEXPIRE', key, window)

return 1
`)

// RateLimiter 基於 Redis Sliding Window 演算法的限速中介軟體
// 參數：
//   - client: Redis 客戶端
//   - route: 用於區分不同路由的 key 前綴（e.g. "login"）
//   - limit: 時間視窗內允許的最大請求數
//   - window: 時間視窗大小（e.g. time.Minute）
//
// func RateLimiter(client *redis.Client, route string, limit int, window time.Duration, disable bool) gin.HandlerFunc {
func RateLimiter(client *redis.Client, route string, limit int, window time.Duration, disable bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if disable {
			c.Next()
			return
		}
		// 取得客戶端 IP
		ip := c.ClientIP()

		// 組合 Redis key
		key := fmt.Sprintf("rate_limit:%s:%s", route, ip)

		// 執行 Lua 腳本
		now := time.Now().UnixMilli()
		ctx := context.Background()
		result, err := slidingWindowScript.Run(
			ctx,
			client,
			[]string{key},
			int(window.Seconds()),
			now,
			limit,
		).Int()

		if err != nil {
			// Redis 錯誤時，放行請求以避免全面阻斷服務（fail-open 策略）
			c.Next()
			return
		}

		if result == 0 {
			// 超過限制，回傳 429
			c.Header("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, models.APIResponse{
				Status:  "error",
				Code:    models.ErrOperationFailed,
				Message: "請求過於頻繁，請稍後再試",
			})
			return
		}

		c.Next()
	}
}
