package cache

import (
	"hecc-blot/contract/cache"
	cacheConf "hecc-blot/entity/config/cache"
)

type Factory struct {
	config *cacheConf.Config
	local  cache.ILocalCache
	redis  cache.IRedisCache
}

func (f Factory) Local() cache.ILocalCache {
	return f.local
}

func (f Factory) Redis() cache.IRedisCache {
	return f.redis
}

func NewCacheFactory(config *cacheConf.Config) cache.ICacheFactory {
	return Factory{
		config: config,
		local:  newLocalCacheSvc(&config.Local),
		redis:  newRedisCacheSvc(&config.Redis),
	}
}
