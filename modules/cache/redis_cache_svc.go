package cache

import (
	"context"
	"errors"
	"time"

	"github.com/bestHeCC/hecc-core/contract/cache"
	iCoreTrace "github.com/bestHeCC/hecc-core/contract/trace"
	cacheConf "github.com/bestHeCC/hecc-core/entity/config/cache"
	"github.com/bestHeCC/hecc-core/util"

	"github.com/redis/go-redis/v9"
)

// noopSpan 是 iCoreTrace.Span 的空实现，避免到处判空
type noopSpan struct{}

func (noopSpan) End()                             {}
func (noopSpan) SetAttribute(string, interface{}) {}
func (noopSpan) RecordError(error)                {}
func (noopSpan) Name() string                     { return "" }

type redisCacheSvc struct {
	client   *redis.Client
	traceSvc iCoreTrace.ITrace
}

// startSpan 统一创建 trace span，traceSvc 为空时返回 noopSpan
func (r *redisCacheSvc) startSpan(ctx context.Context, name string, extra ...interface{}) (context.Context, iCoreTrace.Span) {
	if r.traceSvc == nil {
		return ctx, noopSpan{}
	}
	attrs := make([]interface{}, 0, len(extra))
	attrs = append(attrs, extra...)
	return r.traceSvc.Start(ctx, name, attrs...)
}

func (r *redisCacheSvc) Set(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.SET",
		"db.system", "redis", "db.operation", "SET", "db.redis.key", key)
	defer span.End()
	err := r.client.Set(ctx, key, val, expire).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (r *redisCacheSvc) Get(ctx context.Context, key string) (interface{}, error) {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.GET",
		"db.system", "redis", "db.operation", "GET", "db.redis.key", key)
	defer span.End()
	result, err := r.client.Get(ctx, key).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		span.RecordError(err)
	}
	return result, err
}

func (r *redisCacheSvc) Del(ctx context.Context, key string) error {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.DEL",
		"db.system", "redis", "db.operation", "DEL", "db.redis.key", key)
	defer span.End()
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (r *redisCacheSvc) Exists(ctx context.Context, key string) (bool, error) {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.EXISTS",
		"db.system", "redis", "db.operation", "EXISTS", "db.redis.key", key)
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

func (r *redisCacheSvc) HSet(ctx context.Context, key string, values ...interface{}) error {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.HSET",
		"db.system", "redis", "db.operation", "HSET", "db.redis.key", key)
	defer span.End()
	err := r.client.HSet(ctx, key, values...).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
}

func (r *redisCacheSvc) HGet(ctx context.Context, key, field string) (string, error) {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.HGET",
		"db.system", "redis", "db.operation", "HGET", "db.redis.key", key, "db.redis.field", field)
	defer span.End()
	result, err := r.client.HGet(ctx, key, field).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		span.RecordError(err)
	}
	return result, err
}

func (r *redisCacheSvc) HDel(ctx context.Context, key string, fields ...string) error {
	ctx = util.ExtractContext(ctx)
	ctx, span := r.startSpan(ctx, "redis.HDEL",
		"db.system", "redis", "db.operation", "HDEL", "db.redis.key", key)
	defer span.End()
	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		span.RecordError(err)
	}
	return err
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
