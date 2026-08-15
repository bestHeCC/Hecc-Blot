package mocks

import (
	"context"
	"time"

	"github.com/bestHeCC/hecc-cache/contract"
)

// MockLocalCache 是 ILocalCache 接口的 mock 实现。
type MockLocalCache struct {
	GetFn    func(ctx context.Context, key string) (interface{}, error)
	SetFn    func(ctx context.Context, key string, val interface{}, expire time.Duration) error
	DelFn    func(ctx context.Context, key string) error
	ExistsFn func(ctx context.Context, key string) (bool, error)
}

func (m *MockLocalCache) Get(ctx context.Context, key string) (interface{}, error) {
	return m.GetFn(ctx, key)
}

func (m *MockLocalCache) Set(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	return m.SetFn(ctx, key, val, expire)
}

func (m *MockLocalCache) Del(ctx context.Context, key string) error {
	return m.DelFn(ctx, key)
}

func (m *MockLocalCache) Exists(ctx context.Context, key string) (bool, error) {
	return m.ExistsFn(ctx, key)
}

// MockRedisCache 是 IRedisCache 接口的 mock 实现。
type MockRedisCache struct {
	MockLocalCache
	HSetFn func(ctx context.Context, key string, values ...interface{}) error
	HGetFn func(ctx context.Context, key, field string) (string, error)
	HDelFn func(ctx context.Context, key string, fields ...string) error
	CloseF func() error
}

func (m *MockRedisCache) HSet(ctx context.Context, key string, values ...interface{}) error {
	return m.HSetFn(ctx, key, values...)
}

func (m *MockRedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	return m.HGetFn(ctx, key, field)
}

func (m *MockRedisCache) HDel(ctx context.Context, key string, fields ...string) error {
	return m.HDelFn(ctx, key, fields...)
}

func (m *MockRedisCache) Close() error {
	if m.CloseF != nil {
		return m.CloseF()
	}
	return nil
}

// MockCacheFactory 是 ICacheFactory 接口的 mock 实现。
type MockCacheFactory struct {
	LocalCache cache.ILocalCache
	RedisCache cache.IRedisCache
}

func (m *MockCacheFactory) Local() cache.ILocalCache { return m.LocalCache }

func (m *MockCacheFactory) Redis() cache.IRedisCache { return m.RedisCache }
