package server

import (
	envType "core/enum/env"
)

type Config struct {
	Env         envType.Value `mapstructure:"env"`
	Name        string        `mapstructure:"name"`
	Port        string        `mapstructure:"port"`
	EnableTrace bool          `mapstructure:"enable_trace"`
}
