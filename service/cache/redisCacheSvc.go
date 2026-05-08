package cache

import (
	"context"
	"errors"
	"time"

	cacheConf "core/config/cache"
	"core/contract/cache"

	"github.com/redis/go-redis/v9"
)

type redisCacheSvc struct {
	client *redis.Client
}

var ctx = context.Background()

func (r *redisCacheSvc) Set(key string, val interface{}, expire time.Duration) error {
	return r.client.Set(ctx, key, val, expire).Err()
}

func (r *redisCacheSvc) Get(key string) (interface{}, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *redisCacheSvc) Del(key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheSvc) Exists(key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	if n == 0 {
		return false, errors.New("key no exists")
	}

	return true, nil
}

func (r *redisCacheSvc) HSet(key string, values ...interface{}) error {
	return r.client.HSet(ctx, key, values...).Err()
}

func (r *redisCacheSvc) HGet(key, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

func (r *redisCacheSvc) HDel(key string, fields ...string) error {
	return r.client.HDel(ctx, key, fields...).Err()
}

func (r *redisCacheSvc) Close() error {
	return r.client.Close()
}

func newRedisCacheSvc(redisConf *cacheConf.Redis) cache.IRedisCache {
	c := redis.NewClient(&redis.Options{
		Addr:     redisConf.Addr,
		Password: redisConf.Password,
		DB:       redisConf.DB,
		PoolSize: redisConf.PoolSize,
	})

	return &redisCacheSvc{
		client: c,
	}
}
