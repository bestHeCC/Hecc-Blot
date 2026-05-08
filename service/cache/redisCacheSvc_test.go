package cache

import (
	"testing"

	"core/config/cache"

	"github.com/stretchr/testify/assert"
)

var mockRedisConf = cache.Redis{
	Addr:     "127.0.0.1:6379",
	Password: "",
	PoolSize: 10,
	DB:       0,
}

func TestRedisSet(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-set"

	err := redisSvc.Set(key, "1", 0)
	assert.Equal(t, nil, err)
}

func TestRedisGet(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-get"
	val := "1"
	_ = redisSvc.Set(key, val, 0)

	v, err := redisSvc.Get(key)
	assert.Equal(t, nil, err)
	assert.Equalf(t, val, v, "取值不一致")
}

func TestRedisDel(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-del"
	val := "1"
	_ = redisSvc.Set(key, val, 0)

	err := redisSvc.Del(key)
	assert.Equal(t, nil, err)
}

func TestRedisExists(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-exists"
	val := "1"
	_ = redisSvc.Set(key, val, 0)

	exists, err := redisSvc.Exists(key)
	assert.Equal(t, nil, err)
	assert.Equalf(t, true, exists, "值不存在")
}

func TestRedisHSet(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-hset"
	field := "1"
	val := "1"
	err := redisSvc.HSet(key, field, val)
	assert.Equal(t, nil, err)
}

func TestRedisHGet(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-hget"
	field := "1"
	val := "1"
	err := redisSvc.HSet(key, field, val)
	assert.Equal(t, nil, err)

	v, err := redisSvc.HGet(key, field)
	assert.Equal(t, nil, err)
	assert.Equalf(t, val, v, "值不一致")
}

func TestRedisHDel(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	defer redisSvc.Close()

	key := "hcc-hget"
	field := "1"
	val := "1"
	err := redisSvc.HSet(key, field, val)
	assert.Equal(t, nil, err)

	err = redisSvc.HDel(key, field)
	assert.Equal(t, nil, err)
}

func TestRedisClose(t *testing.T) {
	redisSvc := newRedisCacheSvc(&mockRedisConf)
	err := redisSvc.Close()
	assert.Equal(t, nil, err)
}
