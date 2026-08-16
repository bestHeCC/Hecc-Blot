package cache

import (
	"context"
	"testing"

	"github.com/bestHeCC/hecc-core/contract/ratelimit"

	"github.com/stretchr/testify/assert"
)

// TestRedisLimiter 验证 Redis 滑动窗口限流（依赖真实 Redis，与其它 cache 测试一致）。
func TestRedisLimiter(t *testing.T) {
	ctx := context.Background()
	key := "hcc-ratelimit-test"

	// 复用框架内置 redis 实例，并用它的 Del 清理残留 key，保证测试确定性
	redisCache := newRedisCacheSvc(&mockRedisConf, nil)
	defer redisCache.Close()
	_ = redisCache.Del(ctx, "hecc:ratelimit:"+key)

	limiter := NewRedisLimiter(redisCache, ratelimit.Config{Algorithm: ratelimit.SlidingWindow, Limit: 2, Window: 60})

	assert.True(t, limiter.Allow(ctx, key))
	assert.True(t, limiter.Allow(ctx, key))
	assert.False(t, limiter.Allow(ctx, key))
}
