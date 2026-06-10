package server

import (
	envType "hecc-blot/enum/env"
)

type Config struct {
	Env           envType.Value `mapstructure:"env"`
	Port          string        `mapstructure:"port"`
	EnableTrace   bool          `mapstructure:"enable_trace"`
	ReadTimeout   int           `mapstructure:"read_timeout"`
	WriteTimeout  int           `mapstructure:"write_timeout"`
	IdleTimeout   int           `mapstructure:"idle_timeout"`
	BodySizeLimit int64         `mapstructure:"body_size_limit"`
}
