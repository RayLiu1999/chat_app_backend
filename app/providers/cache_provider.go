package providers

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheProvider 定義了快取提供者的介面
type CacheProvider interface {
	Get(key string) (string, error)
	Set(key string, value string, expiration time.Duration) error
	Delete(key string) error
	SetNX(key string, value string, expiration time.Duration) (bool, error)
}

// RedisCacheProvider 是 CacheProvider 的 Redis 實作
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
	if p.client == nil {
		return "", nil
	}
	val, err := p.client.Get(context.Background(), key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", err
	}
	return val, nil
}

// Set 將一個值存入 Redis
func (p *RedisCacheProvider) Set(key string, value string, expiration time.Duration) error {
	if p.client == nil {
		return nil
	}
	return p.client.Set(context.Background(), key, value, expiration).Err()
}

// Delete 從 Redis 刪除一個值
func (p *RedisCacheProvider) Delete(key string) error {
	if p.client == nil {
		return nil
	}
	return p.client.Del(context.Background(), key).Err()
}

// SetNX 將一個值存入 Redis，僅當該值不存在時才執行 (Set if Not eXists)
func (p *RedisCacheProvider) SetNX(key string, value string, expiration time.Duration) (bool, error) {
	if p.client == nil {
		return true, nil
	}
	//nolint:staticcheck // SetNX is still functional and widely used in this project style
	return p.client.SetNX(context.Background(), key, value, expiration).Result()
}

// InMemoryCacheProvider 是 CacheProvider 的本地記憶體實作
type InMemoryCacheProvider struct {
	data sync.Map
}

type cacheItem struct {
	value      string
	expiration int64
}

// NewInMemoryCacheProvider 建立一個新的 InMemoryCacheProvider
func NewInMemoryCacheProvider() *InMemoryCacheProvider {
	return &InMemoryCacheProvider{}
}

func (p *InMemoryCacheProvider) Get(key string) (string, error) {
	val, ok := p.data.Load(key)
	if !ok {
		return "", nil
	}

	item := val.(cacheItem)
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		p.data.Delete(key)
		return "", nil
	}

	return item.value, nil
}

func (p *InMemoryCacheProvider) Set(key string, value string, expiration time.Duration) error {
	var exp int64 = 0
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	p.data.Store(key, cacheItem{
		value:      value,
		expiration: exp,
	})
	return nil
}

func (p *InMemoryCacheProvider) Delete(key string) error {
	p.data.Delete(key)
	return nil
}

func (p *InMemoryCacheProvider) SetNX(key string, value string, expiration time.Duration) (bool, error) {
	// 先檢查是否過期
	if val, ok := p.data.Load(key); ok {
		item := val.(cacheItem)
		if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
			p.data.Delete(key)
		} else {
			return false, nil
		}
	}

	var exp int64 = 0
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	_, loaded := p.data.LoadOrStore(key, cacheItem{
		value:      value,
		expiration: exp,
	})

	return !loaded, nil
}

// NoopCacheProvider 是 CacheProvider 的空實作，不做任何事
type NoopCacheProvider struct{}

// NewNoopCacheProvider 建立一個新的 NoopCacheProvider
func NewNoopCacheProvider() *NoopCacheProvider {
	return &NoopCacheProvider{}
}

func (p *NoopCacheProvider) Get(key string) (string, error) {
	return "", nil
}

func (p *NoopCacheProvider) Set(key string, value string, expiration time.Duration) error {
	return nil
}

func (p *NoopCacheProvider) Delete(key string) error {
	return nil
}

func (p *NoopCacheProvider) SetNX(key string, value string, expiration time.Duration) (bool, error) {
	return true, nil
}
