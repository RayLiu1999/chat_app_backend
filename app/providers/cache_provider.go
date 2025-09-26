package providers

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ICacheProvider 定義了快取提供者的介面
type CacheProvider interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Delete(key string) error
}

// RedisCacheProvider 是 ICacheProvider 的 Redis 實作
type RedisCacheProvider struct {
	client *redis.Client
}

// NewRedisCacheProvider 建立一個新的 RedisCacheProvider
func NewRedisCacheProvider(client *redis.Client) *RedisCacheProvider {
	return &RedisCacheProvider{
		client: client,
	}
}

// Get 從 Redis 獲取一個值
func (p *RedisCacheProvider) Get(key string) (string, error) {
	val, err := p.client.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil // 在 Redis v9 中，應使用 errors.Is 來檢查 redis.Nil
		}
		return "", err
	}
	return val, nil
}

// Set 將一個值存入 Redis
func (p *RedisCacheProvider) Set(key string, value string, expiration time.Duration) error {
	return p.client.Set(context.Background(), key, value, expiration).Err()
}

// Delete 從 Redis 刪除一個值
func (p *RedisCacheProvider) Delete(key string) error {
	return p.client.Del(context.Background(), key).Err()
}
