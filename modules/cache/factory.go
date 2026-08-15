package cache

import (
	cacheContract "github.com/bestHeCC/hecc-cache/contract"
	"github.com/bestHeCC/hecc-trace/contract"
	cacheConf "github.com/bestHeCC/hecc-cache/config"
)

type Factory struct {
	config *cacheConf.Config
	local  cacheContract.ILocalCache
	redis  cacheContract.IRedisCache
}

func (f Factory) Local() cacheContract.ILocalCache {
	return f.local
}

func (f Factory) Redis() cacheContract.IRedisCache {
	return f.redis
}

func NewCacheFactory(config *cacheConf.Config, traceSvc trace.ITrace) cacheContract.ICacheFactory {
	return Factory{
		config: config,
		local:  newLocalCacheSvc(&config.Local, traceSvc),
		redis:  newRedisCacheSvc(&config.Redis, traceSvc),
	}
}
