package cache

import (
	"context"
	"testing"

	"github.com/bestHeCC/hecc-cache/config"
	traceConf "github.com/bestHeCC/hecc-trace/config"
	"github.com/bestHeCC/hecc-trace"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var mockRedisConf = cache.Redis{
	Addr:     "127.0.0.1:6379",
	Password: "",
	PoolSize: 10,
	DB:       0,
}

var mockTraceConf = traceConf.Config{
	ServiceName: "Hecc-Blot",
	Endpoint:    "127.0.0.1:4318",
	Sampler: traceConf.SamplerConfig{
		Type:  "always",
		Ratio: 0.5,
	},
}

func TestRedisSvc(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf, nil)
	defer redisSvc.Close()

	ctx := context.Background()
	t.Run("set", func(t *testing.T) {
		key := "hcc-set"

		err := redisSvc.Set(ctx, key, "1", 0)
		assert.Equal(t, nil, err)
	})

	t.Run("get", func(t *testing.T) {
		key := "hcc-get"
		val := "1"
		err := redisSvc.Set(ctx, key, val, 0)
		assert.Equal(t, nil, err)

		v, err := redisSvc.Get(ctx, key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "取值不一致")
	})

	t.Run("delWithLock", func(t *testing.T) {
		key := "hcc-delWithLock"
		val := "1"
		err := redisSvc.Set(ctx, key, val, 0)
		assert.Equal(t, nil, err)

		err = redisSvc.Del(ctx, key)
		assert.Equal(t, nil, err)

		_, err = redisSvc.Get(ctx, key)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("exists", func(t *testing.T) {
		key := "hcc-exists"
		val := "1"
		err := redisSvc.Set(ctx, key, val, 0)
		assert.Equal(t, nil, err)

		exists, err := redisSvc.Exists(ctx, key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, true, exists, "key不存在")
	})

	t.Run("hset", func(t *testing.T) {
		key := "hcc-hset"
		field := "1"
		val := "1"
		err := redisSvc.HSet(ctx, key, field, val)
		assert.Equal(t, nil, err)
	})

	t.Run("hget", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redisSvc.HSet(ctx, key, field, val)
		assert.Equal(t, nil, err)

		v, err := redisSvc.HGet(ctx, key, field)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "值不一致")
	})

	t.Run("hdel", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redisSvc.HSet(ctx, key, field, val)
		assert.Equal(t, nil, err)

		err = redisSvc.HDel(ctx, key, field)
		assert.Equal(t, nil, err)
	})
}

func TestRedisSvcWithTrace(t *testing.T) {
	traceSvc, traceClearUp, err := trace.NewTraceSvc(&mockTraceConf)
	assert.Equal(t, nil, err)

	redisSvc := newRedisCacheSvc(&mockRedisConf, traceSvc)
	defer func() {
		traceClearUp()
		redisSvc.Close()
	}()

	ctx := context.Background()
	t.Run("set", func(t *testing.T) {
		key := "hcc-set"

		err := redisSvc.Set(ctx, key, "1", 0)
		assert.Equal(t, nil, err)
	})

	t.Run("get", func(t *testing.T) {
		key := "hcc-get"
		val := "1"
		err := redisSvc.Set(ctx, key, val, 0)
		assert.Equal(t, nil, err)

		v, err := redisSvc.Get(ctx, key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "取值不一致")
	})

	t.Run("delWithLock", func(t *testing.T) {
		key := "hcc-delWithLock"
		val := "1"
		err := redisSvc.Set(ctx, key, val, 0)
		assert.Equal(t, nil, err)

		err = redisSvc.Del(ctx, key)
		assert.Equal(t, nil, err)

		_, err = redisSvc.Get(ctx, key)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("exists", func(t *testing.T) {
		key := "hcc-exists"
		val := "1"
		err := redisSvc.Set(ctx, key, val, 0)
		assert.Equal(t, nil, err)

		exists, err := redisSvc.Exists(ctx, key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, true, exists, "key不存在")
	})

	t.Run("hset", func(t *testing.T) {
		key := "hcc-hset"
		field := "1"
		val := "1"
		err := redisSvc.HSet(ctx, key, field, val)
		assert.Equal(t, nil, err)
	})

	t.Run("hget", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redisSvc.HSet(ctx, key, field, val)
		assert.Equal(t, nil, err)

		v, err := redisSvc.HGet(ctx, key, field)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "值不一致")
	})

	t.Run("hdel", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redisSvc.HSet(ctx, key, field, val)
		assert.Equal(t, nil, err)

		err = redisSvc.HDel(ctx, key, field)
		assert.Equal(t, nil, err)
	})
}
