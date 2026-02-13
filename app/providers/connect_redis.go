package providers

import (
	"chat_app_backend/config"
	"context"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// RedisWrapper 結構體封裝了 Redis 客戶端
type RedisWrapper struct {
	Client *redis.Client
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
	return &RedisWrapper{Client: redisClient}, nil
}

// Ping 測試 Redis 連線
func (rw *RedisWrapper) Ping() error {
	_, err := rw.Client.Ping(context.Background()).Result()
	return err
}

// Close 關閉 Redis 連線
func (rw *RedisWrapper) Close() {
	if rw.Client != nil {
		rw.Client.Close()
	}
}
