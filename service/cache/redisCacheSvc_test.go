package cache

import (
	"testing"

	"core/entity/config/cache"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

var mockRedisConf = cache.Redis{
	Addr:     "127.0.0.1:6379",
	Password: "",
	PoolSize: 10,
	DB:       0,
}

func TestRedisSvc(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	t.Run("set", func(t *testing.T) {
		key := "hcc-set"

		err := redisSvc.Set(key, "1", 0)
		assert.Equal(t, nil, err)
	})

	t.Run("get", func(t *testing.T) {
		key := "hcc-get"
		val := "1"
		err := redisSvc.Set(key, val, 0)
		assert.Equal(t, nil, err)

		v, err := redisSvc.Get(key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "取值不一致")
	})

	t.Run("del", func(t *testing.T) {
		key := "hcc-del"
		val := "1"
		err := redisSvc.Set(key, val, 0)
		assert.Equal(t, nil, err)

		err = redisSvc.Del(key)
		assert.Equal(t, nil, err)

		_, err = redisSvc.Get(key)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("exists", func(t *testing.T) {
		key := "hcc-exists"
		val := "1"
		err := redisSvc.Set(key, val, 0)
		assert.Equal(t, nil, err)

		exists, err := redisSvc.Exists(key)
		assert.Equal(t, nil, err)
		assert.Equalf(t, true, exists, "key不存在")
	})

	t.Run("hset", func(t *testing.T) {
		key := "hcc-hset"
		field := "1"
		val := "1"
		err := redisSvc.HSet(key, field, val)
		assert.Equal(t, nil, err)
	})

	t.Run("hget", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redisSvc.HSet(key, field, val)
		assert.Equal(t, nil, err)

		v, err := redisSvc.HGet(key, field)
		assert.Equal(t, nil, err)
		assert.Equalf(t, val, v, "值不一致")
	})

	t.Run("hdel", func(t *testing.T) {
		key := "hcc-hget"
		field := "1"
		val := "1"
		err := redisSvc.HSet(key, field, val)
		assert.Equal(t, nil, err)

		err = redisSvc.HDel(key, field)
		assert.Equal(t, nil, err)
	})
}
