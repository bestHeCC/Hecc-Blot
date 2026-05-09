package config

import (
	"core/entity/config/cache"
	"core/entity/config/db"
	"core/entity/config/log"
	"core/entity/config/server"
)

type Config struct {
	Cache  cache.Config
	Db     db.Config
	Log    log.Config
	Server server.Config
}
