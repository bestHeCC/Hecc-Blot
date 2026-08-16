package config

import (
	envType "github.com/bestHeCC/hecc-core/enum/env"
)

type Config struct {
	Env           envType.Value `mapstructure:"env"`
	Port          string        `mapstructure:"port"`
	ReadTimeout   int           `mapstructure:"read_timeout"`
	WriteTimeout  int           `mapstructure:"write_timeout"`
	IdleTimeout   int           `mapstructure:"idle_timeout"`
	BodySizeLimit int64         `mapstructure:"body_size_limit"`
	RateLimit     RateLimitConfig `mapstructure:"rate_limit"`
}

// RateLimitConfig 请求频率限流配置。是否启用由组装层决定（是否注册限流中间件），
// 本配置仅描述启用后的行为。
type RateLimitConfig struct {
	Backend   string `mapstructure:"backend"`   // 后端：memory(默认) | redis
	Algorithm string `mapstructure:"algorithm"` // 算法：sliding_window(默认) | token_bucket
	Limit     int    `mapstructure:"limit"`     // 窗口内最大请求数 / 令牌桶容量
	Window    int    `mapstructure:"window"`    // 窗口时长（秒）
}
