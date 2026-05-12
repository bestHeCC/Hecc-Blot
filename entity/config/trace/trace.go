package trace

type Config struct {
	ServiceName string        `mapstructure:"service_name"`
	Endpoint    string        `mapstructure:"endpoint"`
	Sampler     SamplerConfig `mapstructure:"sampler"`
	Trace       TraceConfig   `mapstructure:"trace"`
}

type SamplerConfig struct {
	Type  string  `mapstructure:"type"`
	Ratio float64 `mapstructure:"ratio"`
}

type TraceConfig struct {
	MysqlEnable bool `mapstructure:"mysql_enable"`
	RedisEnable bool `mapstructure:"redis_enable"`
}
