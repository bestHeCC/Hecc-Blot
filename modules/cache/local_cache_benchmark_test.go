package cache

import (
	"context"
	"testing"
	"time"

	cacheConfig "github.com/bestHeCC/hecc-cache/config"
)

// BenchmarkLocalCacheSet 度量本地缓存写入开销（写锁 + map 写入 + 过期时间计算）。
func BenchmarkLocalCacheSet(b *testing.B) {
	c := newLocalCacheSvc(&cacheConfig.Local{ClearInterval: 0}, nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = c.Set(ctx, "bench", "value", time.Minute)
	}
}

// BenchmarkLocalCacheGet 度量本地缓存读取开销（读锁 + map 查找 + 过期判断）。
func BenchmarkLocalCacheGet(b *testing.B) {
	c := newLocalCacheSvc(&cacheConfig.Local{ClearInterval: 0}, nil)
	ctx := context.Background()
	_ = c.Set(ctx, "bench", "value", time.Minute)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_, _ = c.Get(ctx, "bench")
	}
}
