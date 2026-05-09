package log

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"core/entity/config"
	"core/entity/config/log"
	"core/entity/config/server"
	envEnum "core/enum/env"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newTestLogSvc() *LogSvc {
	return &LogSvc{
		logger: zap.NewNop(),
	}
}

func TestLogSvc_Debug(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Debug(ctx, "debug message")
}

func TestLogSvc_Info(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Info(ctx, "info message")
}

func TestLogSvc_Warn(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Warn(ctx, "warn message")
}

func TestLogSvc_Error(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Error(ctx, "error message")
}

func TestLogSvc_Panic(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	assert.Panics(t, func() {
		svc.Panic(ctx, "panic message")
	})
}

func TestLogSvc_Debug_WithFields(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Debug(ctx, "debug with fields", zap.String("key", "value"))
}

func TestLogSvc_Info_WithFields(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Info(ctx, "info with fields", zap.Int("count", 123))
}

func TestLogSvc_Error_WithFields(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	svc.Error(ctx, "error with fields", zap.Any("data", map[string]string{"a": "b"}))
}

func TestLogSvc_getTraceId(t *testing.T) {
	svc := newTestLogSvc()
	ctx := context.Background()

	traceId := svc.getTraceId(ctx)
	assert.Equal(t, "", traceId)
}

func TestLogSvc_buildFields_Empty(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("")
	assert.Empty(t, fields)
}

func TestLogSvc_buildFields_WithTraceId(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("trace-123")
	assert.Len(t, fields, 1)
	assert.Equal(t, "traceId", fields[0].Key)
	assert.Equal(t, "trace-123", fields[0].String)
}

func TestLogSvc_buildFields_WithZapFields(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("", zap.String("custom", "value"))
	assert.Len(t, fields, 1)
	assert.Equal(t, "custom", fields[0].Key)
}

func TestLogSvc_buildFields_WithTraceIdAndZapFields(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("trace-456", zap.Int("count", 10))
	assert.Len(t, fields, 2)
	assert.Equal(t, "traceId", fields[0].Key)
	assert.Equal(t, "count", fields[1].Key)
}

func TestLogSvc_buildFields_IgnoreNonZapFields(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("", "not a zap field", 123, struct{}{})
	assert.Len(t, fields, 0)
}

func TestLogSvc_buildFields_MixedFields(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("trace-789", zap.String("key1", "val1"), "ignore", zap.Int("key2", 42))
	assert.Len(t, fields, 3)
	assert.Equal(t, "traceId", fields[0].Key)
	assert.Equal(t, "key1", fields[1].Key)
	assert.Equal(t, "key2", fields[2].Key)
}

func TestNewLogger(t *testing.T) {
	tmpDir := t.TempDir()

	conf := &config.Config{
		Server: server.Config{
			Env:  envEnum.DevMode,
			Name: "test-server",
			Port: "8080",
		},
		Log: log.Config{
			RootDir:    tmpDir,
			MaxSize:    1,
			MaxBackups: 3,
			MaxAge:     7,
			Compress:   false,
		},
	}

	logger, err := NewLogger(conf)
	assert.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestNewLogger_CreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "logs", "nested")

	conf := &config.Config{
		Server: server.Config{
			Env:  envEnum.TestMode,
			Name: "test-app",
			Port: "9090",
		},
		Log: log.Config{
			RootDir:    newDir,
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	_, err := NewLogger(conf)
	assert.NoError(t, err)

	info, err := os.Stat(newDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestNewLogger_CreatesLogFiles(t *testing.T) {
	tmpDir := t.TempDir()

	conf := &config.Config{
		Server: server.Config{
			Env:  envEnum.ProductMode,
			Name: "prod-app",
			Port: "80",
		},
		Log: log.Config{
			RootDir:    tmpDir,
			MaxSize:    1,
			MaxBackups: 2,
			MaxAge:     30,
			Compress:   true,
		},
	}

	logger, err := NewLogger(conf)
	assert.NoError(t, err)
	assert.NotNil(t, logger)

	levels := []string{"debug", "info", "warn", "error", "panic"}
	for _, level := range levels {
		filePath := filepath.Join(tmpDir, "prod-app_"+level+".log")
		_, err := os.Stat(filePath)
		assert.NoError(t, err, "log file should exist: %s", filePath)
	}
}

func TestNewLogger_WritesToLog(t *testing.T) {
	tmpDir := t.TempDir()

	conf := &config.Config{
		Server: server.Config{
			Env:  envEnum.DevMode,
			Name: "write-test",
			Port: "8081",
		},
		Log: log.Config{
			RootDir:    tmpDir,
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	logger, err := NewLogger(conf)
	assert.NoError(t, err)

	ctx := context.Background()
	logger.Info(ctx, "test message")

	logFile := filepath.Join(tmpDir, "write-test_info.log")

	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "test message")
}

func TestNewLogger_InvalidPath(t *testing.T) {
	conf := &config.Config{
		Server: server.Config{
			Env:  envEnum.DevMode,
			Name: "test",
			Port: "8080",
		},
		Log: log.Config{
			RootDir:    "/invalid/path/that/does/not/exist/and/cannot/be/created",
			MaxSize:    1,
			MaxBackups: 1,
			MaxAge:     1,
			Compress:   false,
		},
	}

	_, err := NewLogger(conf)
	assert.Error(t, err)
}

func TestLogSvc_Interface(t *testing.T) {
	svc := newTestLogSvc()
	var _ interface {
		Debug(ctx context.Context, msg string, fields ...interface{})
		Error(ctx context.Context, msg string, fields ...interface{})
		Info(ctx context.Context, msg string, fields ...interface{})
		Panic(ctx context.Context, msg string, fields ...interface{})
		Warn(ctx context.Context, msg string, fields ...interface{})
	} = svc
}

func TestLogSvc_buildFields_AllZapTypes(t *testing.T) {
	svc := newTestLogSvc()

	fields := svc.buildFields("trace-id",
		zap.String("s", "str"),
		zap.Int("i", 42),
		zap.Int64("i64", 64),
		zap.Float64("f64", 3.14),
		zap.Bool("b", true),
		zap.Duration("d", 0),
		zap.Time("t", time.Now()),
	)

	assert.Len(t, fields, 7)
}
