package providers

import (
	"chat_app_backend/config"
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

// RedisWrapper 結構體封裝了 Redis 客戶端
type RedisWrapper struct {
	Client *redis.Client
	cb     *gobreaker.CircuitBreaker // 熔斷器：防止 Redis 不穩定時雪崩
}

// Execute 使用熔斷器包裹 Redis 操作
// 若熔斷器處於開啟狀態，fn 不會被執行，直接回傳服務不可用錯誤
func (rw *RedisWrapper) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return ExecuteWithBreaker(rw.cb, fn)
}

// NewRedisClient 創建並返回一個新的 Redis 客戶端實例
func NewRedisClient(cfg *config.Config) (*RedisWrapper, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       0, // use default DB
	})

	// 測試連線
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		slog.Error("Redis連線失敗", "error", err)
		return nil, err
	}

	slog.Info("Redis連線成功")
	return &RedisWrapper{
		Client: redisClient,
		cb:     newCircuitBreaker("redis"),
	}, nil
}

// Ping 測試 Redis 連線
func (rw *RedisWrapper) Ping() error {
	_, err := rw.Client.Ping(context.Background()).Result()
	return err
}

// Close 關閉 Redis 連線
func (rw *RedisWrapper) Close() {
	if rw.Client != nil {
		if err := rw.Client.Close(); err != nil {
			slog.Warn("無法關閉 Redis 連線", "error", err)
		}
	}
}
