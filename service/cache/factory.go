package cache

import (
	"hecc-blot/contract/cache"
	iCoreTrace "hecc-blot/contract/trace"
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

func NewCacheFactory(config *cacheConf.Config, traceSvc iCoreTrace.ITrace) cache.ICacheFactory {
	return Factory{
		config: config,
		local:  newLocalCacheSvc(&config.Local, traceSvc),
		redis:  newRedisCacheSvc(&config.Redis, traceSvc),
	}
}
