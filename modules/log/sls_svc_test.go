package log

import (
	"context"
	"testing"

	logConfig "github.com/bestHeCC/hecc-log/config"
	"github.com/bestHeCC/hecc-core/enum/trace"

	"github.com/stretchr/testify/assert"
)

var conf = &logConfig.SlsConfig{
	Enable:      true,
	Endpoint:    "cn-hangzhou.log.aliyuncs.com",
	AccessKey:   "",
	SecretKey:   "",
	SecretToken: "",
	Project:     "",
	LogStore:    "",
}

func TestSlsSvc(t *testing.T) {
	logger, err := newSlsSvc(conf)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	ctx := context.Background()
	ctxWithTraceId := context.WithValue(context.Background(), trace.TraceIdKey, "test-trace-id")

	cases := []struct {
		level string
		text  string
	}{
		{
			level: "debug",
			text:  "test debug",
		},
		{
			level: "info",
			text:  "test info",
		},
		{
			level: "warn",
			text:  "test warn",
		},
		{
			level: "error",
			text:  "test error",
		},
	}
	t.Run("log", func(t *testing.T) {
		for _, c := range cases {
			t.Run(c.level, func(t *testing.T) {
				switch c.level {
				case "debug":
					logger.Debug(ctx, c.text)
				case "info":
					logger.Info(ctx, c.text)
				case "warn":
					logger.Warn(ctx, c.text)
				case "error":
					logger.Error(ctx, c.text)
				}
			})
		}
	})

	t.Run("log with traceId", func(t *testing.T) {
		for _, c := range cases {
			t.Run(c.level, func(t *testing.T) {
				switch c.level {
				case "debug":
					logger.Debug(ctxWithTraceId, c.text)
				case "info":
					logger.Info(ctxWithTraceId, c.text)
				case "warn":
					logger.Warn(ctxWithTraceId, c.text)
				case "error":
					logger.Error(ctxWithTraceId, c.text)
				}
			})
		}
	})
}
