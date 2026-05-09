package log

import (
	"context"
	"os"
	"time"

	"core/entity/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogSvc struct {
	logger *zap.Logger
}

func (b LogSvc) Debug(ctx context.Context, msg string, fields ...interface{}) {
	traceId := b.getTraceId(ctx)
	b.logger.Debug(msg, b.buildFields(traceId, fields...)...)
}

func (b LogSvc) Error(ctx context.Context, msg string, fields ...interface{}) {
	traceId := b.getTraceId(ctx)
	b.logger.Error(msg, b.buildFields(traceId, fields...)...)
}

func (b LogSvc) Info(ctx context.Context, msg string, fields ...interface{}) {
	traceId := b.getTraceId(ctx)
	b.logger.Info(msg, b.buildFields(traceId, fields...)...)
}

func (b LogSvc) Panic(ctx context.Context, msg string, fields ...interface{}) {
	traceId := b.getTraceId(ctx)
	b.logger.Panic(msg, b.buildFields(traceId, fields...)...)
}

func (b LogSvc) Warn(ctx context.Context, msg string, fields ...interface{}) {
	traceId := b.getTraceId(ctx)
	b.logger.Warn(msg, b.buildFields(traceId, fields...)...)
}

func (b LogSvc) getTraceId(ctx context.Context) string {
	return ""
}

func (b LogSvc) buildFields(traceId string, fields ...interface{}) []zapcore.Field {
	var zapFields []zapcore.Field

	if traceId != "" {
		zapFields = append(zapFields, zap.String("traceId", traceId))
	}

	for _, v := range fields {
		if field, ok := v.(zapcore.Field); ok {
			zapFields = append(zapFields, field)
		}
	}

	return zapFields
}

func NewLogger(logConf *config.Config, levels map[zapcore.Level]string) (*zap.Logger, error) {
	var rootDir string
	var maxSize, maxBackups, maxAge int
	var compress bool

	rootDir = logConf.Log.RootDir
	maxSize = logConf.Log.MaxSize
	maxBackups = logConf.Log.MaxBackups
	maxAge = logConf.Log.MaxAge
	compress = logConf.Log.Compress

	// 判断目录是否存在
	_, err := os.Stat(rootDir)
	if os.IsNotExist(err) {
		// 创建根目录
		if err = os.MkdirAll(rootDir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// 扩展Zap
	var encoder zapcore.Encoder

	// 调整编码器默认配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(time.Format("[" + "2006-01-02 15:04:05.000" + "]"))
	}
	encoderConfig.EncodeLevel = func(l zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(string(logConf.Server.Env) + "." + l.String())
	}

	// 设置编码器
	encoder = zapcore.NewJSONEncoder(encoderConfig)

	var cores []zapcore.Core
	for level, suffix := range levels {
		cores = append(cores, zapcore.NewCore(
			encoder,
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   rootDir + "/" + logConf.Server.Name + "_" + suffix + ".log",
				MaxSize:    maxSize,
				MaxBackups: maxBackups,
				MaxAge:     maxAge,
				Compress:   compress,
			}),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl == level
			}),
		))
	}

	return zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)), nil
}
