package main

import (
	cacheConfig "github.com/bestHeCC/hecc-cache/config"
	dbConfig "github.com/bestHeCC/hecc-db/config"
	logConfig "github.com/bestHeCC/hecc-log/config"
	serverConfig "github.com/bestHeCC/hecc-api/config"
	traceConfig "github.com/bestHeCC/hecc-trace/config"
)

// Config 业务方配置聚合，按模块组装各模块的配置。
type Config struct {
	Cache  cacheConfig.Config
	Db     dbConfig.Config
	Log    logConfig.Config
	Server serverConfig.Config
	Trace  traceConfig.Config
}
