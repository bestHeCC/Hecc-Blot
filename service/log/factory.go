package log

import (
	"context"
	"fmt"

	ilog "core/contract/log"
	"core/entity/config"
	"core/enum/trace"
)

func getTraceId(ctx context.Context) string {
	traceId := ctx.Value(trace.TraceIdKey)
	if traceId != nil {
		return traceId.(string)
	}

	return ""
}

func NewLogger(config *config.Config) (ilog.ILog, error) {
	// 优先使用sls日志服务
	slsConfig := config.Log.Sls
	if slsConfig.Enable {
		return NewSlsSvc(&slsConfig)
	}

	// sls日志服务未开启，则使用本地日志服务
	localConfig := config.Log.Local
	if localConfig.Enable {
		return NewLogSvc(&localConfig)
	}

	return nil, fmt.Errorf("未配置日志服务")
}
