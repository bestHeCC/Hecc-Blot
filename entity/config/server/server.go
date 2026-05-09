package server

import (
	envType "core/enum/env"
)

type Config struct {
	Env  envType.Value
	Name string
	Port string
}
