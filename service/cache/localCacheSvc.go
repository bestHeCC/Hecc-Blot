package cache

import (
	"sync"
	"time"

	cacheConf "core/config/cache"
	"core/contract/cache"
)

type localCacheSvc struct {
	// values is 缓存数据
	values map[string]*memCacheVal
	// locker is 读写锁
	locker sync.RWMutex
	// clearInterval is 清除缓存的时间间隔
	clearInterval time.Duration
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
func (c *localCacheSvc) Set(key string, val interface{}, expire time.Duration) error {
	// 上锁，保证线程安全
	c.locker.Lock()
	defer c.locker.Unlock()

	c.del(key)
	c.add(key, &memCacheVal{
		content:    val,
		expireTime: time.Now().Add(expire),
		expire:     expire,
	})

	return nil
}

// Get 根据key值获取value
func (c *localCacheSvc) Get(key string) (interface{}, error) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	v, ok := c.get(key)
	if ok {
		// 判断是否过期
		if v.expire != 0 && v.expireTime.Before(time.Now()) {
			c.del(key)
			return nil, nil
		}
		return v.content, nil
	}

	return nil, nil
}

// Del 删除key值
func (c *localCacheSvc) Del(key string) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.del(key)

	return nil
}

// Exists 判断key是否存在
func (c *localCacheSvc) Exists(key string) (bool, error) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	_, ok := c.get(key)

	return ok, nil
}

// get is 内部方法，获取缓存
func (c *localCacheSvc) get(key string) (*memCacheVal, bool) {
	val, ok := c.values[key]
	return val, ok
}

// add is 内部方法，添加缓存
func (c *localCacheSvc) add(key string, val *memCacheVal) {
	c.values[key] = val
}

// del is 内部方法，删除缓存
func (c *localCacheSvc) del(key string) {
	_, ok := c.get(key)
	if ok {
		delete(c.values, key)
	}
}

// clearExpired is 定期清除过期缓存
func (c *localCacheSvc) clearExpired() {
	// 声明定时器
	timeTicker := time.NewTicker(c.clearInterval)
	defer timeTicker.Stop()

	// for死循环，保存协程不会退出
	for {
		select {
		// 接收到定时器发来的消息
		case <-timeTicker.C:
			for k, item := range c.values {
				if item.expire != 0 && item.expireTime.Before(time.Now()) {
					c.locker.Lock()
					c.del(k)
					c.locker.Unlock()
				}
			}
		}
	}
}

func newLocalCacheSvc(config *cacheConf.Local) cache.ILocalCache {
	localCache := &localCacheSvc{
		values:        make(map[string]*memCacheVal),
		clearInterval: time.Duration(config.ClearInterval) * time.Second,
	}

	// 开启Goroutine清除过期缓存
	go localCache.clearExpired()
	return localCache
}
