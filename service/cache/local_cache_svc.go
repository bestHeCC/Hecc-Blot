package cache

import (
	"context"
	"sync"
	"time"

	"hecc-blot/contract/cache"
	iCoreTrace "hecc-blot/contract/trace"
	cacheConf "hecc-blot/entity/config/cache"
	"hecc-blot/util"
)

type localCacheSvc struct {
	// values is 缓存数据
	values map[string]*memCacheVal
	// locker is 读写锁
	locker sync.RWMutex
	// clearInterval is 清除缓存的时间间隔
	clearInterval time.Duration
	// traceSvc is 追踪服务
	traceSvc iCoreTrace.ITrace
}

type memCacheVal struct {
	//实际缓存数据
	content interface{}
	// 过期时间
	expireTime time.Time
	// 有效时长
	expire time.Duration
}

// Set 将value写入缓存
func (c *localCacheSvc) Set(ctx context.Context, key string, val interface{}, expire time.Duration) error {
	ctx = util.ExtractContext(ctx)

	if c.traceSvc != nil {
		_, span := c.traceSvc.Start(ctx, "localCache.SET",
			"cache.type", "local",
			"cache.operation", "SET",
			"cache.key", key,
		)
		defer span.End()
		c.set(key, val, expire)
		return nil
	}
	c.set(key, val, expire)
	return nil
}

// Get 根据key值获取value
func (c *localCacheSvc) Get(ctx context.Context, key string) (interface{}, error) {
	ctx = util.ExtractContext(ctx)

	if c.traceSvc != nil {
		_, span := c.traceSvc.Start(ctx, "localCache.GET",
			"cache.type", "local",
			"cache.operation", "GET",
			"cache.key", key,
		)
		defer span.End()
		result, ok := c.get(key)
		if !ok {
			return nil, nil
		}
		return result.content, nil
	}

	result, ok := c.get(key)
	if !ok {
		return nil, nil
	}
	return result.content, nil
}

// Del 删除key值
func (c *localCacheSvc) Del(ctx context.Context, key string) error {
	ctx = util.ExtractContext(ctx)

	if c.traceSvc != nil {
		_, span := c.traceSvc.Start(ctx, "localCache.DEL",
			"cache.type", "local",
			"cache.operation", "DEL",
			"cache.key", key,
		)
		defer span.End()
		c.delWithLock(key)
		return nil
	}
	c.delWithLock(key)
	return nil
}

// Exists 判断key是否存在
func (c *localCacheSvc) Exists(ctx context.Context, key string) (bool, error) {
	ctx = util.ExtractContext(ctx)

	if c.traceSvc != nil {
		_, span := c.traceSvc.Start(ctx, "localCache.EXISTS",
			"cache.type", "local",
			"cache.operation", "EXISTS",
			"cache.key", key,
		)
		defer span.End()
		_, ok := c.get(key)
		return ok, nil
	}

	_, ok := c.get(key)
	return ok, nil
}

// set is 内部方法，设置缓存
func (c *localCacheSvc) set(key string, val interface{}, expire time.Duration) {
	c.locker.Lock()
	defer c.locker.Unlock()

	// 直接删除而不调用 delWithLock（避免死锁）
	c.del(key)
	c.values[key] = &memCacheVal{
		content:    val,
		expireTime: time.Now().Add(expire),
		expire:     expire,
	}
}

// get is 内部方法，获取缓存
func (c *localCacheSvc) get(key string) (*memCacheVal, bool) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	v, ok := c.values[key]
	if ok {
		// 判断是否过期，直接删除而不调用 delWithLock（避免死锁）
		if v.expire != 0 && v.expireTime.Before(time.Now()) {
			c.del(key)
			return nil, false
		}
		return v, true
	}

	return nil, false
}

// delWithLock is 加锁形式删除缓存
func (c *localCacheSvc) delWithLock(key string) {
	c.locker.Lock()
	defer c.locker.Unlock()
	delete(c.values, key)
}

// del is 直接删除缓存
func (c *localCacheSvc) del(key string) {
	delete(c.values, key)
}

// clearExpired is 定期清除过期缓存
func (c *localCacheSvc) clearExpired() {
	// 检查间隔是否有效，无效则不启动清理
	if c.clearInterval <= 0 {
		return
	}

	// 声明定时器
	timeTicker := time.NewTicker(c.clearInterval)
	defer timeTicker.Stop()

	// for死循环，保存协程不会退出
	for {
		select {
		// 接收到定时器发来的消息
		case <-timeTicker.C:
			c.locker.Lock()
			for k, item := range c.values {
				if item.expire != 0 && item.expireTime.Before(time.Now()) {
					c.del(k)
				}
			}
			c.locker.Unlock()
		}
	}
}

func newLocalCacheSvc(config *cacheConf.Local, traceSvc iCoreTrace.ITrace) cache.ILocalCache {
	localCache := &localCacheSvc{
		values:        make(map[string]*memCacheVal),
		clearInterval: time.Duration(config.ClearInterval) * time.Second,
		traceSvc:      traceSvc,
	}

	// 开启Goroutine清除过期缓存
	go localCache.clearExpired()
	return localCache
}
