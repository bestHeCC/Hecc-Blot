package config

import (
	"github.com/bestHeCC/hecc-core/entity/config/cache"
	"github.com/bestHeCC/hecc-core/entity/config/db"
	"github.com/bestHeCC/hecc-core/entity/config/log"
	"github.com/bestHeCC/hecc-core/entity/config/server"
	"github.com/bestHeCC/hecc-core/entity/config/trace"
)

type Config struct {
	Cache  cache.Config
	Db     db.Config
	Log    log.Config
	Server server.Config
	Trace  trace.Config
}
