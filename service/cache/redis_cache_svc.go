package cache

import (
	"context"
	"errors"
	"time"

	"hecc-blot/contract/cache"
	iCoreTrace "hecc-blot/contract/trace"
	cacheConf "hecc-blot/entity/config/cache"

	"github.com/redis/go-redis/v9"
)

type redisCacheSvc struct {
	client   *redis.Client
	traceSvc iCoreTrace.ITrace
}

func (r *redisCacheSvc) Set(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.SET",
			"db.system", "redis",
			"db.operation", "SET",
			"db.redis.key", key,
		)
		defer span.End()
		err := r.client.Set(ctx, key, val, expire).Err()
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
	return r.client.Set(ctx, key, val, expire).Err()
}

func (r *redisCacheSvc) Get(ctx context.Context, key string) (interface{}, error) {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.GET",
			"db.system", "redis",
			"db.operation", "GET",
			"db.redis.key", key,
		)
		defer span.End()
		result, err := r.client.Get(ctx, key).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			span.RecordError(err)
		}
		return result, err
	}
	return r.client.Get(ctx, key).Result()
}

func (r *redisCacheSvc) Del(ctx context.Context, key string) error {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.DEL",
			"db.system", "redis",
			"db.operation", "DEL",
			"db.redis.key", key,
		)
		defer span.End()
		err := r.client.Del(ctx, key).Err()
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheSvc) Exists(ctx context.Context, key string) (bool, error) {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.EXISTS",
			"db.system", "redis",
			"db.operation", "EXISTS",
			"db.redis.key", key,
		)
		defer span.End()
		n, err := r.client.Exists(ctx, key).Result()
		if err != nil {
			span.RecordError(err)
			return false, err
		}
		if n == 0 {
			return false, errors.New("key no exists")
		}
		return true, nil
	}
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return false, errors.New("key no exists")
	}
	return true, nil
}

func (r *redisCacheSvc) HSet(ctx context.Context, key string, values ...interface{}) error {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.HSET",
			"db.system", "redis",
			"db.operation", "HSET",
			"db.redis.key", key,
		)
		defer span.End()
		err := r.client.HSet(ctx, key, values...).Err()
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
	return r.client.HSet(ctx, key, values...).Err()
}

func (r *redisCacheSvc) HGet(ctx context.Context, key, field string) (string, error) {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.HGET",
			"db.system", "redis",
			"db.operation", "HGET",
			"db.redis.key", key,
			"db.redis.field", field,
		)
		defer span.End()
		result, err := r.client.HGet(ctx, key, field).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			span.RecordError(err)
		}
		return result, err
	}
	return r.client.HGet(ctx, key, field).Result()
}

func (r *redisCacheSvc) HDel(ctx context.Context, key string, fields ...string) error {
	if r.traceSvc != nil {
		ctx, span := r.traceSvc.Start(ctx, "redis.HDEL",
			"db.system", "redis",
			"db.operation", "HDEL",
			"db.redis.key", key,
		)
		defer span.End()
		err := r.client.HDel(ctx, key, fields...).Err()
		if err != nil {
			span.RecordError(err)
		}
		return err
	}
	return r.client.HDel(ctx, key, fields...).Err()
}

func (r *redisCacheSvc) Close() error {
	return r.client.Close()
}

func newRedisCacheSvc(redisConf *cacheConf.Redis, traceSvc iCoreTrace.ITrace) cache.IRedisCache {
	c := redis.NewClient(&redis.Options{
		Addr:     redisConf.Addr,
		Password: redisConf.Password,
		DB:       redisConf.DB,
		PoolSize: redisConf.PoolSize,
	})

	return &redisCacheSvc{
		client:   c,
		traceSvc: traceSvc,
	}
}
