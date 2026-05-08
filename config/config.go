package config

import (
	"core/config/cache"
	"core/config/db"
	"core/config/log"
	"core/config/server"
)

type Config struct {
	Cache  cache.Config
	Db     db.Config
	Log    log.Config
	Server server.Config
}
