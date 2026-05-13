package config

import (
	"hecc-blot/entity/config/cache"
	"hecc-blot/entity/config/db"
	"hecc-blot/entity/config/log"
	"hecc-blot/entity/config/server"
	"hecc-blot/entity/config/trace"
)

type Config struct {
	Cache  cache.Config
	Db     db.Config
	Log    log.Config
	Server server.Config
	Trace  trace.Config
}
