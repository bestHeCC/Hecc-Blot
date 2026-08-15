package cache

import (
	"github.com/bestHeCC/hecc-core/contract/cache"
	iCoreTrace "github.com/bestHeCC/hecc-core/contract/trace"
	cacheConf "github.com/bestHeCC/hecc-core/entity/config/cache"
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
