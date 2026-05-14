package log

import (
	"context"
	"encoding/json"
	"time"

	ilog "hecc-blot/contract/log"
	"hecc-blot/entity/config/log"
	"hecc-blot/util"

	sls "github.com/aliyun/aliyun-log-go-sdk"
)

type SlsSvc struct {
	client   sls.ClientInterface
	project  string
	logStore string
}

func (s *SlsSvc) Debug(ctx context.Context, msg string, fields ...interface{}) {
	s.send(ctx, "debug", msg, fields...)
}

func (s *SlsSvc) Error(ctx context.Context, msg string, fields ...interface{}) {
	s.send(ctx, "error", msg, fields...)
}

func (s *SlsSvc) Info(ctx context.Context, msg string, fields ...interface{}) {
	s.send(ctx, "info", msg, fields...)
}

func (s *SlsSvc) Warn(ctx context.Context, msg string, fields ...interface{}) {
	s.send(ctx, "warn", msg, fields...)
}

func (s *SlsSvc) send(ctx context.Context, level string, msg string, fields ...interface{}) {
	traceId := getTraceId(util.ExtractContext(ctx))
	contents := []*sls.LogContent{
		newLogContent("level", level),
		newLogContent("message", msg),
		newLogContent("time", time.Now().Format("2006-01-02 15:04:05.000")),
	}

	if traceId != "" {
		contents = append(contents, newLogContent("traceId", traceId))
	}

	if len(fields) > 0 {
		if data, err := json.Marshal(fields); err == nil {
			contents = append(contents, newLogContent("fields", string(data)))
		}
	}

	logGroup := &sls.LogGroup{
		Logs: []*sls.Log{
			{
				Time:     new(uint32(time.Now().Unix())),
				Contents: contents,
			},
		},
	}

	_ = s.client.PutLogs(s.project, s.logStore, logGroup)
}

func newLogContent(key, value string) *sls.LogContent {
	return &sls.LogContent{
		Key:   &key,
		Value: &value,
	}
}

func newSlsSvc(conf *log.SlsConfig) (ilog.ILog, error) {
	provider := sls.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKey, conf.SecretToken)
	client := sls.CreateNormalInterfaceV2(conf.Endpoint, provider)

	return &SlsSvc{
		client:   client,
		project:  conf.Project,
		logStore: conf.LogStore,
	}, nil
}
